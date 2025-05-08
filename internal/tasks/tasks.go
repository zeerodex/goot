package tasks

import "time"

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Due         time.Time `json:"due"`
	Completed   bool      `json:"status"`
}

type Tasks []Task
