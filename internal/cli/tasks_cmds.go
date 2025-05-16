package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/tui"
	"github.com/zeerodex/goot/pkg/timeutil"
)

func NewAllTasksCmd(repo tasks.TaskRepository) *cobra.Command {
	var jsonFormat bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
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
					task.Task()
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
		Use:   "add [title] [date (Today if none)]",
		Short: "Creates a task",
		Args:  cobra.RangeArgs(1, 4),
		Run: func(cmd *cobra.Command, args []string) {
			var task tasks.Task
			// TODO: max length to cfg
			if len(args[0]) > 1024 {
				fmt.Println("Length of title is up to 1024 characters")
				return
			}
			task.Title = args[0]
			if len(description) > 8192 {
				fmt.Println("Length of description is up to 8192 characters")
				return

			}
			task.Description = description

			var dueStr string
			switch len(args) {
			case 4:
				dueStr = args[1] + " " + args[2] + " " + args[3]
			case 3:
				dueStr = args[1] + " " + args[2]
			case 2:
				dueStr = args[1]
			case 1:
				dueStr = "today"
			}

			if dueTimeStr != "" {
				dueStr += " " + dueTimeStr
			}

			due, err := timeutil.ParseAndValidateTimestamp(dueStr)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			task.Due = due

			err = repo.Create(task.Title, task.Description, task.Due)
			if err != nil {
				fmt.Printf("Error creating task:%v", err)
				return
			}
			// HACK:
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
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var id int
			if len(args) == 0 {
				tasks, err := repo.GetAll()
				if err != nil {
					fmt.Println(err)
					return
				}
				id = tui.ChooseList(tasks)
			} else {
				var err error
				id, err = strconv.Atoi(args[0])
				if err != nil {
					fmt.Printf("Incorrect task id: %v", err)
					return
				}
			}
			err := repo.DeleteByID(id)
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}

func NewDoneTaskCmd(repo tasks.TaskRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "done [task id]",
		Short: "Marks task completed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("Incorrect task id: %v", err)
				return
			}
			err = repo.ToggleCompleted(id, false)
			if err != nil {
				fmt.Printf("Failed to mark task completed: %v", err)
				return
			}
		},
	}
}
