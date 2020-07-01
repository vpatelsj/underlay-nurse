package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func CollectDiagnosticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "collect-diag",
		Short: "collects bunch of node diagnostic information",
		Long: "Use the collect command to gather bunch of node diagnostic information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Collecting node information")
			return nil
		},
	}

	return cmd
}