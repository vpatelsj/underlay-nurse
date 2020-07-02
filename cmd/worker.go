package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/spf13/cobra"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ProcessDiangnosticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "process-diag",
		Short: "process bunch of node diagnostic information",
		Long: "Use the process command to process bunch of node diagnostic information",
		RunE: runE,
	}
	return cmd
}

func runE(cmd *cobra.Command, args []string) error {
	log.Print("Starting worker task to process messages from queue")
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
	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping execution loop", "error", ctx.Err())
			return nil
		default:
			pullFromQueue()
		}
	}
}

func pullFromQueue() {
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

	msgUrl := queueUrl.NewMessagesURL()

	dequeueResp, err := msgUrl.Dequeue(ctx, 32, 10*time.Second)

	if err != nil {
		log.Fatal("Error dequeueing message: ", err)
	}

	for i := int32(0); i < dequeueResp.NumMessages(); i++ {
		msg := dequeueResp.Message(i)
		log.Printf("Deleting %v: {%v}", i, msg.Text)

		msgIdUrl := msgUrl.NewMessageIDURL(msg.ID)

		// PopReciept is required to delete the Message. If deletion fails using this popreceipt then the message has
		// been dequeued by another client.
		_, err = msgIdUrl.Delete(ctx, msg.PopReceipt)
		if err != nil {
			log.Fatal("Error deleting message: ", err)
		}
	}

}