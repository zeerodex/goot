package models

import (
	"time"
)

type Snapshot struct {
	API       string
	Timestamp time.Time
	IDs       []string
}

type Snapshots []Snapshot
