package cli

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/services"
	"github.com/zeerodex/goot/internal/tui"
)

func newRootCmd(s services.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "goot",
		Short: "Sleek cli/tui task manager with APIs integration",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := tea.NewProgram(tui.InitialMainModel(s)).Run()
			cobra.CheckErr(err)
		},
	}
}

func Execute(s services.TaskService, cfg *config.Config) {
	rootCmd := newRootCmd(s)

	commands := []*cobra.Command{
		NewCreateCmd(s),
		NewAllTasksCmd(s),
		NewDeleteTaskCmd(s),
		NewDoneTaskCmd(s),

		NewDaemonCmd(s),

		NewSyncCmd(s, cfg.APIs),
	}
	rootCmd.AddCommand(commands...)

	if cfg.SyncOnStartup {
		err := s.Sync()
		if err != nil {
			log.Printf("Failed to sync tasks on startup: %v", err)
		}
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
