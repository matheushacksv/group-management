-- +goose Up
CREATE TABLE removal_jobs (
    id bigserial PRIMARY KEY,
    session_id text NOT NULL,
    group_wa_id text NOT NULL REFERENCES groups(wa_id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE removal_items (
    id bigserial PRIMARY KEY,
    job_id bigint NOT NULL REFERENCES removal_jobs(id),
    member_wa_id text NOT NULL,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'ok', 'failed')),
    error text,
    processed_at timestamptz
);
CREATE INDEX ON removal_items (status, id);

-- +goose Down
DROP TABLE removal_items, removal_jobs;
