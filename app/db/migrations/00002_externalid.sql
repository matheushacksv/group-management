-- +goose Up
ALTER TABLE membership_events
    ADD COLUMN external_id text UNIQUE,
    ADD COLUMN actor_wa_id text;

-- +goose Down
ALTER TABLE membership_events
    DROP COLUMN external_id,
    DROP COLUMN actor_wa_id;
