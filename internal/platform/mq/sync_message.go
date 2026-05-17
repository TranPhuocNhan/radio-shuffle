package mq

import (
	"encoding/json"
	"time"
)

// SyncRequestedMessage is the MQ payload for sync requests.
type SyncRequestedMessage struct {
	RequestID   string          `json:"request_id"`
	RequestedBy string          `json:"requested_by"`
	Scope       json.RawMessage `json:"scope"`
	RequestedAt time.Time       `json:"requested_at"`
}
