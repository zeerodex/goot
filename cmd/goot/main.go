package main

import (
	"log"

	cmd "github.com/zeerodex/goot/internal/cli"
	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/database"
	"github.com/zeerodex/goot/internal/tasks"
)

func main() {
	cfg, err := config.LoadConfig("internal/config")
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)
	}

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Unable to init database: %v", err)
	}
	defer db.Close()

	repo := tasks.NewTaskRepository(db)

	cmd.Execute(repo, cfg)
}
