package cli

import (
	"github.com/spf13/cobra"

	"github.com/zeerodex/goot/internal/apis/gtasksapi"
)

func NewGetGTasksCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "gtasks",
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			srv := gtasksapi.GetService()
			gtasksapi.GetTasksInDefault(srv)
		},
	}
}
