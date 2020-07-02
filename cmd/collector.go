package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/magefile/mage/sh"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log"
)
type DiagInfo struct {
	HostName string
	DockerDaemon string
	ChronyDaemon string
	KubeletDameon string
	InternetReachable string
}

func CollectDiagnosticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "collect-diag",
		Short: "collects bunch of node diagnostic information",
		Long: "Use the collect command to gather bunch of node diagnostic information",
		RunE: run,
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

func collectDiagnosticsInfo() error {
	fmt.Println("Collecting node information")
	di := DiagInfo{
		HostName: getHostName(),
		DockerDaemon:      checkDaemonStatus("docker"),
		ChronyDaemon:      checkDaemonStatus("chrony"),
		KubeletDameon:     checkDaemonStatus("kubelet"),
		InternetReachable: "Not checked",
	}

	b, err := json.Marshal(di);
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func run(cmd *cobra.Command, args []string) error {

	ctx := context.Background()
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)

	defer cancel()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sc

		log.Printf("Received signal, exting now", "signal", sig.Signal)
		cancel()
	}()
	timer := time.NewTimer(5*time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping execution loop", "error", ctx.Err())
			timer.Stop()
			return nil
		case <-timer.C:
			log.Printf("Collecting Diagnostics Logs")
			collectDiagnosticsInfo()
			log.Printf("Pushing Diagnostics Logs")
			//TODO: Push to queue
			timer.Reset(5*time.Second)
		}
	}
}