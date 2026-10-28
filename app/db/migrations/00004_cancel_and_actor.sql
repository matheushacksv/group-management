-- +goose Up
ALTER TABLE removal_items
    DROP CONSTRAINT removal_items_status_check,
    ADD CONSTRAINT removal_items_status_check CHECK (status IN ('pending', 'ok', 'failed', 'canceled'));

-- who asked for the removal (email of the panel user), NULL for jobs created before this column
ALTER TABLE removal_jobs ADD COLUMN requested_by text;

-- +goose Down
UPDATE removal_items SET status = 'failed', error = 'canceled' WHERE status = 'canceled';
ALTER TABLE removal_items
    DROP CONSTRAINT removal_items_status_check,
    ADD CONSTRAINT removal_items_status_check CHECK (status IN ('pending', 'ok', 'failed'));

ALTER TABLE removal_jobs DROP COLUMN requested_by;
