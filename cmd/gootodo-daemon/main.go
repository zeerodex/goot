package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/zeerodex/go-todo-tui/internal/database"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

type contextKey string

var tasksKey contextKey = "tasks"

type TaskProcessor struct {
	repo         tasks.TaskRepository
	batchSize    int
	timeWindow   time.Duration
	pollInterval time.Duration
}

func NewTaskProcessor(repo tasks.TaskRepository, batchSize int, timeWindow, pollInterval time.Duration) *TaskProcessor {
	return &TaskProcessor{repo: repo, batchSize: batchSize, timeWindow: timeWindow, pollInterval: pollInterval}
}

func (tp *TaskProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(tp.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Tick")
			ctx, err := tp.FetchTasks(ctx)
			fmt.Println(ctx.Value(tasksKey))
			if err != nil {
				log.Printf("error fetching tasks: %v", err)
			}

			tp.ProcessTasks(ctx)
		case <-ctx.Done():
			log.Println("Daemon shutting down...")
			return
		}
	}
}

func (tp *TaskProcessor) ProcessTasks(ctx context.Context) (context.Context, error) {
	tasks := ctx.Value(tasksKey).(tasks.Tasks)
	if len(tasks) < 1 {
		return nil, errors.New("no tasks")
	}
	for _, task := range tasks {
		go SendTaskDueNofitication(task)
	}
	return ctx, nil
}

func (tp *TaskProcessor) FetchTasks(ctx context.Context) (context.Context, error) {
	now := time.Now()
	minTime := now.Add(-tp.timeWindow)
	maxTime := now.Add(tp.timeWindow)

	tasks, err := tp.repo.GetByDueWithWindow(minTime, maxTime, tp.batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	ctx = context.WithValue(ctx, tasksKey, tasks)
	return ctx, nil
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

	tp := NewTaskProcessor(tasks.NewTaskRepository(db), 5, time.Minute, 5*time.Second)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tp.Start(ctx)

	<-sigs
	log.Println("Interrupt received, shutting down...")
	cancel()
}
