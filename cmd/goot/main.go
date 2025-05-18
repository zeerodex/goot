package main

import (
	"log"

	"github.com/zeerodex/goot/internal/apis/gtasksapi"
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

	gApi := gtasksapi.NewGTasksApi(cfg.Google.ListId)
	service := services.NewTaskService(repositories.NewTaskRepository(db), gApi, cfg.Google.Sync)

	cmd.Execute(service, cfg)
}
