package cli

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	gtaskscmds "github.com/zeerodex/goot/internal/cli/apis/gtasks_cmds"
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
		NewSyncTasks(s),

		NewDaemonCmd(s),
	}
	rootCmd.AddCommand(commands...)

	rootCmd.AddCommand(gtaskscmds.NewGoogleCmds(s.GetGApi()))

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
