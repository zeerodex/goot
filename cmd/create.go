package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

func NewCreateCmd(repo tasks.TaskRepository) *cobra.Command {
	var description string
	cmd := &cobra.Command{
		Use:   "create [title]",
		Short: "Create a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var task tasks.Task
			if args[0] != "" {
				task.Title = args[0]
			}
			if description != "" {
				task.Description = description
			}
			err := repo.Create(task.Title, task.Description)
			if err != nil {
				fmt.Printf("error creating task:%v", err)
				return
			}
			if task.Description != "" {
				fmt.Printf("Title: %s\nDescription:%s\n", task.Title, task.Description)
				return
			}
			fmt.Printf("Title: %s\n", task.Title)
			return
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the task")
	return cmd
}
