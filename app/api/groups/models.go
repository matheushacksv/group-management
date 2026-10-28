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
		// group.join / group.leave
		GroupID        string   `json:"groupId"`
		ParticipantIDs []string `json:"participantIds"`
		Timestamp      int64    `json:"timestamp"`
		ActorID        string   `json:"actorId"`

		// message.received: only who sent it is used, the message content is never read
		ChatID  string `json:"chatId"`
		Author  string `json:"author"`
		IsGroup bool   `json:"isGroup"`
		FromMe  bool   `json:"fromMe"`
		Contact struct {
			Name     string `json:"name"`
			PushName string `json:"pushName"`
		} `json:"contact"`
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
	ClosedAt *time.Time `json:"closedAt,omitempty"`
}

type ActiveRemoval struct {
	ID         int64  `json:"id"`
	GroupID    string `json:"groupId"`
	CloseGroup bool   `json:"closeGroup"`
	Pending    int64  `json:"pending"`
}

type Closing struct {
	SessionID string
	GroupID   string
}

type Person struct {
	ID       string        `json:"id"`
	Number   string        `json:"number"`
	PushName *string       `json:"pushName,omitempty"`
	Groups   []PersonGroup `json:"groups"`
}

type PersonGroup struct {
	GroupID string    `json:"groupId"`
	Name    string    `json:"name"`
	Present bool      `json:"present"`
	Since   time.Time `json:"since"`
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
