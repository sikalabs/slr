package list_mimio_s3_buckets

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var (
	FlagEndpoint  string
	FlagAccessKey string
	FlagSecretKey string
	FlagRegion    string
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagEndpoint, "endpoint", "e", getEnv("S3_ENDPOINT", ""), "S3 endpoint URL (required for MinIO)")
	Cmd.Flags().StringVarP(&FlagAccessKey, "access-key", "a", getEnv("AWS_ACCESS_KEY_ID", ""), "MinIO access key (required)")
	Cmd.Flags().StringVarP(&FlagSecretKey, "secret-key", "s", getEnv("AWS_SECRET_ACCESS_KEY", ""), "MinIO secret key (required)")
	Cmd.Flags().StringVarP(&FlagRegion, "region", "r", getEnv("AWS_REGION", "us-east-1"), "AWS region")
}

var Cmd = &cobra.Command{
	Use:   "list-mimio-s3-buckets",
	Short: "List all S3 buckets on MinIO",
	Long: `List all S3 buckets on MinIO server.

All credentials can be provided via flags or environment variables:
  --endpoint/-e or S3_ENDPOINT
  --access-key/-a or AWS_ACCESS_KEY_ID
  --secret-key/-s or AWS_SECRET_ACCESS_KEY
  --region/-r or AWS_REGION (default: us-east-1)`,
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if err := listBuckets(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func listBuckets() error {
	// Validate required parameters
	if FlagEndpoint == "" {
		return fmt.Errorf("endpoint is required (use --endpoint or S3_ENDPOINT env var)")
	}
	if FlagAccessKey == "" {
		return fmt.Errorf("access key is required (use --access-key or AWS_ACCESS_KEY_ID env var)")
	}
	if FlagSecretKey == "" {
		return fmt.Errorf("secret key is required (use --secret-key or AWS_SECRET_ACCESS_KEY env var)")
	}

	ctx := context.Background()

	// Create S3 client
	client, err := createS3Client(ctx)
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	// List buckets
	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to list buckets: %w", err)
	}

	fmt.Printf("Buckets on %s:\n\n", FlagEndpoint)
	for _, bucket := range result.Buckets {
		fmt.Printf("%s\n", aws.ToString(bucket.Name))
	}
	fmt.Printf("\nTotal buckets: %d\n", len(result.Buckets))

	return nil
}

func createS3Client(ctx context.Context) (*s3.Client, error) {
	// Use static credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(FlagRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			FlagAccessKey,
			FlagSecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	// Create S3 client with custom endpoint for MinIO
	clientOpts := []func(*s3.Options){
		func(o *s3.Options) {
			o.BaseEndpoint = aws.String(FlagEndpoint)
			o.UsePathStyle = true // MinIO requires path-style addressing
		},
	}

	return s3.NewFromConfig(cfg, clientOpts...), nil
}
