package rabbitmq_consumer

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var (
	FlagURI   string
	FlagQueue string
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagURI, "uri", "amqp://guest:guest@127.0.0.1:5672/", "RabbitMQ connection URI")
	Cmd.Flags().StringVarP(&FlagQueue, "queue", "q", "slr-test", "Queue name")
}

var Cmd = &cobra.Command{
	Use:     "rabbitmq-consumer",
	Aliases: []string{"rmq-consumer", "rmqc"},
	Short:   "Consume and print messages from a RabbitMQ queue",
	Args:    cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		rabbitmqConsumer(FlagURI, FlagQueue)
	},
}

func rabbitmqConsumer(uri, queue string) {
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

	msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Waiting for messages on queue=%s\n", queue)
	for msg := range msgs {
		log.Printf("RECEIVED queue=%s message=%s\n", queue, string(msg.Body))
	}
	log.Fatalln("Connection closed")
}
