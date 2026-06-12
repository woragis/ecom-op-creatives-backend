-- Phase 3: image provider + user input assets

ALTER TABLE creative_runs
    ADD COLUMN IF NOT EXISTS image_provider TEXT NOT NULL DEFAULT 'flux';

ALTER TABLE creative_runs
    ADD COLUMN IF NOT EXISTS input_assets JSONB NOT NULL DEFAULT '{}'::jsonb;
