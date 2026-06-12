-- Phase 0 foundation: products, campaigns, creative runs, pipeline steps

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT,
    niche TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    scheduled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_campaigns_product_id ON campaigns(product_id);

CREATE TABLE IF NOT EXISTS creative_runs (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    campaign_id UUID REFERENCES campaigns(id) ON DELETE SET NULL,
    video_provider TEXT NOT NULL DEFAULT 'kling',
    status TEXT NOT NULL DEFAULT 'draft',
    hook TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_creative_runs_product_id ON creative_runs(product_id);
CREATE INDEX IF NOT EXISTS idx_creative_runs_status ON creative_runs(status);

CREATE TABLE IF NOT EXISTS pipeline_steps (
    id UUID PRIMARY KEY,
    creative_run_id UUID NOT NULL REFERENCES creative_runs(id) ON DELETE CASCADE,
    step_type TEXT NOT NULL,
    step_order INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    input_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    output_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempt_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    provider_used TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (creative_run_id, step_type)
);

CREATE INDEX IF NOT EXISTS idx_pipeline_steps_run_id ON pipeline_steps(creative_run_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_steps_status ON pipeline_steps(status);
