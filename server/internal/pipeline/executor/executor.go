package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	directoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	hooksagent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/hooks"
	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/research"
	researchagent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/research"
	scriptagent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	supervisoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/supervisor"
	creativerunrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/repository"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/tts"
	imagemedia "github.com/woragis/ecom-op-creatives-backend/server/internal/media/image"
	postprocessmedia "github.com/woragis/ecom-op-creatives-backend/server/internal/media/postprocess"
	rendermedia "github.com/woragis/ecom-op-creatives-backend/server/internal/media/render"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/subtitles"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/video"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	pipelinesvc "github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline/service"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/applog"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/runctx"
	productrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/product/repository"
	"time"
)

type Executor struct {
	cfg       config.Config
	repo      *creativerunrepo.Repository
	products  *productrepo.Repository
	pipeline  *pipelinesvc.Service
	storage   *storage.Local
	tts       tts.Synthesizer
	research  *researchagent.Agent
	hooks     *hooksagent.Agent
	script    *scriptagent.Agent
	director  *directoragent.Agent
	prompter  *prompteragent.Agent
	supervisor *supervisoragent.Agent
	image      *imagemedia.Service
	video      *video.Service
	subtitles  *subtitles.Service
	postprocess *postprocessmedia.Processor
}

type Deps struct {
	Cfg        config.Config
	Repo       *creativerunrepo.Repository
	Products   *productrepo.Repository
	Pipeline   *pipelinesvc.Service
	Storage    *storage.Local
	TTS        tts.Synthesizer
	Research   *researchagent.Agent
	Hooks      *hooksagent.Agent
	Script     *scriptagent.Agent
	Director   *directoragent.Agent
	Prompter   *prompteragent.Agent
	Supervisor *supervisoragent.Agent
	Image       *imagemedia.Service
	Video       *video.Service
	Subtitles   *subtitles.Service
	Postprocess *postprocessmedia.Processor
}

func New(d Deps) *Executor {
	return &Executor{
		cfg: d.Cfg, repo: d.Repo, products: d.Products, pipeline: d.Pipeline,
		storage: d.Storage, tts: d.TTS,
		research: d.Research, hooks: d.Hooks, script: d.Script,
		director: d.Director, prompter: d.Prompter, supervisor: d.Supervisor,
		image: d.Image, video: d.Video,
		subtitles: d.Subtitles, postprocess: d.Postprocess,
	}
}

type runContext struct {
	run     *models.CreativeRun
	product *models.Product
	outputs map[string]json.RawMessage
	assets  *models.RunAssets
}

func (e *Executor) ProcessStep(ctx context.Context, stepID uuid.UUID) error {
	step, err := e.repo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	log := applog.ForStep(step.CreativeRunID, step.StepType)
	if step.Status == models.StepStatusDone {
		return e.advance(ctx, step)
	}
	if step.Status == models.StepStatusFailed {
		log.Warn("ignoring stale queue message for failed step")
		return nil
	}

	if err := e.repo.MarkStepRunning(ctx, stepID); err != nil {
		return err
	}

	started := time.Now()
	log.Info("step started")

	rc, err := e.loadContext(ctx, step.CreativeRunID)
	if err != nil {
		return e.failStep(ctx, step, log, started, err)
	}

	stepCtx := runctx.With(ctx, step.CreativeRunID.String(), stepID.String(), step.StepType)
	output, err := e.executeStep(stepCtx, step.StepType, rc)
	if err != nil {
		return e.failStep(ctx, step, log, started, err)
	}

	provider := extractProvider(output)
	if err := e.repo.CompleteStep(ctx, stepID, output, provider); err != nil {
		return err
	}
	_ = e.storage.WriteStepArtifact(step.CreativeRunID.String(), step.StepType, output)
	attrs := []any{"duration_ms", time.Since(started).Milliseconds()}
	if provider != nil {
		attrs = append(attrs, "provider", *provider)
	}
	log.Info("step completed", attrs...)
	return e.advance(ctx, step)
}

func (e *Executor) failStep(ctx context.Context, step *models.PipelineStep, log *slog.Logger, started time.Time, err error) error {
	msg := err.Error()
	_ = e.storage.WriteStepErrorArtifact(step.CreativeRunID.String(), step.StepType, msg)
	_ = e.repo.FailStep(ctx, step.ID, msg)
	_ = e.repo.UpdateRunStatus(ctx, step.CreativeRunID, models.RunStatusFailed)
	log.Error("step failed", "error", msg, "duration_ms", time.Since(started).Milliseconds())
	return nil
}

