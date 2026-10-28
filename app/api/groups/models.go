package groups

import "time"

type Group struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	Owner        string        `json:"owner,omitempty"`
	CreatedAt    int64         `json:"createdAt,omitempty"`
	Participants []Participant `json:"participants,omitempty"`
}

type Participant struct {
	ID           string `json:"id"`
	Number       string `json:"number"`
	IsAdmin      bool   `json:"isAdmin"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

type WebhookEvent struct {
	Event          string `json:"event"`
	IdempotencyKey string `json:"idempotencyKey"`
	Data           struct {
		GroupID        string   `json:"groupId"`
		ParticipantIDs []string `json:"participantIds"`
		Timestamp      int64    `json:"timestamp"`
		ActorID        string   `json:"actorId"`
	} `json:"data"`
}

type Member struct {
	ID       string    `json:"id"`
	Number   string    `json:"number"`
	PushName *string   `json:"pushName,omitempty"`
	Since    time.Time `json:"since"`
}

type Returned struct {
	ID       string    `json:"id"`
	Number   string    `json:"number"`
	PushName *string   `json:"pushName,omitempty"`
	Exits    int64     `json:"exits"`
	LastExit time.Time `json:"lastExit"`
	LastJoin time.Time `json:"lastJoin"`
}

type RemoveResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Status  int    `json:"status"`
}

type RemovalItem struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Error       *string    `json:"error,omitempty"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
}

type StoredGroup struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Managed  bool       `json:"managed"`
	SyncedAt *time.Time `json:"syncedAt,omitempty"`
}

type Contact struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	PushName string `json:"pushName"`
}

type RemovalJob struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	RequestedBy *string   `json:"requestedBy,omitempty"`
	Pending     int64     `json:"pending"`
	OK          int64     `json:"ok"`
	Failed      int64     `json:"failed"`
	Canceled    int64     `json:"canceled"`
}
