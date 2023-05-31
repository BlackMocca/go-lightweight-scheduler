package repository

import (
	"context"
	"encoding/json"
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

func (p psqlRepository) setTrigger(ptr *models.Trigger) {
	ptr.ExecuteDatetime = constants.TIME_CHANGE(ptr.ExecuteDatetime, true)
	if ptr.ConfigString != "" {
		var m = map[string]interface{}{}
		if err := json.Unmarshal([]byte(ptr.ConfigString), &m); err == nil {
			ptr.Config = m
		}
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
			p.setTrigger(ptrs[index])
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
		trigger.ExecuteDatetime.Format(constants.TIME_FORMAT_RFC339),
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

func (p psqlRepository) UpsertJobTask(ctx context.Context, jobTask *models.JobTask) error {
	sql := `
	INSERT INTO "job_tasks" ("scheduler_name", "job_id", "task_status", "task_name", "task_type", "execution_name", "start_datetime", "end_datetime", "exception", "stacktrace", "created_at", "updated_at")
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT (job_id, task_name)
	DO UPDATE SET
		task_status=?,
		task_type=?,
		execution_name=?,
		start_datetime=?,
		end_datetime=?,
		exception=?,
		stacktrace=?,
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
		jobTask.SchedulerName,
		jobTask.JobId,
		jobTask.Status,
		jobTask.TaskName,
		jobTask.TaskType,
		jobTask.ExecutionName,
		jobTask.StartDateTime,
		jobTask.EndDatetime,
		jobTask.TaskException,
		jobTask.StackTrace,
		jobTask.CreatedAt,
		jobTask.UpdatedAt,
		/* update */
		jobTask.Status,
		jobTask.TaskType,
		jobTask.ExecutionName,
		jobTask.StartDateTime,
		jobTask.EndDatetime,
		jobTask.TaskException,
		jobTask.StackTrace,
		jobTask.CreatedAt,
		jobTask.UpdatedAt,
	)

	return err
}

func (p psqlRepository) GetOneTriggerByJobId(ctx context.Context, jobId string) (*models.Trigger, error) {
	var ptr = new(models.Trigger)
	sql := `
		SELECT 
			*
		FROM
			triggers
		WHERE
			job_id = ?
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, ptr, jobId); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	p.setTrigger(ptr)
	return ptr, nil
}

func (p psqlRepository) GetOneJob(ctx context.Context, jobId string) (*models.Job, error) {
	var ptr = new(models.Job)
	sql := `
		SELECT 
			*
		FROM
			jobs
		WHERE
			job_id = ?
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, ptr, jobId); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return ptr, nil
}

func (p psqlRepository) GetOneJobTaskByJobId(ctx context.Context, jobId string) ([]*models.JobTask, error) {
	var ptrs = []*models.JobTask{}
	sql := `
		SELECT 
			*
		FROM
			job_tasks
		WHERE
			job_id = ?
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &ptrs, jobId); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return ptrs, nil
}
