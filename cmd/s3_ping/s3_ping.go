package s3_ping

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var (
	FlagEndpoint  string
	FlagAccessKey string
	FlagSecretKey string
	FlagBucket    string
	FlagRegion    string
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagEndpoint, "endpoint", "e", "", "S3 endpoint URL (e.g. https://minio.example.com)")
	Cmd.MarkFlagRequired("endpoint")
	Cmd.Flags().StringVarP(&FlagAccessKey, "access-key", "a", "", "S3 access key")
	Cmd.MarkFlagRequired("access-key")
	Cmd.Flags().StringVarP(&FlagSecretKey, "secret-key", "s", "", "S3 secret key")
	Cmd.MarkFlagRequired("secret-key")
	Cmd.Flags().StringVarP(&FlagBucket, "bucket", "b", "", "S3 bucket name")
	Cmd.MarkFlagRequired("bucket")
	Cmd.Flags().StringVarP(&FlagRegion, "region", "r", "us-east-1", "S3 region")
}

var Cmd = &cobra.Command{
	Use:   "s3-ping",
	Short: "Test S3 (MinIO, Silo, ...) by writing, reading and deleting a test object",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if err := s3Ping(FlagEndpoint, FlagAccessKey, FlagSecretKey, FlagBucket, FlagRegion); err != nil {
			log.Fatalln(err)
		}
	},
}

func s3Ping(endpoint, accessKey, secretKey, bucket, region string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := s3.New(s3.Options{
		Region:                     region,
		BaseEndpoint:               aws.String(endpoint),
		UsePathStyle:               true,
		Credentials:                credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})

	key := ".ping_" + randomHex(8) + ".txt"
	content := []byte("ping " + time.Now().UTC().Format(time.RFC3339Nano) + "\n")

	fmt.Printf("Endpoint: %s\nBucket:   %s\nObject:   %s\n\n", endpoint, bucket, key)

	// Write
	start := time.Now()
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(content),
		ContentLength: aws.Int64(int64(len(content))),
		ContentType:   aws.String("text/plain"),
	})
	if err != nil {
		return fmt.Errorf("WRITE failed: %w", err)
	}
	fmt.Printf("WRITE  OK  (%s)\n", time.Since(start).Round(time.Millisecond))

	// Read
	start = time.Now()
	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		cleanup(ctx, client, bucket, key)
		return fmt.Errorf("READ failed: %w", err)
	}
	data, err := io.ReadAll(out.Body)
	out.Body.Close()
	if err != nil {
		cleanup(ctx, client, bucket, key)
		return fmt.Errorf("READ failed: %w", err)
	}
	if !bytes.Equal(data, content) {
		cleanup(ctx, client, bucket, key)
		return fmt.Errorf("READ failed: content mismatch (expected %q, got %q)", content, data)
	}
	fmt.Printf("READ   OK  (%s)\n", time.Since(start).Round(time.Millisecond))

	// Delete
	start = time.Now()
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("DELETE failed: %w", err)
	}
	fmt.Printf("DELETE OK  (%s)\n", time.Since(start).Round(time.Millisecond))

	fmt.Println("\nS3 ping OK")
	return nil
}

func cleanup(ctx context.Context, client *s3.Client, bucket, key string) {
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Printf("cleanup of %s failed: %v\n", key, err)
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
