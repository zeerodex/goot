/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
)

// createCmd represents the create command

func NewCreateCmd(repo tasks.TaskRepository) *cobra.Command {
	var description string
	cmd := &cobra.Command{
		Use:   "create [title]",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Args: cobra.ExactArgs(1),
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
