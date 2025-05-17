package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/apis/gtasksapi"
)

func NewGetGTasksCmd(api *gtasksapi.GTasksApi) *cobra.Command {
	return &cobra.Command{
		Use:  "gtasks",
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := api.GetTasksFromDefault()
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, task := range res.Items {
				fmt.Println(task.Due)
			}
		},
	}
}
