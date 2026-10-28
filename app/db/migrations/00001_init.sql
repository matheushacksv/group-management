-- +goose Up
CREATE TABLE groups(
    wa_id text PRIMARY KEY,
    name  text NOT NULL,
    managed boolean NOT NULL DEFAULT false,
    synced_at timestamptz
);

CREATE TABLE members(
    wa_id text PRIMARY KEY,
    number text NOT NULL,
    push_name text
);

CREATE TABLE membership_events(
    id bigserial PRIMARY KEY,
    group_wa_id text NOT NULL REFERENCES groups(wa_id),
    member_wa_id text NOT NULL REFERENCES members(wa_id),
    action text NOT NULL CHECK (action IN ('seen', 'join', 'leave', 'remove')),
    source text NOT NULL CHECK (source IN ('sync', 'webhook')),
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ON membership_events (group_wa_id, member_wa_id, occurred_at);

-- +goose Down
DROP TABLE membership_events, members, groups;
