package executor

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/elevenlabs"
	rendermedia "github.com/woragis/ecom-op-creatives-backend/server/internal/media/render"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/subtitles"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	pipelinesvc "github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline/service"
	productrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/product/repository"
)

type Executor struct {
	cfg       config.Config
	repo      *creativerunrepo.Repository
	products  *productrepo.Repository
	pipeline  *pipelinesvc.Service
	storage   *storage.Local
	tts       *elevenlabs.Client
	research  *researchagent.Agent
	hooks     *hooksagent.Agent
	script    *scriptagent.Agent
	director  *directoragent.Agent
	prompter  *prompteragent.Agent
	supervisor *supervisoragent.Agent
}

type Deps struct {
	Cfg        config.Config
	Repo       *creativerunrepo.Repository
	Products   *productrepo.Repository
	Pipeline   *pipelinesvc.Service
	Storage    *storage.Local
	TTS        *elevenlabs.Client
	Research   *researchagent.Agent
	Hooks      *hooksagent.Agent
	Script     *scriptagent.Agent
	Director   *directoragent.Agent
	Prompter   *prompteragent.Agent
	Supervisor *supervisoragent.Agent
}

func New(d Deps) *Executor {
	return &Executor{
		cfg: d.Cfg, repo: d.Repo, products: d.Products, pipeline: d.Pipeline,
		storage: d.Storage, tts: d.TTS,
		research: d.Research, hooks: d.Hooks, script: d.Script,
		director: d.Director, prompter: d.Prompter, supervisor: d.Supervisor,
	}
}

type runContext struct {
	run     *models.CreativeRun
	product *models.Product
	outputs map[string]json.RawMessage
}

func (e *Executor) ProcessStep(ctx context.Context, stepID uuid.UUID) error {
	step, err := e.repo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step.Status == models.StepStatusDone {
		return e.advance(ctx, step)
	}

	if err := e.repo.MarkStepRunning(ctx, stepID); err != nil {
		return err
	}

	rc, err := e.loadContext(ctx, step.CreativeRunID)
	if err != nil {
		return err
	}

	output, err := e.executeStep(ctx, step.StepType, rc)
	if err != nil {
		_ = e.repo.FailStep(ctx, stepID, err.Error())
		return err
	}

	if err := e.repo.CompleteStep(ctx, stepID, output); err != nil {
		return err
	}
	return e.advance(ctx, step)
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
			return e.repo.UpdateRunStatus(ctx, step.CreativeRunID, models.RunStatusNeedsReview)
		}
		return nil
	}
	return e.pipeline.EnqueueStep(ctx, next)
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
	return &runContext{run: run, product: product, outputs: outputs}, nil
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
		Script:   script,
		Director: dir,
		Product:  rc.product.Name,
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
	audio, err := e.tts.Synthesize(ctx, text)
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
		"provider":    "elevenlabs",
	})
}

func (e *Executor) stepImage(ctx context.Context, rc *runContext) ([]byte, error) {
	_ = ctx
	return json.Marshal(map[string]any{
		"skipped": true,
		"reason":  "phase1-uses-remotion-backgrounds",
	})
}

func (e *Executor) stepVideo(ctx context.Context, rc *runContext) ([]byte, error) {
	_ = ctx
	return json.Marshal(map[string]any{
		"skipped":  true,
		"reason":   "phase1-no-ai-video",
		"provider": rc.run.VideoProvider,
	})
}

func (e *Executor) stepSubtitles(ctx context.Context, rc *runContext) ([]byte, error) {
	_ = ctx
	var script *scriptwriter.Output
	_ = json.Unmarshal(rc.outputs["script"], &script)
	out := subtitles.FromScript(script)
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
	manifest := rendermedia.BuildManifest(rc.run.ID.String(), rc.product.Name, narrationURL, script, dir, caps)
	manifestBytes, err := manifest.JSON()
	if err != nil {
		return nil, err
	}
	manifestPath, err := e.storage.WriteFile(rc.run.ID.String(), "manifest.json", manifestBytes)
	if err != nil {
		return nil, err
	}

	outputPath := e.storage.FilePath(rc.run.ID.String(), "draft.mp4")
	if err := e.runRender(ctx, manifestPath, outputPath); err != nil {
		return nil, err
	}

	return json.Marshal(map[string]any{
		"manifestPath": manifestPath,
		"draftUrl":     e.storage.PublicPath(rc.run.ID.String(), "draft.mp4"),
		"format":       dir.Format,
	})
}

func (e *Executor) runRender(ctx context.Context, manifestPath, outputPath string) error {
	if os.Getenv("RENDER_MOCK") == "1" || os.Getenv("RENDER_MOCK") == "true" {
		return os.WriteFile(outputPath, []byte("MOCK_MP4_PHASE1"), 0o644)
	}
	renderDir := e.cfg.RenderDir
	script := filepath.Join(renderDir, "scripts", "render.mjs")
	if _, err := os.Stat(script); err != nil {
		return os.WriteFile(outputPath, []byte("MOCK_MP4_PHASE1"), 0o644)
	}
	cmd := exec.CommandContext(ctx, "node", script, manifestPath, outputPath)
	cmd.Dir = renderDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("remotion render: %w: %s", err, string(out))
	}
	return nil
}

func (e *Executor) stepPostprocess(ctx context.Context, rc *runContext) ([]byte, error) {
	_ = ctx
	draftPath := e.storage.FilePath(rc.run.ID.String(), "draft.mp4")
	finalPath := e.storage.FilePath(rc.run.ID.String(), "final.mp4")

	data, err := os.ReadFile(draftPath)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(finalPath, data, 0o644); err != nil {
		return nil, err
	}

	thumbPath := e.storage.FilePath(rc.run.ID.String(), "thumbnail.jpg")
	_ = os.WriteFile(thumbPath, []byte("MOCK_JPG"), 0o644)

	return json.Marshal(map[string]any{
		"finalVideoUrl": e.storage.PublicPath(rc.run.ID.String(), "final.mp4"),
		"thumbnailUrl":  e.storage.PublicPath(rc.run.ID.String(), "thumbnail.jpg"),
	})
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
