package app

import (
	hooksagent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/hooks"
	directoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	researchagent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/research"
	scriptagent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	supervisoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/supervisor"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/serper"
	creativerunrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/repository"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/elevenlabs"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/video"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline/executor"
	pipelinesvc "github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline/service"
	productrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/product/repository"
)

func NewExecutor(
	cfg config.Config,
	repo *creativerunrepo.Repository,
	products *productrepo.Repository,
	pipelineSvc *pipelinesvc.Service,
) (*executor.Executor, error) {
	store, err := storage.NewLocal(cfg.StorageDir)
	if err != nil {
		return nil, err
	}
	llmClient := llm.NewFromConfig(cfg)
	serperClient := serper.NewFromConfig(cfg)
	tts := elevenlabs.New(cfg.ElevenLabsKey, cfg.ElevenLabsVoice, cfg.ElevenLabsMock || cfg.ElevenLabsKey == "")
	videoRegistry := video.NewRegistry(cfg)
	videoSvc := video.NewService(cfg, videoRegistry, store)

	return executor.New(executor.Deps{
		Cfg:        cfg,
		Repo:       repo,
		Products:   products,
		Pipeline:   pipelineSvc,
		Storage:    store,
		TTS:        tts,
		Research:   researchagent.New(llmClient, serperClient),
		Hooks:      hooksagent.New(llmClient),
		Script:     scriptagent.New(llmClient),
		Director:   directoragent.New(llmClient),
		Prompter:   prompteragent.New(llmClient),
		Supervisor: supervisoragent.New(llmClient, cfg.SupervisorMin),
		Video:      videoSvc,
	}), nil
}
