package main

import (
	"log"

	cmd "github.com/zeerodex/goot/internal/cli"
	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/database"
	"github.com/zeerodex/goot/internal/repositories"
	"github.com/zeerodex/goot/internal/services"
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

	service, err := services.NewTaskService(repositories.NewTaskRepository(db), cfg)
	if err != nil {
		log.Fatalf("Unable to initialize service: %v", err)
	}

	cmd.Execute(service, cfg)
}
