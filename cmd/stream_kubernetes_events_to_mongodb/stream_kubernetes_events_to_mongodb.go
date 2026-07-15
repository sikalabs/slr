package stream_kubernetes_events_to_mongodb

import (
	"context"
	"fmt"
	"os"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	FlagMongoURI      string
	FlagMongoDatabase string
	FlagCollection    string
	FlagNamespace     string
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagMongoURI, "mongo-uri", "u", getEnv("MONGO_URI", "mongodb://localhost:27017"), "MongoDB connection URI (include credentials, e.g. mongodb://user:pass@host:27017)")
	Cmd.Flags().StringVarP(&FlagMongoDatabase, "mongo-database", "d", getEnv("MONGO_DATABASE", "kubernetes"), "MongoDB database name")
	Cmd.Flags().StringVarP(&FlagCollection, "collection", "c", getEnv("MONGO_COLLECTION", "events"), "MongoDB collection name")
	Cmd.Flags().StringVarP(&FlagNamespace, "namespace", "n", getEnv("KUBERNETES_NAMESPACE", ""), "Namespace to watch events in (empty for all namespaces)")
}

var Cmd = &cobra.Command{
	Use:   "stream-kubernetes-events-to-mongodb",
	Short: "Watch Kubernetes events and stream them into a MongoDB collection",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if err := streamKubernetesEventsToMongodb(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func getKubernetesConfig() (*rest.Config, error) {
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}

func streamKubernetesEventsToMongodb() error {
	config, err := getKubernetesConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubernetes config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(FlagMongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to mongodb: %w", err)
	}
	defer mongoClient.Disconnect(ctx)
	collection := mongoClient.Database(FlagMongoDatabase).Collection(FlagCollection)

	watcher, err := clientset.CoreV1().Events(FlagNamespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to watch events: %w", err)
	}
	defer watcher.Stop()

	fmt.Printf("Streaming kubernetes events (namespace=%q) to mongodb %s/%s.%s\n", FlagNamespace, FlagMongoURI, FlagMongoDatabase, FlagCollection)

	for event := range watcher.ResultChan() {
		kubeEvent, ok := event.Object.(*corev1.Event)
		if !ok {
			continue
		}

		doc := bson.M{
			"id":        string(kubeEvent.ObjectMeta.UID),
			"timestamp": kubeEvent.ObjectMeta.CreationTimestamp.Unix(),
			"namespace": kubeEvent.ObjectMeta.Namespace,
			"kind":      kubeEvent.InvolvedObject.Kind,
			"name":      kubeEvent.InvolvedObject.Name,
			"reason":    kubeEvent.Reason,
			"type":      kubeEvent.Type,
			"message":   kubeEvent.Message,
		}

		_, err := collection.UpdateOne(
			ctx,
			bson.M{"id": doc["id"]},
			bson.M{"$set": doc},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			fmt.Printf("Error: failed to insert event %s: %v\n", doc["id"], err)
			continue
		}

		fmt.Printf(
			"id=%s timestamp=%d namespace=%s kind=%s name=%s message=%s\n",
			doc["id"], doc["timestamp"], doc["namespace"], doc["kind"], doc["name"], doc["message"],
		)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
