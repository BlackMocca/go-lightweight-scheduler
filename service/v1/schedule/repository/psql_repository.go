package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/BlackMocca/sqlx"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
	"github.com/gofrs/uuid"
	"github.com/spf13/cast"
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

func (p psqlRepository) CreateTriggerByJobScheduler(ctx context.Context, trigger *models.Trigger) error {
	var tx, err = p.client.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`
			INSERT INTO "triggers" ("scheduler_name", "execute_datetime", "job_id", "config", "type", "is_trigger", "is_active", "created_at", "updated_at")
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (scheduler_name, execute_datetime, type) WHERE (type = '%s')
			DO NOTHING
			RETURNING job_id;
		`,
		string(constants.TRIGGER_TYPE_SCHEDULE),
	)
	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx,
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
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var jobId *uuid.UUID
	for rows.Next() {
		vals, err := rows.SliceScan()
		if err != nil {
			return err
		}
		if len(vals) > 0 {
			uid := uuid.FromStringOrNil(cast.ToString(vals[0]))
			jobId = &uid
		}
	}

	if jobId == nil || jobId.String() != trigger.JobId {
		return errors.New(constants.ERROR_ALREADY_EXISTS)
	}

	return tx.Commit()
}

func (p psqlRepository) ExecuteFutureJob(ctx context.Context, trigger *models.Trigger) (*models.Trigger, error) {
	var ptr = new(models.Trigger)
	var tx, err = p.client.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	sqlSelect := sqlx.Rebind(sqlx.DOLLAR, `
		SELECT * 
		FROM triggers
		WHERE job_id = ? 
		FOR UPDATE;
	`)
	queryStmt, err := tx.PreparexContext(ctx, sqlSelect)
	if err != nil {
		return nil, err
	}
	defer queryStmt.Close()
	if _, err := queryStmt.ExecContext(ctx, trigger.JobId); err != nil {
		return nil, err
	}

	sql := `
		UPDATE triggers 
		SET is_trigger = true 
		WHERE job_id = ? AND is_trigger = false AND is_active = true
		RETURNING triggers.* 
		`
	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, ptr,
		/* update */
		trigger.JobId,
	); err != nil && err.Error() != "sql: no rows in result set" {
		tx.Rollback()
		return nil, err
	}

	if ptr == nil || ptr.JobId == "" {
		tx.Rollback()
		return nil, nil
	}

	p.setTrigger(ptr)
	return ptr, tx.Commit()
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

func (p psqlRepository) filterJob(args *sync.Map) ([]string, []interface{}) {
	var conds = []string{}
	var vals = []interface{}{}

	if v, ok := args.Load("search_word"); ok {
		sql := "triggers.scheduler_name LIKE CONCAT('%',?::text,'%')"

		conds = append(conds, sql)
		vals = append(vals, v)
	}

	if v, ok := args.Load("job_id"); ok {
		sql := "triggers.job_id = ?"

		conds = append(conds, sql)
		vals = append(vals, v)
	}

	if v, ok := args.Load("status"); ok {
		sql := "jobs.status = ?"

		conds = append(conds, sql)
		vals = append(vals, strings.ToUpper(cast.ToString(v)))
	}

	var startDate, startDateOK = args.Load("start_date")
	var endDate, endDateOK = args.Load("end_date")
	if startDateOK && endDateOK {
		sql := "DATE(triggers.execute_datetime) >= ? AND DATE(triggers.execute_datetime) <= ?"

		conds = append(conds, sql)
		vals = append(vals, startDate, endDate)
	} else if startDateOK && !endDateOK {
		sql := "DATE(triggers.execute_datetime) >= ?"

		conds = append(conds, sql)
		vals = append(vals, startDate)
	} else if !startDateOK && endDateOK {
		sql := "DATE(triggers.execute_datetime) <= ?"

		conds = append(conds, sql)
		vals = append(vals, endDate)
	}

	return conds, vals
}

