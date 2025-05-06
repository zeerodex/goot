package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zeerodex/go-todo-tui/internal/database"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
	"github.com/zeerodex/go-todo-tui/internal/tui"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		fmt.Println("error init db:", err)
		return
	}
	defer db.Close()

	repo := tasks.NewTaskRepository(db)

	if _, err := tea.NewProgram(tui.InitialMainModel(repo)).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
