package groups

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var errSyncRunning = errors.New("sync already running")

// syncMu prevents the manual and the automatic sync from running at the same time
// ponytail: in-process lock, fine with a single API instance; use pg_try_advisory_lock if it scales out
var syncMu sync.Mutex

// syncAll saves every group of the session and the members of the managed ones, returns how many groups were saved
func (g GroupHandler) syncAll(ctx context.Context, sessionID string) (int, error) {
	if !syncMu.TryLock() {
		return 0, errSyncRunning
	}
	defer syncMu.Unlock()

	groups, err := g.WA.ListGroups(sessionID)
	if err != nil {
		return 0, err
	}
	for _, group := range groups {
		if err := SaveGroup(ctx, g.Pool, group); err != nil {
			return 0, err
		}
	}

	ids, err := ManagedGroupsIDS(ctx, g.Pool)
	if err != nil {
		return len(groups), err
	}

	names := g.contactNames(sessionID)

	for _, id := range ids {
		if err := g.syncGroup(ctx, sessionID, id, names); err != nil {
			log.Printf("sync %s: %v", id, err)
		}

		// pause between groups to not hammer WhatsApp, stops early on shutdown
		select {
		case <-ctx.Done():
			return len(groups), ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return len(groups), nil
}

// RunAutoSync syncs the session every interval until ctx is canceled
// keeps names fresh and catches exits missed while the API was down
func (g GroupHandler) RunAutoSync(ctx context.Context, sessionID string, every time.Duration) {
	if sessionID == "" || every <= 0 {
		log.Println("auto sync disabled (set OPENWA_SESSION_ID and SYNC_INTERVAL)")
		return
	}
	log.Printf("auto sync every %s", every)

	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		start := time.Now()
		saved, err := g.syncAll(ctx, sessionID)
		switch {
		case errors.Is(err, errSyncRunning):
			log.Println("auto sync skipped: manual sync running")
		case err != nil:
			log.Printf("auto sync: %v", err)
		default:
			log.Printf("auto sync: %d groups in %s", saved, time.Since(start).Round(time.Second))
		}
	}
}