func (p psqlRepository) GetJobs(ctx context.Context, args *sync.Map, page int, perPage int) ([]*models.Job, int, error) {
	var ptrs = make([]*models.Job, 0)
	var totalRow int
	var where string
	var conds, vals = p.filterJob(args)
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	limit := fmt.Sprintf("LIMIT %d", perPage)
	offset := fmt.Sprintf("OFFSET %d", (page-1)*perPage)

	sql := fmt.Sprintf(`
		SELECT 
			triggers.scheduler_name,
			triggers.execute_datetime,
			triggers.job_id,
			triggers.config,
			jobs.status,
			jobs.start_datetime,
			jobs.end_datetime,
			COUNT(*) OVER() as total_row
		FROM
			triggers
		JOIN
			jobs
		ON
			triggers.job_id = jobs.job_id
		%s 	
		ORDER BY
			triggers.execute_datetime DESC
		%s 
		%s 
	`,
		where,
		limit,
		offset,
	)

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return nil, totalRow, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, vals...)
	if err != nil {
		return nil, totalRow, err
	}
	defer rows.Close()

	for rows.Next() {
		rv, err := rows.SliceScan()
		if err != nil {
			return nil, totalRow, err
		}
		if len(rv) > 0 {
			sn := cast.ToString(rv[0])
			exdt, _ := rv[1].(time.Time)
			jobId := cast.ToString(rv[2])
			configStr := cast.ToString(rv[3])
			jobstatus := constants.JobStatus(cast.ToString(rv[4]))
			st, _ := rv[5].(time.Time)
			en, _ := rv[6].(time.Time)
			totalRow = cast.ToInt(rv[7])
			job := &models.Job{
				SchedulerName: sn,
				JobId:         jobId,
				Status:        jobstatus,
				StartDateTime: &st,
				EndDatetime:   &en,
				Trigger: &models.Trigger{
					SchedulerName:   sn,
					ExecuteDatetime: exdt,
					JobId:           jobId,
					ConfigString:    configStr,
				},
			}
			p.setTrigger(job.Trigger)

			ptrs = append(ptrs, job)
		}
	}

	return ptrs, totalRow, nil
}

func (p psqlRepository) filterFutureJob(args *sync.Map) ([]string, []interface{}) {
	var conds = []string{}
	var vals = []interface{}{}

	if v, ok := args.Load("search_word"); ok {
		sql := "triggers.scheduler_name LIKE CONCAT('%',?::text,'%')"

		conds = append(conds, sql)
		vals = append(vals, v)
	}

	if v, ok := args.Load("job_id"); ok {
		sql := "triggers.job_id = ?"

		conds = append(conds, sql)
		vals = append(vals, v)
	}

	if v, ok := args.Load("job_id"); ok {
		sql := "triggers.job_id = ?"

		conds = append(conds, sql)
		vals = append(vals, v)
	}

	var startDate, startDateOK = args.Load("start_date")
	var endDate, endDateOK = args.Load("end_date")
	if startDateOK && endDateOK {
		sql := "DATE(triggers.execute_datetime) >= ? AND DATE(triggers.execute_datetime) <= ?"

		conds = append(conds, sql)
		vals = append(vals, startDate, endDate)
	} else if startDateOK && !endDateOK {
		sql := "DATE(triggers.execute_datetime) >= ?"

		conds = append(conds, sql)
		vals = append(vals, startDate)
	} else if !startDateOK && endDateOK {
		sql := "DATE(triggers.execute_datetime) <= ?"

		conds = append(conds, sql)
		vals = append(vals, endDate)
	}

	return conds, vals
}

func (p psqlRepository) GetFutureJob(ctx context.Context, args *sync.Map, page int, perPage int) ([]*models.Trigger, int, error) {
	var ptrs = make([]*models.Trigger, 0)
	var totalRow int
	var where string
	var conds, vals = p.filterFutureJob(args)
	if len(conds) > 0 {
		where = "AND " + strings.Join(conds, " AND ")
	}
	limit := fmt.Sprintf("LIMIT %d", perPage)
	offset := fmt.Sprintf("OFFSET %d", (page-1)*perPage)

	sql := fmt.Sprintf(`
		SELECT 
			triggers.scheduler_name,
			triggers.execute_datetime,
			triggers.job_id,
			triggers.config,
			triggers.type,
			triggers.is_trigger,
			triggers.is_active,
			COUNT(*) OVER() as total_row
		FROM
			triggers
		WHERE
			(triggers.execute_datetime >= NOW() AND triggers.is_trigger = false AND triggers.is_active = true)
			%s
		ORDER BY
			triggers.execute_datetime ASC
		%s 
		%s 
	`,
		where,
		limit,
		offset,
	)

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return nil, totalRow, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, vals...)
	if err != nil {
		return nil, totalRow, err
	}
	defer rows.Close()

	for rows.Next() {
		rv, err := rows.SliceScan()
		if err != nil {
			return nil, totalRow, err
		}
		if len(rv) > 0 {
			sn := cast.ToString(rv[0])
			exdt, _ := rv[1].(time.Time)
			jobId := cast.ToString(rv[2])
			configStr := cast.ToString(rv[3])
			triggerType := cast.ToString(rv[4])
			isTrigger := cast.ToBool(rv[5])
			isActive := cast.ToBool(rv[6])
			totalRow = cast.ToInt(rv[7])

			trigger := &models.Trigger{
				SchedulerName:   sn,
				JobId:           jobId,
				ExecuteDatetime: exdt,
				ConfigString:    configStr,
				TriggerType:     constants.TriggerType(triggerType),
				IsTrigger:       isTrigger,
				IsActive:        isActive,
			}
			p.setTrigger(trigger)

			ptrs = append(ptrs, trigger)
		}
	}

	return ptrs, totalRow, nil
}

