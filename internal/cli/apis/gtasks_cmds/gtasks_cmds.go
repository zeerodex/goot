package gtaskscmds

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/apis/gtasksapi"
	"github.com/zeerodex/goot/internal/tasks"
)

func NewGoogleCmds(api *gtasksapi.GTasksApi) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gtasks",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Please specify what you want to do. Use --help for options")
		},
	}
	cmd.AddCommand(NewGetGTaskListsCmd(api))
	cmd.AddCommand(NewGetGTasksCmd(api))
	return cmd
}

func NewGetGTaskListsCmd(api *gtasksapi.GTasksApi) *cobra.Command {
	return &cobra.Command{
		Use:  "get",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := api.GetTasksLists()
			if err != nil {
				return fmt.Errorf("unable to get task lists: %v", err)
			}
			fmt.Println("Task Lists:")
			if len(res.Items) > 0 {
				for _, l := range res.Items {
					fmt.Printf("%s (%s)\n", l.Title, l.Id)
				}
			} else {
				fmt.Print("No task lists found.")
			}
			return nil
		},
	}
}

func NewDeleteGTasksCmd(api *gtasksapi.GTasksApi) *cobra.Command {
	return &cobra.Command{
		Use:  "list",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := api.GetTasks()
			if err != nil {
				return fmt.Errorf("unable to get tasks: %v", err)
			}
			if len(res.Items) > 0 {
				var tasks tasks.Tasks
				for _, t := range res.Items {
					task, err := gtasksapi.ConvertGTask(t)
					if err != nil {
						return fmt.Errorf("unable to get tasks: %v", err)
					}
					tasks = append(tasks, *task)
				}
			} else {
				fmt.Println("There is no tasks in gtasks.")
			}
			return nil
		},
	}
}

func NewGetGTasksCmd(api *gtasksapi.GTasksApi) *cobra.Command {
	return &cobra.Command{
		Use:  "list",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := api.GetTasks()
			if err != nil {
				return fmt.Errorf("unable to get tasks: %v", err)
			}
			fmt.Println("Tasks:")
			if len(res.Items) > 0 {
				for _, t := range res.Items {
					fmt.Printf("%s (%s)\n", t.Title, t.Id)
				}
			} else {
				fmt.Println("No tasks found.")
			}
			return nil
		},
	}
}
