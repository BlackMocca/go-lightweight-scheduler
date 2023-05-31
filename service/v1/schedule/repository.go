package schedule

import (
	"context"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
)

type Repository interface {
	GetOneTriggerByJobId(ctx context.Context, jobId string) (*models.Trigger, error)
	GetOneJob(ctx context.Context, jobId string) (*models.Job, error)
	GetOneJobTaskByJobId(ctx context.Context, jobId string) ([]*models.JobTask, error)
	GetTriggerTimer(ctx context.Context, schedulerName string) ([]*models.Trigger, error)
	UpsertTrigger(ctx context.Context, trigger *models.Trigger) error
	UpsertJob(ctx context.Context, job *models.Job) error
	UpsertJobTask(ctx context.Context, jobTask *models.JobTask) error
}
