package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/zeerodex/go-todo-tui/internal/database"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

type TaskProcessor struct {
	db           *sql.DB
	repo         tasks.TaskRepository
	batchSize    int
	timeWindow   time.Duration
	pollInterval time.Duration
}

func NewTaskProcessor(db *sql.DB, repo tasks.TaskRepository, batchSize int, timeWindow, pollInterval time.Duration) *TaskProcessor {
	return &TaskProcessor{db: db, repo: repo, batchSize: batchSize, timeWindow: timeWindow, pollInterval: pollInterval}
}

func (tp *TaskProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(tp.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			log.Println("Daemon shutting down")
			return
		}
	}
}

func (tp *TaskProcessor) CheckTasks(repo tasks.TaskRepository) error {
	now := time.Now()
	due := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	task, err := repo.GetByDue(due)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	go SendTaskDueNofitication(task)
	return nil
}

func SendTaskDueNofitication(task tasks.Task) error {
	icon := "task-due-symbolic"
	cmd := exec.Command("notify-send", task.Title, task.Description, "-i", icon)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tp := NewTaskProcessor(db, tasks.NewTaskRepository(db), 5, time.Minute, 5*time.Second)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tp.Start(ctx)

	<-sigs
	log.Println("Interrupt received, shutting down...")
	cancel()

	log.Println("Daemon shutdown completed")
}
