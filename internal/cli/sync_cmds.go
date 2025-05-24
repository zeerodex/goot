package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeerodex/goot/internal/config"
	"github.com/zeerodex/goot/internal/services"
)

func NewSyncTasks(s services.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync tasks with apis",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Sync()
		},
	}
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
