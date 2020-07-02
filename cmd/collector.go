package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/magefile/mage/sh"
	"github.com/spf13/cobra"
	"os"
)
type DiagInfo struct {
	HostName string
	DockerDaemon string
	ChronyDaemon string
	KubeletDameon string
	AzSecPack string
	HostInfo string
	InternetReachable bool
}

func CollectDiagnosticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "collect-diag",
		Short: "collects bunch of node diagnostic information",
		Long: "Use the collect command to gather bunch of node diagnostic information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Collecting node information")
			di := DiagInfo{
				HostName: getHostName(),
				DockerDaemon:      checkDaemonStatus("docker"),
				ChronyDaemon:      checkDaemonStatus("chrony"),
				KubeletDameon:     checkDaemonStatus("kubelet"),
				AzSecPack:         checkDaemonStatus("azsecpack"),
				HostInfo:          checkDaemonStatus("hostinfo"),
				InternetReachable: false,
			}

			b, err := json.Marshal(di);
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}

	return cmd
}


func checkDaemonStatus(name string) string {
	if err := sh.Run("systemctl", "is-active","--quiet", "service", name); err != nil {
		fmt.Println(err)
		return "NotRunning"
	}
	return "Running"
}

func getHostName() string {
	name, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return name
}