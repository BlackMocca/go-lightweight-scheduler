package schedule

import (
	"context"
	"sync"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
	"github.com/gofrs/uuid"
)

type Repository interface {
	GetOneTriggerByJobId(ctx context.Context, jobId string) (*models.Trigger, error)
	GetOneJob(ctx context.Context, jobId string) (*models.Job, error)
	GetOneJobTaskByJobId(ctx context.Context, jobId string) ([]*models.JobTask, error)
	GetTriggerTimer(ctx context.Context, schedulerName string) ([]*models.Trigger, error)
	GetJobs(ctx context.Context, args *sync.Map, page int, perPage int) ([]*models.Job, int, error)
	GetJobTasks(ctx context.Context, args *sync.Map, page int, perPage int) ([]*models.JobTask, int, error)
	GetFutureJob(ctx context.Context, args *sync.Map, page int, perPage int) ([]*models.Trigger, int, error)
	ExecuteFutureJob(ctx context.Context, trigger *models.Trigger) (*models.Trigger, error)
	UpsertTrigger(ctx context.Context, trigger *models.Trigger) error
	CreateTriggerByJobScheduler(ctx context.Context, trigger *models.Trigger) error
	UpsertJob(ctx context.Context, job *models.Job) error
	UpsertJobTask(ctx context.Context, jobTask *models.JobTask) error
	UnActivatedTrigger(ctx context.Context, schedulerName string, configKey string, configValue interface{}) error
	UnActivatedTriggerByJobId(ctx context.Context, jobId *uuid.UUID) error
}
