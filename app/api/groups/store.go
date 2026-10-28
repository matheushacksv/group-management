package groups

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SaveGroup saves all groups from a session in database
func SaveGroup(ctx context.Context, pool *pgxpool.Pool, g Group) error {
	_, err := pool.Exec(
		ctx,
		`
		INSERT INTO groups (wa_id, name, synced_at) VALUES ($1, $2, now())
		ON CONFLICT (wa_id) DO UPDATE SET name = EXCLUDED.name, synced_at = now()
		`, g.ID, g.Name,
	)
	return err
}

// ManagedGroupsIDS returns all groups ids where session is admin
func ManagedGroupsIDS(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(
		ctx,
		`SELECT wa_id FROM groups WHERE managed`,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// SaveMember saves a member from a group in database, keeping the old push name when the new one is nil
func SaveMember(ctx context.Context, pool *pgxpool.Pool, p Participant, pushName *string) error {
	_, err := pool.Exec(
		ctx,
		`INSERT INTO members (wa_id, number, push_name) VALUES ($1, $2, $3)
		ON CONFLICT (wa_id) DO UPDATE SET push_name = COALESCE(EXCLUDED.push_name, members.push_name)`,
		p.ID, p.Number, pushName,
	)
	return err
}

// ListStoredGroups lists the groups saved in database
func ListStoredGroups(ctx context.Context, pool *pgxpool.Pool) ([]StoredGroup, error) {
	rows, err := pool.Query(
		ctx,
		`SELECT wa_id, name, managed, synced_at FROM groups ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[StoredGroup])
}

// UpdateManaged sets if the group is controlled by the system, returns false if the group doesn't exist
func UpdateManaged(ctx context.Context, pool *pgxpool.Pool, groupID string, managed bool) (bool, error) {
	tag, err := pool.Exec(
		ctx,
		`UPDATE groups SET managed = $2 WHERE wa_id = $1`,
		groupID, managed,
	)
	return tag.RowsAffected() > 0, err
}

// SaveSeen saves a "seen" in the groups one time
func SaveSeen(ctx context.Context, pool *pgxpool.Pool, groupID, memberID string) error {
	_, err := pool.Exec(
		ctx,
		`INSERT INTO membership_events (group_wa_id, member_wa_id, action, source)
		SELECT $1, $2, 'seen', 'sync'
		WHERE NOT EXISTS (SELECT 1 FROM membership_events WHERE group_wa_id = $1 AND member_wa_id = $2)
		`,
		groupID, memberID,
	)
	return err
}

// IsManaged returns if the group is controlled by the system
func IsManaged(ctx context.Context, pool *pgxpool.Pool, groupID string) (bool, error) {
	var ok bool
	err := pool.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM groups WHERE wa_id = $1 AND managed)`,
		groupID,
	).Scan(&ok)
	return ok, err
}

// SaveEvent saves a event from webhook
func SaveEvent(ctx context.Context, pool *pgxpool.Pool, groupID, memberID, action, actorID, externalID string, ts int64) error {
	_, err := pool.Exec(
		ctx,
		`INSERT INTO membership_events (group_wa_id, member_wa_id, action, source, actor_wa_id, occurred_at, external_id)
		VALUES ($1, $2, $3, 'webhook', $4, to_timestamp($5), $6)
		ON CONFLICT (external_id) DO NOTHING
		`,
		groupID, memberID, action, actorID, ts, externalID,
	)
	return err
}

// CurrentMembers list whos in the group right now
func CurrentMembers(ctx context.Context, pool *pgxpool.Pool, groupID string, minDays int) ([]Member, error) {
	rows, err := pool.Query(
		ctx,
		`WITH last AS (
		SELECT DISTINCT ON (member_wa_id) member_wa_id, action FROM membership_events
		WHERE group_wa_id = $1 ORDER BY member_wa_id, occurred_at DESC, id DESC),
		since AS (SELECT member_wa_id, max(occurred_at) as since FROM membership_events
		WHERE group_wa_id = $1 AND action IN ('join', 'seen') GROUP BY member_wa_id)
		SELECT m.wa_id, m.number, m.push_name, s.since FROM last l JOIN since s USING (member_wa_id)
		JOIN members m ON m.wa_id = l.member_wa_id WHERE l.action IN ('join', 'seen')
		AND s.since <= now() - make_interval(days => $2) ORDER BY s.since`,
		groupID, minDays,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Member])
}

// ReturnedMembers lists who leave and enters again
func ReturnedMembers(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]Returned, error) {
	rows, err := pool.Query(
		ctx,
		`SELECT m.wa_id, m.number, m.push_name,
		       count(*)           FILTER (WHERE e.action IN ('leave', 'remove')) AS exits,
		       max(e.occurred_at) FILTER (WHERE e.action IN ('leave', 'remove')) AS last_exit,
		       max(e.occurred_at) FILTER (WHERE e.action = 'join')               AS last_join
		FROM membership_events e
		JOIN members m ON m.wa_id = e.member_wa_id
		WHERE e.group_wa_id = $1
		GROUP BY m.wa_id, m.number, m.push_name
		HAVING max(e.occurred_at) FILTER (WHERE e.action = 'join')
		     > min(e.occurred_at) FILTER (WHERE e.action IN ('leave', 'remove'))
		ORDER BY last_join DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Returned])
}

// CreateRemovalJob saves the Job and Items
func CreateRemovalJob(ctx context.Context, pool *pgxpool.Pool, sessionID, groupID, requestedBy string, members []string) (int64, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var jobID int64

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO removal_jobs (session_id, group_wa_id, requested_by) VALUES ($1, $2, NULLIF($3, '')) RETURNING id
		`, sessionID, groupID, requestedBy,
	).Scan(&jobID)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO removal_items (job_id, member_wa_id) SELECT $1, unnest ($2::text[])`,
		jobID, members,
	)
	if err != nil {
		return 0, err
	}
	return jobID, tx.Commit(ctx)
}

// RemovalItems lists items of a job
func RemovalItems(ctx context.Context, pool *pgxpool.Pool, jobID int64) ([]RemovalItem, error) {
	rows, err := pool.Query(
		ctx,
		`
		SELECT member_wa_id, status, error, processed_at
		FROM removal_items
		WHERE job_id = $1
		ORDER BY id
		`, jobID,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[RemovalItem])
}

// RemovalJobs lists the removal jobs of a group with the count of each status
func RemovalJobs(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]RemovalJob, error) {
	rows, err := pool.Query(
		ctx,
		`
		SELECT j.id, j.created_at, j.requested_by,
		       count(*) FILTER (WHERE i.status = 'pending')  AS pending,
		       count(*) FILTER (WHERE i.status = 'ok')       AS ok,
		       count(*) FILTER (WHERE i.status = 'failed')   AS failed,
		       count(*) FILTER (WHERE i.status = 'canceled') AS canceled
		FROM removal_jobs j
		JOIN removal_items i ON i.job_id = j.id
		WHERE j.group_wa_id = $1
		GROUP BY j.id
		ORDER BY j.id DESC
		`, groupID,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[RemovalJob])
}

// RemovalJobInfo returns when and by whom a removal job was created
func RemovalJobInfo(ctx context.Context, pool *pgxpool.Pool, jobID int64) (createdAt time.Time, requestedBy *string, err error) {
	err = pool.QueryRow(
		ctx,
		`SELECT created_at, requested_by FROM removal_jobs WHERE id = $1`,
		jobID,
	).Scan(&createdAt, &requestedBy)
	return createdAt, requestedBy, err
}

// CancelRemovalJob cancels the pending items of a job, returns how many were canceled
// the item the worker is processing is locked (FOR UPDATE): this UPDATE waits for it and,
// as it is no longer pending after the commit, it keeps its real result
func CancelRemovalJob(ctx context.Context, pool *pgxpool.Pool, jobID int64) (int64, error) {
	tag, err := pool.Exec(
		ctx,
		`
		UPDATE removal_items
		SET status = 'canceled', processed_at = now()
		WHERE job_id = $1 AND status = 'pending'
		`, jobID,
	)
	return tag.RowsAffected(), err
}

// ReconcileGroup fixes events missed while the API was down, comparing the diary with the current participants:
// in the diary but not in the group gets a "leave", back in the group after an exit gets a "join"
// the sync doesn't know if it was a leave or a remove, nor the real time, so source = 'sync' and no actor
func ReconcileGroup(ctx context.Context, pool *pgxpool.Pool, groupID string, current []string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO membership_events (group_wa_id, member_wa_id, action, source)
		SELECT $1, l.member_wa_id, 'leave', 'sync'
		FROM (
			SELECT DISTINCT ON (member_wa_id) member_wa_id, action
			FROM membership_events
			WHERE group_wa_id = $1
			ORDER BY member_wa_id, occurred_at DESC, id DESC
		) l
		WHERE l.action IN ('join', 'seen')
		  AND l.member_wa_id <> ALL ($2::text[])
		`, groupID, current,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO membership_events (group_wa_id, member_wa_id, action, source)
		SELECT $1, l.member_wa_id, 'join', 'sync'
		FROM (
			SELECT DISTINCT ON (member_wa_id) member_wa_id, action
			FROM membership_events
			WHERE group_wa_id = $1
			ORDER BY member_wa_id, occurred_at DESC, id DESC
		) l
		WHERE l.action IN ('leave', 'remove')
		  AND l.member_wa_id = ANY ($2::text[])
		`, groupID, current,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
