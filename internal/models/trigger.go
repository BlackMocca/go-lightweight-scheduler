package models

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
)

type Trigger struct {
	TableName       struct{}               `json:"-" db:"triggers"`
	SchedulerName   string                 `json:"scheduler_name" db:"scheduler_name"`
	ExecuteDatetime time.Time              `json:"execute_datetime" db:"execute_datetime"`
	JobId           string                 `json:"job_id" db:"job_id"`
	ConfigString    string                 `json:"-" db:"config"`
	Config          map[string]interface{} `json:"config" db:"-"`
	TriggerType     constants.TriggerType  `json:"type" db:"type"`
	IsTrigger       bool                   `json:"is_trigger" db:"is_trigger"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

func (t *Trigger) GetConfigString() string {
	c := t.Config
	if t.Config == nil {
		c = make(map[string]interface{})
	}
	bu, _ := json.Marshal(c)
	return string(bu)
}

func (t *Trigger) GetConfigMutex() *sync.Map {
	sm := new(sync.Map)
	if t.Config != nil {
		for k, v := range t.Config {
			sm.Store(k, v)
		}
	}
	return sm
}
