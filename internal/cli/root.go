package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/apis/gtasksapi"
	gtaskscmds "github.com/zeerodex/goot/internal/cli/apis/gtasks_cmds"
	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/tui"
)

func newRootCmd(repo tasks.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "goot",
		Short: "Sleek cli/tui task manager with APIs integration",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := tea.NewProgram(tui.InitialMainModel(repo)).Run()
			cobra.CheckErr(err)
		},
	}
}

func NewTuiCmd(repo tasks.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "A brief description of your application",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := tea.NewProgram(tui.InitialMainModel(repo)).Run()
			cobra.CheckErr(err)
		},
	}
}

func Execute(repo tasks.TaskRepository, cfg *config.Config) {
	rootCmd := newRootCmd(repo)

	commands := []*cobra.Command{
		NewTuiCmd(repo),
		NewCreateCmd(repo),
		NewAllTasksCmd(repo),
		NewDeleteTaskCmd(repo),
		NewDoneTaskCmd(repo),
		NewDaemonCmd(repo),
	}

	rootCmd.AddCommand(commands...)

	for _, api := range cfg.APIs {
		switch api {
		case "google":
			api := gtasksapi.NewGTasksApi(cfg.Google.ListId)
			rootCmd.AddCommand(gtaskscmds.NewGoogleCmds(api))
		}
	}

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
