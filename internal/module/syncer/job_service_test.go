package syncer

import (
	"context"
	"errors"
	"testing"
)

type fakeJobRepo struct {
	activeJob SyncJob
	activeErr error
	created   []CreateSyncJobInput
	gotID     string
}

func (f *fakeJobRepo) UpsertBatch(_ context.Context, _ []UpsertInput) (int64, error) {
	return 0, errors.New("unexpected UpsertBatch")
}

func (f *fakeJobRepo) CreateSyncJob(_ context.Context, in CreateSyncJobInput) error {
	f.created = append(f.created, in)
	return nil
}

func (f *fakeJobRepo) GetActiveSyncJob(_ context.Context) (SyncJob, error) {
	if f.activeErr != nil {
		return SyncJob{}, f.activeErr
	}
	return f.activeJob, nil
}

func (f *fakeJobRepo) GetSyncJob(_ context.Context, requestID string) (SyncJob, error) {
	f.gotID = requestID
	return SyncJob{RequestID: requestID, Status: SyncJobPending}, nil
}

func (f *fakeJobRepo) MarkSyncJobRunning(_ context.Context, _ string) error {
	return errors.New("unexpected MarkSyncJobRunning")
}

func (f *fakeJobRepo) MarkSyncJobCompleted(_ context.Context, _ string) error {
	return errors.New("unexpected MarkSyncJobCompleted")
}

func (f *fakeJobRepo) MarkSyncJobFailed(_ context.Context, _ string, _ string) error {
	return errors.New("unexpected MarkSyncJobFailed")
}

type fakePublisher struct {
	published []SyncCommand
}

func (f *fakePublisher) PublishSyncCommand(_ context.Context, cmd SyncCommand) error {
	f.published = append(f.published, cmd)
	return nil
}

func TestJobService_TriggerRejectsActiveJob(t *testing.T) {
	repo := &fakeJobRepo{
		activeJob: SyncJob{RequestID: "active", Status: SyncJobRunning},
	}
	svc := NewJobService(repo, &fakePublisher{})

	_, err := svc.Trigger(context.Background(), "admin:1")
	if !errors.Is(err, ErrJobInProgress) {
		t.Fatalf("err = %v, want ErrJobInProgress", err)
	}
	if len(repo.created) != 0 {
		t.Fatalf("created %d jobs, want 0", len(repo.created))
	}
}

func TestJobService_TriggerCreatesJobWhenNoActiveJob(t *testing.T) {
	repo := &fakeJobRepo{activeErr: ErrRepoJobNotFound}
	pub := &fakePublisher{}
	svc := NewJobService(repo, pub)

	job, err := svc.Trigger(context.Background(), "admin:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.RequestID == "" {
		t.Fatal("missing request id")
	}
	if len(repo.created) != 1 {
		t.Fatalf("created %d jobs, want 1", len(repo.created))
	}
	if len(pub.published) != 1 {
		t.Fatalf("published %d commands, want 1", len(pub.published))
	}
	if repo.created[0].RequestID != pub.published[0].RequestID {
		t.Fatalf("created request id %q, published %q", repo.created[0].RequestID, pub.published[0].RequestID)
	}
}
