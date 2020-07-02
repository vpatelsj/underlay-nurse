package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/magefile/mage/sh"
	"github.com/spf13/cobra"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func collectDiagnosticsInfo() (string, error) {
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
		return "", err
	}

	return string(b), nil
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
			log.Printf("Collecting Diagnostics Logs and pushing to queue")
			pushToQueue()
			timer.Reset(5*time.Second)
		}
	}
}

func pushToQueue() {
	storageAccountName := os.Getenv("STORAGE_NAME")
	storageAccountKey  := os.Getenv("STORAGE_KEY")
	storageQueueName   := getHostName()

	_url, err := url.Parse(fmt.Sprintf("https://%s.queue.core.windows.net/%s", storageAccountName, storageQueueName))
	if err != nil {
		log.Fatal("Error parsing url: ", err)
	}

	credential, err := azqueue.NewSharedKeyCredential(storageAccountName, storageAccountKey)
	if err != nil {
		log.Fatal("Error creating credentials: ", err)
	}

	queueUrl := azqueue.NewQueueURL(*_url, azqueue.NewPipeline(credential, azqueue.PipelineOptions{}))

	ctx := context.TODO()

	props, err := queueUrl.GetProperties(ctx)
	if err != nil {
		// https://godoc.org/github.com/Azure/azure-storage-queue-go/azqueue#StorageErrorCodeType
		errorType := err.(azqueue.StorageError).ServiceCode()

		if (errorType == azqueue.ServiceCodeQueueNotFound) {

			log.Print("Queue does not exist, creating")

			_, err = queueUrl.Create(ctx, azqueue.Metadata{})
			if err != nil {
				log.Fatal("Error creating queue: ", err)
			}

			props, err = queueUrl.GetProperties(ctx)
			if err != nil {
				log.Fatal("Error parsing url: ", err)
			}

		} else {
			log.Fatal("Error getting queue properties: ", err)
		}
	}

	messageCount := props.ApproximateMessagesCount()
	log.Printf("Appx number of messages: %d", messageCount)


	newMessageContent, err := collectDiagnosticsInfo()
	if err != nil {
		log.Fatal(err)
	}

	msgUrl := queueUrl.NewMessagesURL()
	// (MessagesURL) Enqueue(context, messageText, visibilityTimeout, timeToLive) (*EnqueueMessageResponse, error)
	_, err = msgUrl.Enqueue(ctx, newMessageContent, 0, 0)
	if err != nil {
		log.Fatal("Error adding message to queue: ", err)
	}

	log.Printf("Added message \"%v\" to the queue", newMessageContent)
}