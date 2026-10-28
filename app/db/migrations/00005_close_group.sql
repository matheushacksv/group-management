-- +goose Up
-- when the group was closed (everybody removed and the bot left), NULL while active
ALTER TABLE groups ADD COLUMN closed_at timestamptz;

-- removal created by "close group": when it finishes without cancel, the bot leaves the group
ALTER TABLE removal_jobs ADD COLUMN close_group boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE removal_jobs DROP COLUMN close_group;
ALTER TABLE groups DROP COLUMN closed_at;
