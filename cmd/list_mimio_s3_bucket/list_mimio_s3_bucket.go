package list_mimio_s3_bucket

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
	FlagBucket    string
	FlagEndpoint  string
	FlagAccessKey string
	FlagSecretKey string
	FlagRegion    string
	FlagPrefix    string
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagBucket, "bucket", "b", getEnv("S3_BUCKET", ""), "S3 bucket name (required)")
	Cmd.Flags().StringVarP(&FlagPrefix, "prefix", "p", getEnv("S3_PREFIX", ""), "Prefix/path in bucket to list")
	Cmd.Flags().StringVarP(&FlagEndpoint, "endpoint", "e", getEnv("S3_ENDPOINT", ""), "S3 endpoint URL (required for MinIO)")
	Cmd.Flags().StringVarP(&FlagAccessKey, "access-key", "a", getEnv("AWS_ACCESS_KEY_ID", ""), "MinIO access key (required)")
	Cmd.Flags().StringVarP(&FlagSecretKey, "secret-key", "s", getEnv("AWS_SECRET_ACCESS_KEY", ""), "MinIO secret key (required)")
	Cmd.Flags().StringVarP(&FlagRegion, "region", "r", getEnv("AWS_REGION", "us-east-1"), "AWS region")
}

var Cmd = &cobra.Command{
	Use:   "list-mimio-s3-bucket",
	Short: "List objects in MinIO S3 bucket",
	Long: `List all objects in a MinIO S3 bucket with their details.

All credentials can be provided via flags or environment variables:
  --bucket/-b or S3_BUCKET
  --endpoint/-e or S3_ENDPOINT
  --access-key/-a or AWS_ACCESS_KEY_ID
  --secret-key/-s or AWS_SECRET_ACCESS_KEY
  --region/-r or AWS_REGION (default: us-east-1)
  --prefix/-p or S3_PREFIX (optional)`,
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if err := listBucket(); err != nil {
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

func listBucket() error {
	// Validate required parameters
	if FlagBucket == "" {
		return fmt.Errorf("bucket name is required (use --bucket or S3_BUCKET env var)")
	}
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

	// List objects
	return listObjects(ctx, client)
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

func listObjects(ctx context.Context, client *s3.Client) error {
	prefixMsg := ""
	if FlagPrefix != "" {
		prefixMsg = fmt.Sprintf(" with prefix '%s'", FlagPrefix)
	}
	fmt.Printf("Listing objects in bucket '%s'%s...\n\n", FlagBucket, prefixMsg)

	var objectCount int64
	var totalSize int64

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(FlagBucket),
		Prefix: aws.String(FlagPrefix),
	})

	fmt.Printf("%-60s %20s %30s\n", "Key", "Size", "Last Modified")
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------------------------")

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			size := aws.ToInt64(obj.Size)
			lastModified := ""
			if obj.LastModified != nil {
				lastModified = obj.LastModified.Format("2006-01-02 15:04:05 MST")
			}

			fmt.Printf("%-60s %20s %30s\n", key, formatBytes(size), lastModified)
			objectCount++
			totalSize += size
		}
	}

	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("\nTotal objects: %s\n", formatNumber(objectCount))
	fmt.Printf("Total size: %s\n", formatBytes(totalSize))

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Add thousands separator
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}
