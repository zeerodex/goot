package main

import (
	"fmt"

	cmd "github.com/zeerodex/go-todo-tui/internal/cli"
	"github.com/zeerodex/go-todo-tui/internal/database"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		fmt.Println("failed to init db:", err)
		return
	}
	defer db.Close()

	repo := tasks.NewTaskRepository(db)
	cmd.Execute(repo)
}
