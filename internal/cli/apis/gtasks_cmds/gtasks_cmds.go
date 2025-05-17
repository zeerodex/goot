package gtaskscmds

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/apis/gtasksapi"
	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/tasks"
	"github.com/zeerodex/goot/internal/tui"
)

func NewGoogleCmds(api *gtasksapi.GTasksApi) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gtasks",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Please specify what you want to do. Use --help for options")
		},
	}
	cmd.AddCommand(newGetGTaskListsCmd(api))
	cmd.AddCommand(newGetGTasksCmd(api))
	cmd.AddCommand(newDeleteGTaskCmd(api))

	cmd.AddCommand(newEnableSyncingCmd())
	return cmd
}

func newEnableSyncingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync (on/off)",
		Short: "Enables sync with google tasks api",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "on":
				config.SetGoogleSync(true)
			case "off":
				config.SetGoogleSync(false)
			default:
				fmt.Println("Valid options are on/off")
			}
		},
	}
}

func newGetGTaskListsCmd(api *gtasksapi.GTasksApi) *cobra.Command {
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

func newDeleteGTaskCmd(api *gtasksapi.GTasksApi) *cobra.Command {
	return &cobra.Command{
		Use:  "rm",
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var taskId string
			if len(args) != 1 {
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
					taskId = tui.ChooseGTask(tasks)
					if taskId == "" {
						return nil
					}
				} else {
					fmt.Println("There is no tasks in gtasks.")
					return nil
				}
			} else {
				taskId = args[0]
			}
			err := api.DeleteTaskById(taskId)
			if err != nil {
				return fmt.Errorf("unable to delete task: %v", err)
			}
			return nil
		},
	}
}

func newGetGTasksCmd(api *gtasksapi.GTasksApi) *cobra.Command {
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
