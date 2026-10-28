package groups

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Worker struct {
	WA   *Client
	Pool *pgxpool.Pool
}

// Run until context be canceled, the item in progress is always finished before returning
func (w Worker) Run(ctx context.Context) {
	for {
		// WithoutCancel: a shutdown in the middle of a removal must not lose its result
		worked, err := w.next(context.WithoutCancel(ctx))
		if err != nil {
			log.Printf("worker: %v", err)
		}

		wait := 5 * time.Second

		if worked {
			wait = 3*time.Second + rand.N(3*time.Second)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

// next process a pending item, returns false if queue is empty
func (w Worker) next(ctx context.Context) (bool, error) {
	tx, err := w.Pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	var itemID int64
	var memberID, sessionID, groupID string

	err = tx.QueryRow(
		ctx,
		`
		SELECT i.id, i.member_wa_id, j.session_id, j.group_wa_id
		FROM removal_items i JOIN removal_jobs j ON j.id = i.job_id
		WHERE i.status = 'pending' ORDER BY i.id
		LIMIT 1
		FOR UPDATE OF i SKIP LOCKED
		`,
	).Scan(&itemID, &memberID, &sessionID, &groupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	status, reason := "ok", ""
	results, err := w.WA.RemoveParticipants(
		sessionID, groupID, []string{memberID},
	)
	switch {
	case err != nil:
		status, reason = "failed", err.Error()
	case len(results) == 0 || !results[0].Success:
		status, reason = "failed", "openwa returned no success"
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE removal_items
		SET status = $2, error = NULLIF($3, ''), processed_at = now()
		WHERE id = $1
		`, itemID, status, reason,
	)
	if err != nil {
		return true, err
	}
	return true, tx.Commit(ctx)
}
