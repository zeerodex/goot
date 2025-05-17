package tasks

import (
	"fmt"
	"time"

	gtasks "google.golang.org/api/tasks/v1"
)

type Task struct {
	ID          int `json:"id"`
	GoogleID    string
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Due         time.Time `json:"due"`
	Completed   bool      `json:"status"`
	Notified    bool      `json:"notified"`
}

type Tasks []Task

func ParseFromGtasks(g *gtasks.Task) Task {
	var t Task
	t.Title = g.Title
	if g.Notes != "" {
		t.Description = g.Notes
	}
	if g.Due != "" {
		t.Due, _ = time.Parse(time.RFC3339, g.Due)
	}
	if g.Status == "completed" {
		t.Completed = true
	} else {
		t.Completed = false
	}
	return t
}

func (t Task) Task() {
	if t.Description != "" {
		fmt.Printf("ID:%d\n\tTitle: %s\n\tDescription:%s\n\tDue:%s\n\tCompleted:%t\n", t.ID, t.Title, t.Description, t.DueStr(), t.Completed)
		return
	}
	fmt.Printf("ID:%d\n\tTitle: %s\n\tDue:%s\n\tCompleted:%t\n", t.ID, t.Title, t.DueStr(), t.Completed)
}

func (t *Task) DueStr() string {
	if t.Due.Hour() == 0 && t.Due.Minute() == 0 {
		return t.Due.Format("2006-01-02")
	}
	return t.Due.Format("2006-01-02 15:04")
}

func (t *Task) SetDue(dueStr string) error {
	due, err := time.Parse(time.RFC3339, dueStr)
	if err != nil {
		return err
	}
	t.Due = due
	return nil
}

func (t Task) FullTitle() string {
	var title string
	title += t.Title
	if !t.Due.IsZero() {
		title += " | " + t.DueStr()
	}
	if t.Completed {
		title += " | Completed"
	} else {
		title += " | Uncompleted"
	}
	return title
}
