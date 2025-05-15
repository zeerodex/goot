package main

import (
	"fmt"

	cmd "github.com/zeerodex/goot/internal/cli"
	"github.com/zeerodex/goot/internal/database"
	"github.com/zeerodex/goot/internal/tasks"
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
