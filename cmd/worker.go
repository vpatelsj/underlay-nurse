package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func ProcessDiangnosticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "process-diag",
		Short: "process bunch of node diagnostic information",
		Long: "Use the process command to process bunch of node diagnostic information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Processing node information")
			return nil
		},
	}
	return cmd
}