package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/services"
	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/tui"
	"github.com/zeerodex/goot/pkg/timeutil"
)

func NewAllTasksCmd(s services.TaskService) *cobra.Command {
	var jsonFormat bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			tasks, err := s.GetAllTasks()
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
					fmt.Println(task.Task())
				}
			}
		},
	}
	cmd.Flags().BoolVarP(&jsonFormat, "json", "j", false, "Output in json format")
	return cmd
}

func NewCreateCmd(s services.TaskService) *cobra.Command {
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

			_, err = s.CreateTask(&task)
			if err != nil {
				fmt.Printf("Error creating task: %v", err)
				return
			}
			fmt.Println(task.Task())
		},
	}
	cmd.Flags().StringVarP(&dueTimeStr, "time", "t", "", "Due time (HH:MM)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the task")
	return cmd
}

func NewDeleteTaskCmd(s services.TaskService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [task id]",
		Short: "Deletes a task",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var id int
			if len(args) == 0 {
				tasks, err := s.GetAllTasks()
				if err != nil {
					fmt.Println(err)
					return
				}
				var ok bool
				id, ok = tui.ChooseTask(tasks)
				if !ok {
					fmt.Println("No tasks specified")
					return
				}
			} else {
				var err error
				id, err = strconv.Atoi(args[0])
				if err != nil {
					fmt.Printf("Incorrect task id: %v", err)
					return
				}
			}
			err := s.DeleteTaskByID(id)
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}
	return cmd
}

func NewDoneTaskCmd(s services.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "done [task id]",
		Short: "Marks task completed",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id int
			if len(args) == 0 {
				tasks, err := s.GetAllTasks()
				if err != nil {
					return err
				}
				var ok bool
				id, ok = tui.ChooseTask(tasks)
				if !ok {
					fmt.Println("No tasks specified")
					return nil
				}
			} else {
				var err error
				id, err = strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("incorrect task id: %w", err)
				}
			}
			err := s.ToggleCompleted(id, true)
			if err != nil {
				return fmt.Errorf("failed to mark task completed: %w", err)
			}
			return nil
		},
	}
}
