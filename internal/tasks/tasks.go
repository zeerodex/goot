package tasks

import (
	"fmt"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Due         time.Time `json:"due"`
	Completed   bool      `json:"status"`
}

type Tasks []Task

func (t Task) Print() {
	if t.Description != "" {
		fmt.Printf("ID:%d\n\tTitle: %s\n\tDescription:%s\n\tDue:%s\n\tCompleted:%t\n", t.ID, t.Title, t.Description, t.DueStr(), t.Completed)
		return
	}
	fmt.Printf("ID:%d\n\tTitle: %s\n\tDue:%s\n\tCompleted:%t\n", t.ID, t.Title, t.DueStr(), t.Completed)
}

func (t *Task) DueStr() string {
	return t.Due.Local().Format("2006-01-02 15:04")
}

func (t *Task) SetDue(dueStr string) error {
	due, err := time.Parse(time.RFC3339, dueStr)
	if err != nil {
		return err
	}
	t.Due = due
	return nil
}
