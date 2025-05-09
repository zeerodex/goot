package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
	"github.com/zeerodex/go-todo-tui/pkg/timeutil"
)

func NewAllTasksCmd(repo tasks.TaskRepository) *cobra.Command {
	var jsonFormat bool
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Get all tasks",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			tasks, err := repo.GetAll()
			if err != nil {
				fmt.Println(err)
				return
			}
			if jsonFormat {
				b, err := json.MarshalIndent(&tasks, "", " ")
				if err != nil {
					fmt.Println(err)
				}
				os.Stdout.Write(b)
				fmt.Println()
			} else {
				for _, task := range tasks {
					task.Print()
				}
			}
		},
	}
	cmd.Flags().BoolVarP(&jsonFormat, "json", "j", false, "Output in json format")
	return cmd
}

func NewCreateCmd(repo tasks.TaskRepository) *cobra.Command {
	var description string
	var dueTimeStr string
	cmd := &cobra.Command{
		Use:   "create [title] [YYYY-MM-DD]",
		Short: "Create a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var task tasks.Task
			if args[0] != "" {
				task.Title = args[0]
			}
			if description != "" {
				task.Description = description
			}

			dueStr := args[1]
			if dueTimeStr != "" {
				dueStr += " " + dueTimeStr
			}
			due, err := timeutil.ParseAndValidateTimestamp(dueStr)
			if err != nil {
				fmt.Println("Invalid format")
				return
			}
			task.Due = due

			err = repo.Create(task.Title, task.Description, task.Due)
			if err != nil {
				fmt.Printf("error creating task:%v", err)
				return
			}

			task.Print()
		},
	}
	cmd.Flags().StringVarP(&dueTimeStr, "time", "t", "", "Due time (HH:MM)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the task")
	return cmd
}

func NewDeleteTaskCmd(repo tasks.TaskRepository) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [task id]",
		Short: "Deletes a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("incorrect id:%v", err)
				return
			}
			err = repo.DeleteByID(id)
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
