package repository

import (
	"context"
	"fmt"

	"github.com/BlackMocca/sqlx"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
)

type psqlRepository struct {
	client *sqlx.DB
}

func NewPsqlRepository(client *sqlx.DB) schedule.Repository {
	return &psqlRepository{
		client: client,
	}
}

func (p psqlRepository) GetTriggerTimer(ctx context.Context, schedulerName string) ([]*models.Trigger, error) {
	var ptrs = []*models.Trigger{}
	sql := fmt.Sprintf(`
		SELECT 
			*
		FROM
			triggers
		WHERE 
			scheduler_name = ? AND "type"::text = ? AND is_trigger = false AND is_active = true
	`,
	)

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &ptrs, schedulerName, constants.TRIGGER_TYPE_EXTERNAL); err != nil {
		return nil, err
	}
	if len(ptrs) > 0 {
		for index, _ := range ptrs {
			ptrs[index].ExecuteDatetime = constants.TIME_CHANGE(ptrs[index].ExecuteDatetime, true)
		}
	}

	return ptrs, nil
}

func (p psqlRepository) UpsertTrigger(ctx context.Context, trigger *models.Trigger) error {
	sql := `
		INSERT INTO "triggers" ("scheduler_name", "execute_datetime", "job_id", "config", "type", "is_trigger", "is_active", "created_at", "updated_at")
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (job_id)
		DO UPDATE SET
			is_trigger=?,
			is_active=?,
			updated_at=?
	`
	sql = sqlx.Rebind(sqlx.DOLLAR, sql)

	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		/* insert */
		trigger.SchedulerName,
		trigger.ExecuteDatetime,
		trigger.JobId,
		trigger.GetConfigString(),
		trigger.TriggerType,
		trigger.IsTrigger,
		trigger.IsActive,
		trigger.CreatedAt,
		trigger.UpdatedAt,
		/* update */
		trigger.IsTrigger,
		trigger.IsActive,
		trigger.UpdatedAt,
	)

	return err
}

func (p psqlRepository) UpsertJob(ctx context.Context, job *models.Job) error {
	sql := `
		INSERT INTO "jobs" ("scheduler_name", "job_id", "status", "start_datetime", "end_datetime", "created_at", "updated_at")
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (scheduler_name, job_id)
		DO UPDATE SET
			status=?,
			start_datetime=?,
			end_datetime=?,
			created_at=?,
			updated_at=?
	`
	sql = sqlx.Rebind(sqlx.DOLLAR, sql)

	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		/* insert */
		job.SchedulerName,
		job.JobId,
		string(job.Status),
		job.StartDateTime,
		job.EndDatetime,
		job.CreatedAt,
		job.UpdatedAt,
		/* update */
		job.Status,
		job.StartDateTime,
		job.EndDatetime,
		job.CreatedAt,
		job.UpdatedAt,
	)

	return err
}
