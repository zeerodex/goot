package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/tui"
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "goot",
		Short: "Sleek cli/tui task manager with APIs integration",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
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

func Execute(repo tasks.TaskRepository) {
	rootCmd := newRootCmd()

	rootCmd.AddCommand(NewTuiCmd(repo))

	rootCmd.AddCommand(NewCreateCmd(repo))
	rootCmd.AddCommand(NewAllTasksCmd(repo))
	rootCmd.AddCommand(NewDeleteTaskCmd(repo))
	rootCmd.AddCommand(NewDoneTaskCmd(repo))

	rootCmd.AddCommand(NewDaemonCmd(repo))

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