func (e *Executor) advance(ctx context.Context, step *models.PipelineStep) error {
	next, err := e.repo.NextPendingStep(ctx, step.CreativeRunID, step.StepOrder)
	if err != nil {
		return err
	}
	if next == nil {
		run, err := e.repo.GetByID(ctx, step.CreativeRunID)
		if err != nil {
			return err
		}
		if run.Status == models.RunStatusRunning {
			applog.ForRun(step.CreativeRunID).Info("pipeline finished — awaiting review")
			return e.repo.UpdateRunStatus(ctx, step.CreativeRunID, models.RunStatusNeedsReview)
		}
		return nil
	}
	if e.cfg.PauseBeforeVideo && step.StepType == "image" && next.StepType == "video" {
		applog.ForRun(step.CreativeRunID).Info("paused before video — awaiting continue")
		return e.repo.UpdateRunStatus(ctx, step.CreativeRunID, models.RunStatusNeedsReview)
	}
	return e.pipeline.EnqueueStep(ctx, next)
}

func extractProvider(output []byte) *string {
	var m map[string]any
	if err := json.Unmarshal(output, &m); err != nil {
		return nil
	}
	p, ok := m["provider"].(string)
	if !ok || p == "" {
		return nil
	}
	return &p
}

func (e *Executor) loadContext(ctx context.Context, runID uuid.UUID) (*runContext, error) {
	run, err := e.repo.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	product, err := e.products.GetByID(ctx, run.ProductID)
	if err != nil {
		return nil, err
	}
	outputs := map[string]json.RawMessage{}
	for _, s := range run.Steps {
		if s.Status == models.StepStatusDone && len(s.OutputJSON) > 0 {
			outputs[s.StepType] = append([]byte(nil), s.OutputJSON...)
		}
	}
	return &runContext{
		run: run, product: product, outputs: outputs,
		assets: models.ParseRunAssets(run.InputAssets),
	}, nil
}

func (e *Executor) executeStep(ctx context.Context, stepType string, rc *runContext) ([]byte, error) {
	switch stepType {
	case "research":
		return e.stepResearch(ctx, rc)
	case "hooks":
		return e.stepHooks(ctx, rc)
	case "script":
		return e.stepScript(ctx, rc)
	case "director":
		return e.stepDirector(ctx, rc)
	case "prompter":
		return e.stepPrompter(ctx, rc)
	case "voice":
		return e.stepVoice(ctx, rc)
	case "image":
		return e.stepImage(ctx, rc)
	case "video":
		return e.stepVideo(ctx, rc)
	case "subtitles":
		return e.stepSubtitles(ctx, rc)
	case "render":
		return e.stepRender(ctx, rc)
	case "postprocess":
		return e.stepPostprocess(ctx, rc)
	case "supervisor":
		return e.stepSupervisor(ctx, rc)
	default:
		return nil, fmt.Errorf("unknown step type: %s", stepType)
	}
}

