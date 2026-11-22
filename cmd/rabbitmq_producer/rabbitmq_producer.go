package rabbitmq_producer

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var (
	FlagURI      string
	FlagQueue    string
	FlagMessage  string
	FlagCount    int
	FlagInterval time.Duration
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "amqp://guest:guest@127.0.0.1:5672/", "RabbitMQ connection URI")
	Cmd.Flags().StringVarP(&FlagQueue, "queue", "q", "slr-test", "Queue name")
	Cmd.Flags().StringVarP(&FlagMessage, "message", "m", "hello", "Message to send (a counter and timestamp are appended)")
	Cmd.Flags().IntVarP(&FlagCount, "count", "c", 0, "Number of messages to send (0 = infinite)")
	Cmd.Flags().DurationVarP(&FlagInterval, "interval", "i", time.Second, "Interval between messages")
}

var Cmd = &cobra.Command{
	Use:     "rabbitmq-producer",
	Aliases: []string{"rmq-producer", "rmqp"},
	Short:   "Send test messages to a RabbitMQ queue",
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		rabbitmqProducer(FlagURI, FlagQueue, FlagMessage, FlagCount, FlagInterval)
	},
}

func rabbitmqProducer(uri, queue, message string, count int, interval time.Duration) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalln(err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalln(err)
	}

	for i := 1; count == 0 || i <= count; i++ {
		body := fmt.Sprintf("%s %d %s", message, i, time.Now().Format(time.RFC3339))
		err = ch.PublishWithContext(context.Background(), "", queue, false, false, amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(body),
		})
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("SENT queue=%s message=%s\n", queue, body)
		if count == 0 || i < count {
			time.Sleep(interval)
		}
	}
}
