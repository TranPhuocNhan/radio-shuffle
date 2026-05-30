package syncer

import (
	"encoding/json"
	"time"
)

type TriggerSyncResponse struct {
	RequestID string `json:"request_id"`
	StatusURL string `json:"status_url"`
}

type SyncStatusResponse struct {
	RequestID   string          `json:"request_id"`
	Status      string          `json:"status"`
	RequestedBy string          `json:"requested_by"`
	Scope       json.RawMessage `json:"scope"`
	Error       *string         `json:"error,omitempty"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	FinishedAt  *time.Time      `json:"finished_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

func toSyncStatusResponse(job SyncJob) SyncStatusResponse {
	return SyncStatusResponse{
		RequestID:   job.RequestID,
		Status:      string(job.Status),
		RequestedBy: job.RequestedBy,
		Scope:       job.Scope,
		Error:       job.ErrorMessage,
		StartedAt:   job.StartedAt,
		FinishedAt:  job.FinishedAt,
		CreatedAt:   job.CreatedAt,
	}
}

