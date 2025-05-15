package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/zeerodex/goot/internal/tasks"
)

type TaskProcessor struct {
	repo         tasks.TaskRepository
	timeWindow   time.Duration
	pollInterval time.Duration
}

func NewTaskProcessor(repo tasks.TaskRepository, timeWindow, pollInterval time.Duration) *TaskProcessor {
	return &TaskProcessor{repo: repo, timeWindow: timeWindow, pollInterval: pollInterval}
}

func (tp *TaskProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(tp.pollInterval)
	defer ticker.Stop()

	log.Printf("[INFO] Task processor started\npollInterval: %s\ntimeWindow:%s", tp.pollInterval, tp.timeWindow)

	for {
		select {
		case <-ticker.C:
			log.Println("[DEBUG] Tick")
			tasks, err := tp.FetchTasks()
			if err != nil {
				log.Printf("[ERROR] Failed to fetch tasks: %v", err)
			}
			if err = tp.ProcessTasks(tasks); err != nil {
				log.Printf("[ERROR] Failed to process tasks: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (tp *TaskProcessor) ProcessTasks(tasks tasks.Tasks) error {
	if len(tasks) < 1 {
		log.Println("[DEBUG] No pending tasks")
		return nil
	}
	now := time.Now()
	now = now.Truncate(time.Minute)
	for _, task := range tasks {
		timeDiff := now.Sub(task.Due)
		if timeDiff >= 0 && timeDiff <= time.Minute && !task.Notified {
			go tp.SendTaskDueNofitication(task)
			if err := tp.repo.MarkAsNotified(task.ID); err != nil {
				return fmt.Errorf("error marking task ID %d as notified: %w", task.ID, err)
			}
			log.Printf("[INFO] Task ID %d has been processed", task.ID)
		}
	}
	return nil
}

func (tp *TaskProcessor) FetchTasks() (tasks.Tasks, error) {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())

	minTime := now.Add(-tp.timeWindow)
	maxTime := now.Add(tp.timeWindow)

	tasks, err := tp.repo.GetPendingTasks(minTime, maxTime)
	if err != nil {
		return nil, fmt.Errorf("error fetching tasks: %w", err)
	}

	return tasks, nil
}

func (tp TaskProcessor) SendTaskDueNofitication(task tasks.Task) {
	icon := "task-due-symbolic"
	cmd := exec.Command("notify-send", "Task due!", task.Title, "-i", icon)
	err := cmd.Run()
	if err != nil {
		log.Printf("[ERROR] Failed to send notification for task ID %d: %v", task.ID, err)
		return
	}
	log.Printf("[INFO] Notification sent for task ID %d", task.ID)
}

func StartDaemon(repo tasks.TaskRepository) {
	tp := NewTaskProcessor(repo, time.Minute, time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	// TODO: SIGHUP
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go tp.Start(ctx)

	log.Printf("[INFO] Signal received: %s, shutting down...", <-sigs)
	cancel()
}
