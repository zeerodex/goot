package models

import (
	"time"

	"github.com/zeerodex/goot/internal/tasks"
)

type Snapshot struct {
	API       string
	Timestamp time.Time
	Tasks     tasks.APITasks
}

type Snapshots []Snapshot
