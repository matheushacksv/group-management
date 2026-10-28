package groups

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxRemovals = 500

type GroupHandler struct {
	WA           *Client
	Pool         *pgxpool.Pool
	WebhookToken string
}

// List handles http and list all groups saved in database, run Sync to refresh them
func (g GroupHandler) List(c *gin.Context) {
	groups, err := ListStoredGroups(c.Request.Context(), g.Pool)
	if err != nil {
		log.Printf("groups list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// SetManaged handles http and sets if the group is controlled by the system
func (g GroupHandler) SetManaged(c *gin.Context) {
	var body struct {
		Managed *bool `json:"managed"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.Managed == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "managed must be a boolean"})
		return
	}

	groupID := c.Param("groupID")
	found, err := UpdateManaged(c.Request.Context(), g.Pool, groupID, *body.Managed)
	if err != nil {
		log.Printf("set managed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found, run sync first"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": groupID, "managed": *body.Managed})
}

// Get handles http and get detailed info from a group
func (g GroupHandler) Get(c *gin.Context) {
	sessionID := c.Param("sessionID")
	groupID := c.Param("groupID")

	group, err := g.WA.GetGroup(sessionID, groupID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

// RemoveParticipant handles http and remove a participant from a group
func (g GroupHandler) RemoveParticipant(c *gin.Context) {
	sessionID := c.Param("sessionID")
	groupID := c.Param("groupID")

	var body struct {
		Particpants []string `json:"participants"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.Particpants) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no participants"})
		return
	}

	results, err := g.WA.RemoveParticipants(sessionID, groupID, body.Particpants)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// Sync handles http and saves the groups in database
func (g GroupHandler) Sync(c *gin.Context) {
	saved, err := g.syncAll(c.Request.Context(), c.Param("sessionID"))
	if errors.Is(err, errSyncRunning) {
		c.JSON(http.StatusConflict, gin.H{"error": "sincronização já em andamento, aguarde terminar"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"saved": saved})
}

// Webhook recieves events from OpenWa
func (g GroupHandler) Webhook(c *gin.Context) {
	if subtle.ConstantTimeCompare([]byte(c.Query("token")), []byte(g.WebhookToken)) != 1 {
		c.Status(http.StatusUnauthorized)
		return
	}

	var ev WebhookEvent
	if err := c.ShouldBindJSON(&ev); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if ev.Event == "message.received" {
		g.saveSenderName(c.Request.Context(), ev)
		c.Status(http.StatusOK)
		return
	}

	if ev.Event != "group.join" && ev.Event != "group.leave" {
		c.Status(http.StatusOK)
		return
	}

	ctx := c.Request.Context()
	managed, err := IsManaged(
		ctx, g.Pool, ev.Data.GroupID,
	)
	if err != nil {
		log.Printf("webhook managed: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	if !managed {
		c.Status(http.StatusOK)
		return
	}

	for _, pid := range ev.Data.ParticipantIDs {
		action := "join"
		if ev.Event == "group.leave" {
			action = "leave"
			if ev.Data.ActorID != pid {
				action = "remove"
			}
		}

		p := Participant{ID: pid, Number: strings.TrimSuffix(pid, "@c.us")}
		if err := SaveMember(ctx, g.Pool, p, nil); err != nil {
			log.Printf("webhook member %s: %v", pid, err)
			c.Status(http.StatusInternalServerError)
			return
		}

		extID := ev.IdempotencyKey + "_" + pid
		if err := SaveEvent(ctx, g.Pool, ev.Data.GroupID, pid, action, ev.Data.ActorID, extID, ev.Data.Timestamp); err != nil {
			log.Printf("webhook event %s: %v", pid, err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}
	c.Status(http.StatusOK)
}

// syncGroup saves the current participants of one group
func (g GroupHandler) syncGroup(ctx context.Context, sessionID, groupID string, names map[string]*string) error {
	group, err := g.WA.GetGroup(sessionID, groupID)
	if err != nil {
		return err
	}

	for _, p := range group.Participants {
		if err := SaveMember(
			ctx,
			g.Pool,
			p,
			names[p.ID],
		); err != nil {
			log.Printf("member %s: %v", p.ID, err)
			continue
		}
		if err := SaveSeen(ctx, g.Pool, groupID, p.ID); err != nil {
			log.Printf("seen %s: %v", p.ID, err)
		}
	}

	// the bot is always in the group, an empty list means OpenWA failed:
	// reconciling it would mark everybody as gone
	if len(group.Participants) == 0 {
		return errors.New("openwa returned no participants, reconcile skipped")
	}

	current := make([]string, 0, len(group.Participants))
	for _, p := range group.Participants {
		current = append(current, p.ID)
	}
	return ReconcileGroup(ctx, g.Pool, groupID, current)
}

// contactNames maps contact id to push name (or saved name), one request for all contacts
// without names the sync still works, so errors are only logged
func (g GroupHandler) contactNames(sessionID string) map[string]*string {
	names := map[string]*string{}
	contacts, err := g.WA.ListContacts(sessionID)
	if err != nil {
		log.Printf("sync contacts: %v", err)
	}
	for _, ct := range contacts {
		name := ct.PushName
		if name == "" {
			name = ct.Name
		}
		if name != "" {
			names[ct.ID] = &name
		}
	}
	return names
}

// SyncGroup handles http and syncs the members of a single managed group
func (g GroupHandler) SyncGroup(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID, groupID := c.Param("sessionID"), c.Param("groupID")

	managed, err := IsManaged(ctx, g.Pool, groupID)
	if err != nil {
		log.Printf("sync group managed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if !managed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group is not managed"})
		return
	}

	if err := g.syncGroup(ctx, sessionID, groupID, g.contactNames(sessionID)); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// saveSenderName stores the name of who sent a message in a managed group.
// WhatsApp only exposes the name of people who interact, so names fill up as members talk.
// Errors are only logged: a missing name is not worth an OpenWA retry
func (g GroupHandler) saveSenderName(ctx context.Context, ev WebhookEvent) {
	d := ev.Data
	if !d.IsGroup || d.FromMe || d.Author == "" {
		return
	}
	// ponytail: members are keyed by @c.us; @lid senders are skipped until LIDs show up in the groups
	if !strings.HasSuffix(d.Author, "@c.us") {
		return
	}

	name := d.Contact.PushName
	if name == "" {
		name = d.Contact.Name
	}
	if name == "" {
		return
	}

	managed, err := IsManaged(ctx, g.Pool, d.ChatID)
	if err != nil {
		log.Printf("webhook sender managed: %v", err)
		return
	}
	if !managed {
		return
	}

	if err := UpdatePushName(ctx, g.Pool, d.Author, name); err != nil {
		log.Printf("webhook sender name %s: %v", d.Author, err)
	}
}

// Members get from database members of the group
func (g GroupHandler) Members(c *gin.Context) {
	minDays, err := strconv.Atoi(c.DefaultQuery("min_days", "0"))
	if err != nil || minDays < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_days must be a non-negative integer"})
		return
	}
	members, err := CurrentMembers(
		c.Request.Context(),
		g.Pool,
		c.Param("groupID"),
		minDays,
	)
	if err != nil {
		log.Printf("members: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, members)
}

// Returned gets the returners of the group
func (g GroupHandler) Returned(c *gin.Context) {
	returners, err := ReturnedMembers(
		c.Request.Context(),
		g.Pool,
		c.Param("groupID"),
	)
	if err != nil {
		log.Printf("returners: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, returners)
}

func (g GroupHandler) CreateRemoval(c *gin.Context) {
	sessionID, groupID := c.Param("sessionID"), c.Param("groupID")

	var body struct {
		Participants []string `json:"participants"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(body.Participants) == 0 || len(body.Participants) > maxRemovals {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf(
			"participants must have 1 to %d items", maxRemovals,
		)})
		return
	}

	managed, err := IsManaged(
		c.Request.Context(),
		g.Pool,
		groupID,
	)
	if err != nil {
		log.Printf("removal managed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if !managed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group is not managed"})
		return
	}

	group, err := g.WA.GetGroup(sessionID, groupID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	inGroup := map[string]Participant{}

	for _, p := range group.Participants {
		inGroup[p.ID] = p
	}

	var queue []string
	skipped := []gin.H{}
	seen := map[string]bool{}

	for _, n := range body.Participants {
		id := n
		if !strings.Contains(id, "@") {
			id += "@c.us"
		}
		if seen[id] {
			continue
		}
		seen[id] = true

		p, ok := inGroup[id]
		switch {
			case !ok:
				skipped = append(skipped, gin.H{"id": id, "reason": "not in group"})
			case p.IsAdmin || p.IsSuperAdmin:
				skipped = append(skipped, gin.H{"id": id, "reason": "admin"})
			default:
				queue = append(queue, id)
		}
	}

	if len(queue) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "nothing to remove", "skipped": skipped,
		})
		return
	}

	// X-Actor is set by the Nuxt proxy with the logged user's email, the only client holding the API key
	jobID, err := CreateRemovalJob(
		c.Request.Context(),
		g.Pool,
		sessionID,
		groupID,
		c.GetHeader("X-Actor"),
		queue,
		false,
	)
	if err != nil {
		log.Printf("removal create: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"jobId": jobID, "queued": len(queue), "skipped": skipped})
}

func (g GroupHandler) GetRemoval(c *gin.Context) {
	jobID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	items, err := RemovalItems(
		c.Request.Context(), g.Pool, jobID,
	)
	if err != nil {
		log.Printf("removal get: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	counts := map[string]int{}

	for _, it := range items {
		counts[it.Status]++
	}

	createdAt, requestedBy, closeGroup, err := RemovalJobInfo(c.Request.Context(), g.Pool, jobID)
	if err != nil {
		log.Printf("removal info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobId":       jobID,
		"createdAt":   createdAt,
		"requestedBy": requestedBy,
		"closeGroup":  closeGroup,
		"counts":      counts,
		"items":       items,
	})
}

// ActiveRemovals handles http and lists the removals still running (sidebar indicator and finish notice)
func (g GroupHandler) ActiveRemovals(c *gin.Context) {
	active, err := ActiveRemovals(c.Request.Context(), g.Pool)
	if err != nil {
		log.Printf("removals active: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, active)
}

// SearchPeople handles http and finds members across all groups by number or name
func (g GroupHandler) SearchPeople(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if len([]rune(q)) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "busca precisa de ao menos 2 caracteres"})
		return
	}

	people, err := SearchPeople(c.Request.Context(), g.Pool, q)
	if err != nil {
		log.Printf("people search: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if people == nil {
		people = []Person{}
	}
	c.JSON(http.StatusOK, people)
}

// CloseGroup handles http and closes a group: removes everybody (admins included) through the
// removal queue and, when it finishes without cancel, the bot leaves the group.
// WhatsApp doesn't let anyone remove the group creator, so the creator stays
func (g GroupHandler) CloseGroup(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID, groupID := c.Param("sessionID"), c.Param("groupID")

	managed, err := IsManaged(ctx, g.Pool, groupID)
	if err != nil {
		log.Printf("close managed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if !managed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group is not managed"})
		return
	}

	running, err := HasPendingRemoval(ctx, g.Pool, groupID)
	if err != nil {
		log.Printf("close pending: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if running {
		c.JSON(http.StatusConflict, gin.H{"error": "há uma remoção em andamento neste grupo, aguarde ou cancele"})
		return
	}

	phone, err := g.WA.SessionPhone(sessionID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	botID := phone + "@c.us"

	group, err := g.WA.GetGroup(sessionID, groupID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	var queue []string
	kept := []gin.H{}
	botIsAdmin := false
	for _, p := range group.Participants {
		switch {
		case p.ID == botID:
			botIsAdmin = p.IsAdmin || p.IsSuperAdmin
		case p.IsSuperAdmin:
			kept = append(kept, gin.H{"id": p.ID, "reason": "criador do grupo"})
		default:
			queue = append(queue, p.ID)
		}
	}
	if !botIsAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "o bot precisa ser admin do grupo para encerrá-lo"})
		return
	}

	// nobody to remove: leave right away
	if len(queue) == 0 {
		if err := g.WA.LeaveGroup(sessionID, groupID); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		if err := MarkGroupClosed(ctx, g.Pool, groupID); err != nil {
			log.Printf("close mark: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"closed": true, "queued": 0, "kept": kept})
		return
	}

	jobID, err := CreateRemovalJob(ctx, g.Pool, sessionID, groupID, c.GetHeader("X-Actor"), queue, true)
	if err != nil {
		log.Printf("close create: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"jobId": jobID, "queued": len(queue), "kept": kept})
}

// CancelRemoval handles http and cancels the pending items of a removal job
func (g GroupHandler) CancelRemoval(c *gin.Context) {
	jobID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	canceled, err := CancelRemovalJob(c.Request.Context(), g.Pool, jobID)
	if err != nil {
		log.Printf("removal cancel: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"canceled": canceled})
}

// ListRemovals handles http and lists the removal jobs of a group
func (g GroupHandler) ListRemovals(c *gin.Context) {
	jobs, err := RemovalJobs(c.Request.Context(), g.Pool, c.Param("groupID"))
	if err != nil {
		log.Printf("removals list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, jobs)
}