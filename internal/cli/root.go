package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/tui"
)

func newRootCmd(repo tasks.TaskRepository) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gootodo",
		Short: "A brief description of your application",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := tea.NewProgram(tui.InitialMainModel(repo)).Run(); err != nil {
				fmt.Println("Error running program:", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	return cmd
}

func Execute(repo tasks.TaskRepository) {
	rootCmd := newRootCmd(repo)

	rootCmd.AddCommand(NewCreateCmd(repo))
	rootCmd.AddCommand(NewAllTasksCmd(repo))
	rootCmd.AddCommand(NewDeleteTaskCmd(repo))

	rootCmd.AddCommand(NewDaemonCmd(repo))

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
