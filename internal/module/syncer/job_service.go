package syncer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrJobNotFound    = errors.New("sync job not found")
	ErrJobDuplicate   = errors.New("sync job already completed")
	ErrJobInProgress  = errors.New("sync job already running")
	defaultScopeBytes = json.RawMessage(`{"mode":"full"}`)
)

// SyncCommand is the domain command for triggering a sync job.
type SyncCommand struct {
	RequestID   string          `json:"request_id"`
	RequestedBy string          `json:"requested_by"`
	Scope       json.RawMessage `json:"scope"`
	RequestedAt time.Time       `json:"requested_at"`
}

type JobPublisher interface {
	PublishSyncCommand(ctx context.Context, cmd SyncCommand) error
}

type JobService interface {
	Trigger(ctx context.Context, requestedBy string) (SyncJob, error)
	Status(ctx context.Context, requestID string) (SyncJob, error)
}

type JobProcessor interface {
	Process(ctx context.Context, cmd SyncCommand) (SyncResult, error)
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
	cmd := SyncCommand{
		RequestID:   requestID,
		RequestedBy: requestedBy,
		Scope:       defaultScopeBytes,
		RequestedAt: time.Now(),
	}
	if err := s.repo.CreateSyncJob(ctx, CreateSyncJobInput{
		RequestID:   cmd.RequestID,
		Status:      SyncJobPending,
		RequestedBy: cmd.RequestedBy,
		Scope:       cmd.Scope,
	}); err != nil {
		return SyncJob{}, err
	}
	if err := s.publisher.PublishSyncCommand(ctx, cmd); err != nil {
		_ = s.repo.MarkSyncJobFailed(ctx, cmd.RequestID, err.Error())
		return SyncJob{}, err
	}
	return s.repo.GetSyncJob(ctx, cmd.RequestID)
}

func (s *jobService) Status(ctx context.Context, requestID string) (SyncJob, error) {
	job, err := s.repo.GetSyncJob(ctx, requestID)
	if err != nil {
		if errors.Is(err, ErrRepoJobNotFound) {
			return SyncJob{}, ErrJobNotFound
		}
		return SyncJob{}, err
	}
	return job, nil
}

func (p *jobProcessor) Process(ctx context.Context, cmd SyncCommand) (SyncResult, error) {
	job, err := p.repo.GetSyncJob(ctx, cmd.RequestID)
	if err != nil {
		if errors.Is(err, ErrRepoJobNotFound) {
			if err := p.repo.CreateSyncJob(ctx, CreateSyncJobInput{
				RequestID:   cmd.RequestID,
				Status:      SyncJobPending,
				RequestedBy: cmd.RequestedBy,
				Scope:       cmd.Scope,
			}); err != nil {
				return SyncResult{}, err
			}
			job, err = p.repo.GetSyncJob(ctx, cmd.RequestID)
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
	if err := p.repo.MarkSyncJobRunning(ctx, cmd.RequestID); err != nil {
		return SyncResult{}, err
	}
	result, err := p.sync.Sync(ctx)
	if err != nil {
		_ = p.repo.MarkSyncJobFailed(ctx, cmd.RequestID, err.Error())
		return result, err
	}
	if err := p.repo.MarkSyncJobCompleted(ctx, cmd.RequestID); err != nil {
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
