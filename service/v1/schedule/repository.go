package schedule

import (
	"context"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
)

type Repository interface {
	GetTriggerTimer(ctx context.Context, schedulerName string) ([]*models.Trigger, error)
	UpsertTrigger(ctx context.Context, trigger *models.Trigger) error
}
