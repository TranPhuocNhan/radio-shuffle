package syncer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrJobNotFound    = errors.New("sync job not found")
	ErrJobDuplicate   = errors.New("sync job already completed")
	ErrJobInProgress  = errors.New("sync job already running")
	defaultScopeBytes = json.RawMessage(`{"mode":"full"}`)
)

type SyncRequest struct {
	RequestID   string          `json:"request_id"`
	RequestedBy string          `json:"requested_by"`
	Scope       json.RawMessage `json:"scope"`
	RequestedAt time.Time       `json:"requested_at"`
}

type JobPublisher interface {
	PublishSyncRequest(ctx context.Context, req SyncRequest) error
}

type JobService interface {
	Trigger(ctx context.Context, requestedBy string) (SyncJob, error)
	Status(ctx context.Context, requestID string) (SyncJob, error)
}

type JobProcessor interface {
	Process(ctx context.Context, req SyncRequest) (SyncResult, error)
}

type jobService struct {
	repo      Repository
	publisher JobPublisher
}

type jobProcessor struct {
	repo Repository
	sync Service
}

func NewJobService(repo Repository, publisher JobPublisher) JobService {
	return &jobService{repo: repo, publisher: publisher}
}

func NewJobProcessor(repo Repository, sync Service) JobProcessor {
	return &jobProcessor{repo: repo, sync: sync}
}

func (s *jobService) Trigger(ctx context.Context, requestedBy string) (SyncJob, error) {
	requestID, err := newRequestID()
	if err != nil {
		return SyncJob{}, err
	}
	req := SyncRequest{
		RequestID:   requestID,
		RequestedBy: requestedBy,
		Scope:       defaultScopeBytes,
		RequestedAt: time.Now(),
	}
	if err := s.repo.CreateSyncJob(ctx, CreateSyncJobInput{
		RequestID:   req.RequestID,
		Status:      SyncJobPending,
		RequestedBy: req.RequestedBy,
		Scope:       req.Scope,
	}); err != nil {
		return SyncJob{}, err
	}
	if err := s.publisher.PublishSyncRequest(ctx, req); err != nil {
		_ = s.repo.MarkSyncJobFailed(ctx, req.RequestID, err.Error())
		return SyncJob{}, err
	}
	return s.repo.GetSyncJob(ctx, req.RequestID)
}

func (s *jobService) Status(ctx context.Context, requestID string) (SyncJob, error) {
	job, err := s.repo.GetSyncJob(ctx, requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SyncJob{}, ErrJobNotFound
		}
		return SyncJob{}, err
	}
	return job, nil
}

func (p *jobProcessor) Process(ctx context.Context, req SyncRequest) (SyncResult, error) {
	job, err := p.repo.GetSyncJob(ctx, req.RequestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := p.repo.CreateSyncJob(ctx, CreateSyncJobInput{
				RequestID:   req.RequestID,
				Status:      SyncJobPending,
				RequestedBy: req.RequestedBy,
				Scope:       req.Scope,
			}); err != nil {
				return SyncResult{}, err
			}
			job, err = p.repo.GetSyncJob(ctx, req.RequestID)
			if err != nil {
				return SyncResult{}, err
			}
		} else {
			return SyncResult{}, err
		}
	}
	if job.RequestID != "" {
		switch job.Status {
		case SyncJobCompleted:
			return SyncResult{}, ErrJobDuplicate
		case SyncJobRunning:
			return SyncResult{}, ErrJobInProgress
		}
	}
	if err := p.repo.MarkSyncJobRunning(ctx, req.RequestID); err != nil {
		return SyncResult{}, err
	}
	result, err := p.sync.Sync(ctx)
	if err != nil {
		_ = p.repo.MarkSyncJobFailed(ctx, req.RequestID, err.Error())
		return result, err
	}
	if err := p.repo.MarkSyncJobCompleted(ctx, req.RequestID); err != nil {
		return result, err
	}
	return result, nil
}

func newRequestID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}