func (p psqlRepository) filterJobTask(args *sync.Map) ([]string, []interface{}) {
	var conds = []string{}
	var vals = []interface{}{}

	if v, ok := args.Load("search_word"); ok {
		sql := "job_tasks.id::text LIKE CONCAT('%',?::text,'%') OR job_tasks.id::task_name LIKE CONCAT('%',?::text,'%')"

		conds = append(conds, sql)
		vals = append(vals, v, v)
	}

	if v, ok := args.Load("job_id"); ok {
		sql := "job_tasks.job_id = ?"

		conds = append(conds, sql)
		vals = append(vals, v)
	}

	if v, ok := args.Load("status"); ok {
		sql := "job_tasks.task_status = ?"

		conds = append(conds, sql)
		vals = append(vals, strings.ToUpper(cast.ToString(v)))
	}

	var startDate, startDateOK = args.Load("start_date")
	var endDate, endDateOK = args.Load("end_date")
	if startDateOK && endDateOK {
		sql := "DATE(job_tasks.start_datetime) >= ? AND DATE(job_tasks.start_datetime) <= ?"

		conds = append(conds, sql)
		vals = append(vals, startDate, endDate)
	} else if startDateOK && !endDateOK {
		sql := "DATE(job_tasks.start_datetime) >= ?"

		conds = append(conds, sql)
		vals = append(vals, startDate)
	} else if !startDateOK && endDateOK {
		sql := "DATE(job_tasks.start_datetime) <= ?"

		conds = append(conds, sql)
		vals = append(vals, endDate)
	}

	return conds, vals
}

func (p psqlRepository) GetJobTasks(ctx context.Context, args *sync.Map, page int, perPage int) ([]*models.JobTask, int, error) {
	var ptrs = make([]*models.JobTask, 0)
	var totalRow int
	var where string
	var conds, vals = p.filterJobTask(args)
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	limit := fmt.Sprintf("LIMIT %d", perPage)
	offset := fmt.Sprintf("OFFSET %d", (page-1)*perPage)

	sql := fmt.Sprintf(`
		SELECT 
			job_tasks.id,
			job_tasks.scheduler_name,
			job_tasks.job_id,
			job_tasks.task_status,
			job_tasks.task_name,
			job_tasks.task_type,
			job_tasks.execution_name,
			job_tasks.start_datetime,
			job_tasks.end_datetime,
			job_tasks.exception,
			COUNT(*) OVER() as total_row
		FROM
			job_tasks
		%s 	
		ORDER BY
			job_tasks.id DESC
		%s 
		%s 
	`,
		where,
		limit,
		offset,
	)

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return nil, totalRow, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, vals...)
	if err != nil {
		return nil, totalRow, err
	}
	defer rows.Close()

	for rows.Next() {
		rv, err := rows.SliceScan()
		if err != nil {
			return nil, totalRow, err
		}
		if len(rv) > 0 {
			id := cast.ToInt64(rv[0])
			schedulerName := cast.ToString(rv[1])
			jobId := cast.ToString(rv[2])
			status := constants.JobStatus(cast.ToString(rv[3]))
			taskName := cast.ToString(rv[4])
			taskType := cast.ToString(rv[5])
			executeName := cast.ToString(rv[6])
			var st time.Time
			var en *time.Time
			if rv[7] != nil {
				if v, ok := rv[7].(time.Time); ok {
					st = v
				}
			}
			if rv[8] != nil {
				if v, ok := rv[8].(time.Time); ok {
					en = &v
				}
			}
			exception := cast.ToString(rv[9])
			totalRow = cast.ToInt(rv[10])
			task := &models.JobTask{
				Id:            id,
				SchedulerName: schedulerName,
				JobId:         jobId,
				Status:        status,
				TaskName:      taskName,
				TaskType:      taskType,
				ExecutionName: executeName,
				StartDateTime: st,
				EndDatetime:   en,
				TaskException: exception,
			}

			ptrs = append(ptrs, task)
		}
	}
	return ptrs, totalRow, nil
}

func (p psqlRepository) UnActivatedTrigger(ctx context.Context, schedulerName string, configKey string, configValue interface{}) error {
	sql := fmt.Sprintf(`
		UPDATE triggers
		SET is_active=?, updated_at=?
		WHERE is_trigger=false AND scheduler_name=? AND config IS NOT NULL AND CAST(config AS jsonb)->>'%s' = ?;
	`,
		configKey,
	)

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, false, time.Now(), schedulerName, configValue)
	return err
}

func (p psqlRepository) UnActivatedTriggerByJobId(ctx context.Context, jobId *uuid.UUID) error {
	sql := `
		UPDATE triggers
		SET is_active=?, updated_at=?
		WHERE job_id = ?;
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, false, time.Now(), jobId)
	return err
}
