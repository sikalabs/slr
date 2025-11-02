package s3_bucket_size

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
	Cmd.Flags().StringVarP(&FlagBucket, "bucket", "b", getEnv("S3_BUCKET", ""), "S3 bucket name (if not provided, shows all buckets)")
	Cmd.Flags().StringVarP(&FlagPrefix, "prefix", "p", getEnv("S3_PREFIX", ""), "Prefix/path in bucket to calculate size for")
	Cmd.Flags().StringVarP(&FlagEndpoint, "endpoint", "e", getEnv("S3_ENDPOINT", ""), "S3 endpoint URL (for MinIO or custom S3)")
	Cmd.Flags().StringVarP(&FlagAccessKey, "access-key", "a", getEnv("AWS_ACCESS_KEY_ID", ""), "AWS/MinIO access key")
	Cmd.Flags().StringVarP(&FlagSecretKey, "secret-key", "s", getEnv("AWS_SECRET_ACCESS_KEY", ""), "AWS/MinIO secret key")
	Cmd.Flags().StringVarP(&FlagRegion, "region", "r", getEnv("AWS_REGION", "us-east-1"), "AWS region")
}

var Cmd = &cobra.Command{
	Use:   "s3-bucket-size",
	Short: "Calculate S3/MinIO bucket size",
	Long: `Calculate the total size of S3/MinIO buckets.

If --bucket is provided, shows size for that specific bucket.
If --bucket is not provided, shows sizes for all buckets.
Optionally filter by --prefix to calculate size of a specific path.`,
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if err := getBucketSize(); err != nil {
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

type bucketSize struct {
	name        string
	objectCount int64
	totalSize   int64
}

func getBucketSize() error {
	ctx := context.Background()

	// Create S3 client
	client, err := createS3Client(ctx)
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	if FlagBucket != "" {
		// Calculate size for specific bucket
		return calculateSingleBucketSize(ctx, client, FlagBucket)
	}

	// Calculate sizes for all buckets
	return calculateAllBucketsSizes(ctx, client)
}

func createS3Client(ctx context.Context) (*s3.Client, error) {
	var cfg aws.Config
	var err error

	if FlagAccessKey != "" && FlagSecretKey != "" {
		// Use static credentials
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(FlagRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				FlagAccessKey,
				FlagSecretKey,
				"",
			)),
		)
	} else {
		// Use default credential chain
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(FlagRegion),
		)
	}

	if err != nil {
		return nil, err
	}

	// Create S3 client with custom endpoint if provided (for MinIO)
	var clientOpts []func(*s3.Options)
	if FlagEndpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(FlagEndpoint)
			o.UsePathStyle = true // MinIO requires path-style addressing
		})
	}

	return s3.NewFromConfig(cfg, clientOpts...), nil
}

func calculateSingleBucketSize(ctx context.Context, client *s3.Client, bucketName string) error {
	prefixMsg := ""
	if FlagPrefix != "" {
		prefixMsg = fmt.Sprintf(" (prefix: %s)", FlagPrefix)
	}
	fmt.Printf("Calculating size for bucket '%s'%s...\n", bucketName, prefixMsg)

	size, err := getBucketSizeInfo(ctx, client, bucketName)
	if err != nil {
		return fmt.Errorf("failed to calculate bucket size: %w", err)
	}

	fmt.Printf("\nBucket: %s\n", bucketName)
	if FlagPrefix != "" {
		fmt.Printf("Prefix: %s\n", FlagPrefix)
	}
	fmt.Printf("Objects: %s\n", formatNumber(size.objectCount))
	fmt.Printf("Total Size: %s\n", formatBytes(size.totalSize))

	return nil
}

func calculateAllBucketsSizes(ctx context.Context, client *s3.Client) error {
	fmt.Println("Listing all buckets...")

	// List all buckets
	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to list buckets: %w", err)
	}

	if len(result.Buckets) == 0 {
		fmt.Println("No buckets found")
		return nil
	}

	fmt.Printf("Found %d bucket(s)\n\n", len(result.Buckets))

	var sizes []bucketSize
	var totalSize int64
	var totalObjects int64

	for _, bucket := range result.Buckets {
		bucketName := aws.ToString(bucket.Name)
		fmt.Printf("Calculating size for '%s'...\n", bucketName)

		size, err := getBucketSizeInfo(ctx, client, bucketName)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		sizes = append(sizes, size)
		totalSize += size.totalSize
		totalObjects += size.objectCount

		fmt.Printf("  Objects: %s, Size: %s\n", formatNumber(size.objectCount), formatBytes(size.totalSize))
	}

	// Print summary
	fmt.Println("\n=== Summary ===")
	fmt.Printf("%-40s %15s %20s\n", "Bucket", "Objects", "Size")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, size := range sizes {
		fmt.Printf("%-40s %15s %20s\n", size.name, formatNumber(size.objectCount), formatBytes(size.totalSize))
	}

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("%-40s %15s %20s\n", "TOTAL", formatNumber(totalObjects), formatBytes(totalSize))

	return nil
}

func getBucketSizeInfo(ctx context.Context, client *s3.Client, bucketName string) (bucketSize, error) {
	var totalSize int64
	var objectCount int64

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(FlagPrefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return bucketSize{}, err
		}

		for _, obj := range page.Contents {
			totalSize += aws.ToInt64(obj.Size)
			objectCount++
		}
	}

	return bucketSize{
		name:        bucketName,
		objectCount: objectCount,
		totalSize:   totalSize,
	}, nil
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
	return fmt.Sprintf("%.2f %s (%d bytes)", float64(bytes)/float64(div), units[exp], bytes)
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
