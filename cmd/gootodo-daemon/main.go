package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/zeerodex/go-todo-tui/internal/database"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

func CheckTasks(repo tasks.TaskRepository) error {
	now := time.Now()
	due := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	task, err := repo.GetByDue(due)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	go SendTaskNofitication(task)
	return nil
}

func SendTaskNofitication(task tasks.Task) error {
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

	repo := tasks.NewTaskRepository(db)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Check")
			err = CheckTasks(repo)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
