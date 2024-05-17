package redis_set_large_data

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"log"

	"github.com/go-redis/redis/v8"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagRedisHost string
var FlagRedisPort int
var FlagRedisPassword string
var FlagRedisKey string
var FlagDataSize int

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagRedisHost, "redis-host", "", "localhost", "Redis Host")
	Cmd.Flags().IntVarP(&FlagRedisPort, "redis-port", "", 6379, "Redis Port")
	Cmd.Flags().StringVarP(&FlagRedisPassword, "redis-password", "", "", "Redis Password")
	Cmd.Flags().StringVarP(&FlagRedisKey, "redis-key", "", "large-data", "Redis Key")
	Cmd.Flags().IntVarP(&FlagDataSize, "data-size", "", 1, "Data size in MB")
}

var Cmd = &cobra.Command{
	Use:   "redis-set-large-data",
	Short: "Save large data to Redis",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		redisSetLargeData(FlagRedisHost, FlagRedisPort, FlagRedisPassword, FlagRedisKey, FlagDataSize)
	},
}

func redisSetLargeData(redisHost string, redisPort int, redisPassword string, redisKey string, dataSize int) {
	var ctx = context.Background()

	// Initialize the Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// Create a  dataSize MB string
	sizeInBytes := dataSize * 1024 * 1024
	randomData := generateRandomString(sizeInBytes)

	// Save the data to Redis
	err := rdb.Set(ctx, redisKey, randomData, 0).Err()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Data saved to Redis successfully")
}

// Function to generate a random string of a given size
func generateRandomString(size int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var sb strings.Builder
	for i := 0; i < size; i++ {
		sb.WriteRune(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}
