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

func (t Task) GTask() *gtasks.Task {
	var g gtasks.Task
	g.Title = t.Title
	g.Notes = t.Description
	g.Id = t.GoogleID
	if t.Completed {
		g.Status = "completed"
	} else if !t.Completed {
		g.Status = "needsAction"
	}
	g.Due = t.Due.Format(time.RFC3339)
	return &g
}

func (t Task) Task() string {
	if t.Description != "" {
		return fmt.Sprintf("ID:%d\n\tTitle: %s\n\tDescription:%s\n\tDue:%s\n\tCompleted:%t\n", t.ID, t.Title, t.Description, t.DueStr(), t.Completed)
	}
	return fmt.Sprintf("ID:%d\n\tTitle: %s\n\tDue:%s\n\tCompleted:%t\n", t.ID, t.Title, t.DueStr(), t.Completed)
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

type TasksList struct {
	ID       string
	GoogleID string
	Title    string
}

type TasksLists []TasksList