func (e *Executor) stepResearch(ctx context.Context, rc *runContext) ([]byte, error) {
	out, err := e.research.Execute(ctx, research.Input{
		ProductName: rc.product.Name,
		Description: rc.product.Description,
		ProductURL:  rc.product.URL,
		Niche:       rc.product.Niche,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (e *Executor) stepHooks(ctx context.Context, rc *runContext) ([]byte, error) {
	var res *research.Output
	if raw, ok := rc.outputs["research"]; ok {
		_ = json.Unmarshal(raw, &res)
	}
	out, err := e.hooks.Execute(ctx, hooksagent.Input{
		ProductName: rc.product.Name,
		Description: rc.product.Description,
		Research:    res,
		UserHook:    rc.run.Hook,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (e *Executor) stepScript(ctx context.Context, rc *runContext) ([]byte, error) {
	var res *research.Output
	var hook *hooksagent.Output
	_ = json.Unmarshal(rc.outputs["research"], &res)
	_ = json.Unmarshal(rc.outputs["hooks"], &hook)
	out, err := e.script.Execute(ctx, scriptwriter.Input{
		ProductName: rc.product.Name,
		Description: rc.product.Description,
		Research:    res,
		Hook:        hook,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (e *Executor) stepDirector(ctx context.Context, rc *runContext) ([]byte, error) {
	var script *scriptwriter.Output
	_ = json.Unmarshal(rc.outputs["script"], &script)
	out, err := e.director.Execute(ctx, script)
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (e *Executor) stepPrompter(ctx context.Context, rc *runContext) ([]byte, error) {
	var script *scriptwriter.Output
	var dir *directoragent.Output
	_ = json.Unmarshal(rc.outputs["script"], &script)
	_ = json.Unmarshal(rc.outputs["director"], &dir)
	out, err := e.prompter.Execute(ctx, prompteragent.Input{
		Script:      script,
		Director:    dir,
		Product:     rc.product.Name,
		Description: rc.product.Description,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (e *Executor) stepVoice(ctx context.Context, rc *runContext) ([]byte, error) {
	var script *scriptwriter.Output
	_ = json.Unmarshal(rc.outputs["script"], &script)
	text := scriptagent.FullNarration(script)
	audio, provider, err := e.tts.Synthesize(ctx, text)
	if err != nil {
		return nil, err
	}
	path, err := e.storage.WriteFile(rc.run.ID.String(), "narration.mp3", audio)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"text":        text,
		"filePath":    path,
		"publicUrl":   e.storage.PublicPath(rc.run.ID.String(), "narration.mp3"),
		"provider":    provider,
	})
}

func (e *Executor) stepImage(ctx context.Context, rc *runContext) ([]byte, error) {
	if e.image == nil {
		return json.Marshal(map[string]any{
			"skipped": true,
			"reason":  "image service not configured",
		})
	}
	var prompterOut *prompteragent.Output
	var script *scriptwriter.Output
	var dir *directoragent.Output
	_ = json.Unmarshal(rc.outputs["prompter"], &prompterOut)
	_ = json.Unmarshal(rc.outputs["script"], &script)
	_ = json.Unmarshal(rc.outputs["director"], &dir)
	if prompterOut == nil || script == nil {
		return nil, fmt.Errorf("missing prompter or script output")
	}
	out, err := e.image.GenerateScenes(ctx, rc.run.ImageProvider, rc.run.ID.String(), prompterOut, script, dir, rc.assets)
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (e *Executor) stepVideo(ctx context.Context, rc *runContext) ([]byte, error) {
	if e.video == nil {
		return json.Marshal(map[string]any{
			"skipped":  true,
			"reason":   "video service not configured",
			"provider": rc.run.VideoProvider,
		})
	}
	var prompterOut *prompteragent.Output
	var script *scriptwriter.Output
	var dir *directoragent.Output
	var imageOut *imagemedia.StepOutput
	_ = json.Unmarshal(rc.outputs["prompter"], &prompterOut)
	_ = json.Unmarshal(rc.outputs["script"], &script)
	_ = json.Unmarshal(rc.outputs["director"], &dir)
	_ = json.Unmarshal(rc.outputs["image"], &imageOut)
	if prompterOut == nil || script == nil {
		return nil, fmt.Errorf("missing prompter or script output")
	}
	out, err := e.video.GenerateScenes(ctx, rc.run.VideoProvider, rc.run.ID.String(), prompterOut, script, dir, imageOut, rc.assets)
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (e *Executor) stepSubtitles(ctx context.Context, rc *runContext) ([]byte, error) {
	var script *scriptwriter.Output
	_ = json.Unmarshal(rc.outputs["script"], &script)
	audioPath := e.storage.FilePath(rc.run.ID.String(), "narration.mp3")

	var out *subtitles.Output
	if e.subtitles != nil {
		res, err := e.subtitles.Generate(ctx, audioPath, script)
		if err != nil {
			return nil, err
		}
		out = res.Output
	} else {
		out = subtitles.FromScript(script)
		out.Source = "script"
	}

	srtBytes := subtitles.ToSRT(out)
	if len(srtBytes) > 0 {
		if _, err := e.storage.WriteFile(rc.run.ID.String(), "captions.srt", srtBytes); err != nil {
			return nil, err
		}
		out.SRTURL = e.storage.PublicPath(rc.run.ID.String(), "captions.srt")
	}
	return json.Marshal(out)
}

func (e *Executor) stepRender(ctx context.Context, rc *runContext) ([]byte, error) {
	var script *scriptwriter.Output
	var dir *directoragent.Output
	var caps *subtitles.Output
	var voice map[string]any
	_ = json.Unmarshal(rc.outputs["script"], &script)
	_ = json.Unmarshal(rc.outputs["director"], &dir)
	_ = json.Unmarshal(rc.outputs["subtitles"], &caps)
	_ = json.Unmarshal(rc.outputs["voice"], &voice)

	narrationURL, _ := voice["publicUrl"].(string)
	var videoOut *video.StepOutput
	_ = json.Unmarshal(rc.outputs["video"], &videoOut)
	introClip := introClipURL(rc.assets)
	manifest := rendermedia.BuildManifest(rendermedia.Input{
		RunID:           rc.run.ID.String(),
		ProductName:     rc.product.Name,
		NarrationURL:    narrationURL,
		IntroClip:       introClip,
		IntroDurationMs: e.cfg.IntroDurationMs,
		MediaBaseURL:    e.cfg.MediaBaseURL(),
		Script:          script,
		Director:        dir,
		Captions:        caps,
		SceneVideos:     video.ClipsBySceneID(videoOut),
	})
	manifestBytes, err := manifest.JSON()
	if err != nil {
		return nil, err
	}
	manifestPath, err := e.storage.WriteFile(rc.run.ID.String(), "manifest.json", manifestBytes)
	if err != nil {
		return nil, err
	}

	outputPath := e.storage.FilePath(rc.run.ID.String(), "draft.mp4")
	if err := e.runRender(ctx, rc.run.ID.String(), manifestPath, outputPath); err != nil {
		return nil, err
	}

	return json.Marshal(map[string]any{
		"manifestPath":    manifestPath,
		"draftUrl":        e.storage.PublicPath(rc.run.ID.String(), "draft.mp4"),
		"format":          dir.Format,
		"durationMs":      manifest.TotalDurationMs(),
		"introDurationMs": manifest.IntroDurationMs,
	})
}

func (e *Executor) runRender(ctx context.Context, runID, manifestPath, outputPath string) error {
	log := applog.FromContext(ctx).With("service", "remotion", "operation", "render")
	if e.cfg.RenderMock {
		log.Info("remotion render mock", "output", outputPath)
		return os.WriteFile(outputPath, []byte("MOCK_MP4_PHASE1"), 0o644)
	}
	renderDir := e.cfg.RenderDir
	script := filepath.Join(renderDir, "scripts", "render.mjs")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("render script not found: %s", script)
	}
	logDir := e.storage.LogsDir(runID)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, "render.log")
	started := time.Now()
	log.Info("remotion render started", "manifest", manifestPath, "output", outputPath, "log_path", logPath)

	cmd := exec.CommandContext(ctx, "node", script, manifestPath, outputPath)
	cmd.Dir = renderDir
	cmd.Env = append(os.Environ(),
		"API_PUBLIC_URL="+e.cfg.MediaBaseURL(),
		"STORAGE_DIR="+e.cfg.StorageDir,
		"CREATIVE_RUN_ID="+runID,
		"RENDER_LOG_PATH="+logPath,
		"RENDER_MOCK=0",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("remotion render failed",
			"duration_ms", time.Since(started).Milliseconds(),
			"output_preview", applog.Truncate(string(out), 500),
			"log_path", logPath,
		)
		return fmt.Errorf("remotion render: %w: %s", err, string(out))
	}
	log.Info("remotion render completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"output", outputPath,
		"log_path", logPath,
	)
	return nil
}

func (e *Executor) stepPostprocess(ctx context.Context, rc *runContext) ([]byte, error) {
	draftPath := e.storage.FilePath(rc.run.ID.String(), "draft.mp4")
	finalPath := e.storage.FilePath(rc.run.ID.String(), "final.mp4")
	thumbPath := e.storage.FilePath(rc.run.ID.String(), "thumbnail.jpg")

	var ppResult *postprocessmedia.Result
	var err error
	if e.postprocess != nil {
		ppResult, err = e.postprocess.Process(ctx, draftPath, finalPath, thumbPath)
	} else {
		ppResult, err = postprocessmedia.New(e.cfg).Process(ctx, draftPath, finalPath, thumbPath)
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]any{
		"finalVideoUrl":   e.storage.PublicPath(rc.run.ID.String(), "final.mp4"),
		"thumbnailUrl":    e.storage.PublicPath(rc.run.ID.String(), "thumbnail.jpg"),
		"loudnessApplied": ppResult.LoudnessApplied,
		"thumbnailFrom":   ppResult.ThumbnailFrom,
	})
}

func introClipURL(assets *models.RunAssets) string {
	if assets == nil {
		return ""
	}
	return assets.IntroClip
}

func (e *Executor) stepSupervisor(ctx context.Context, rc *runContext) ([]byte, error) {
	steps := map[string]any{}
	for k, v := range rc.outputs {
		var parsed any
		_ = json.Unmarshal(v, &parsed)
		steps[k] = parsed
	}
	out, err := e.supervisor.Execute(ctx, supervisoragent.Input{
		ProductName: rc.product.Name,
		Steps:       steps,
	})
	if err != nil {
		return nil, err
	}
	result, _ := json.Marshal(out)
	if out.Approved {
		_ = e.repo.UpdateRunStatus(ctx, rc.run.ID, models.RunStatusApproved)
	} else {
		_ = e.repo.UpdateRunStatus(ctx, rc.run.ID, models.RunStatusNeedsReview)
	}
	return result, nil
}
