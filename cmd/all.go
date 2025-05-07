package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zeerodex/go-todo-tui/internal/tasks"
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
					fmt.Printf("Id:%d\n\tTitle:%s\n\tDescription:%s\n\tStatus:%t\n", task.ID, task.Title, task.Description, task.Status)
				}
			}
		},
	}
	cmd.Flags().BoolVarP(&jsonFormat, "json", "j", false, "Output in json format")
	return cmd
}
