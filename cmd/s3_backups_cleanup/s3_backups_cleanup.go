package s3_backups_cleanup

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var (
	FlagBucket          string
	FlagPrefix          string
	FlagEndpoint        string
	FlagAccessKey       string
	FlagSecretKey       string
	FlagRegion          string
	FlagDateTimePattern string
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(&FlagBucket, "bucket", "b", getEnv("S3_BUCKET", ""), "S3 bucket name")
	Cmd.Flags().StringVarP(&FlagPrefix, "prefix", "p", getEnv("S3_PREFIX", ""), "Prefix/path in bucket to search for backups")
	Cmd.Flags().StringVarP(&FlagEndpoint, "endpoint", "e", getEnv("S3_ENDPOINT", ""), "S3 endpoint URL (for MinIO or custom S3)")
	Cmd.Flags().StringVarP(&FlagAccessKey, "access-key", "a", getEnv("AWS_ACCESS_KEY_ID", ""), "AWS/MinIO access key")
	Cmd.Flags().StringVarP(&FlagSecretKey, "secret-key", "s", getEnv("AWS_SECRET_ACCESS_KEY", ""), "AWS/MinIO secret key")
	Cmd.Flags().StringVarP(&FlagRegion, "region", "r", getEnv("AWS_REGION", "us-east-1"), "AWS region")
	Cmd.Flags().StringVarP(&FlagDateTimePattern, "datetime-pattern", "d", getEnv("S3_DATETIME_PATTERN", `\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}`), "Regex pattern to extract datetime from filename")
}

var Cmd = &cobra.Command{
	Use:   "s3-backups-cleanup",
	Short: "Clean up S3/MinIO backups based on retention policy",
	Long: `Clean up S3/MinIO backups with intelligent retention:
- This week and last week: keep all backups
- This month and last month: keep one backup per day
- Older: keep one backup per month

Requires confirmation before deletion.`,
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if FlagBucket == "" {
			fmt.Println("Error: bucket is required (use --bucket flag or S3_BUCKET env var)")
			os.Exit(1)
		}

		if err := cleanupBackups(); err != nil {
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

type backup struct {
	key      string
	datetime time.Time
}

func cleanupBackups() error {
	ctx := context.Background()

	// Create S3 client
	client, err := createS3Client(ctx)
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	// List all objects in bucket with prefix
	fmt.Printf("Listing backups in bucket '%s' with prefix '%s'...\n", FlagBucket, FlagPrefix)
	backups, err := listBackups(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	fmt.Printf("Found %d backups\n", len(backups))

	// Determine which backups to keep and delete
	toKeep, toDelete := categorizeBackups(backups)

	fmt.Printf("\nBackups to keep: %d\n", len(toKeep))
	fmt.Printf("Backups to delete: %d\n", len(toDelete))

	// Show all backups with their status
	fmt.Println("\n=== All Backups ===")

	// Create a map for quick lookup
	deleteMap := make(map[string]bool)
	for _, b := range toDelete {
		deleteMap[b.key] = true
	}

	for _, b := range backups {
		status := "[KEEP]   "
		if deleteMap[b.key] {
			status = "[DELETE] "
		}
		fmt.Printf("%s %s (%s)\n", status, b.key, b.datetime.Format("2006-01-02 15:04:05"))
	}

	if len(toDelete) == 0 {
		fmt.Println("\nNo backups to delete")
		return nil
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("Total backups: %d\n", len(backups))
	fmt.Printf("Will keep: %d\n", len(toKeep))
	fmt.Printf("Will delete: %d\n", len(toDelete))

	// Ask for confirmation
	if !askForConfirmation() {
		fmt.Println("Deletion cancelled")
		return nil
	}

	// Delete backups
	fmt.Println("\nDeleting backups...")
	deleted := 0
	for _, b := range toDelete {
		err := deleteBackup(ctx, client, b.key)
		if err != nil {
			fmt.Printf("  Error deleting %s: %v\n", b.key, err)
		} else {
			fmt.Printf("  Deleted: %s\n", b.key)
			deleted++
		}
	}

	fmt.Printf("\nSuccessfully deleted %d/%d backups\n", deleted, len(toDelete))
	return nil
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

func listBackups(ctx context.Context, client *s3.Client) ([]backup, error) {
	var backups []backup
	dateTimeRegex := regexp.MustCompile(FlagDateTimePattern)

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(FlagBucket),
		Prefix: aws.String(FlagPrefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)

			// Extract datetime from filename
			matches := dateTimeRegex.FindString(key)
			if matches == "" {
				// Skip files that don't match the datetime pattern
				continue
			}

			// Parse datetime (format: 2024-11-29_14-35-02)
			datetime, err := parseDateTime(matches)
			if err != nil {
				fmt.Printf("Warning: could not parse datetime from '%s': %v\n", key, err)
				continue
			}

			backups = append(backups, backup{
				key:      key,
				datetime: datetime,
			})
		}
	}

	// Sort by datetime (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].datetime.After(backups[j].datetime)
	})

	return backups, nil
}

func parseDateTime(dateTimeStr string) (time.Time, error) {
	// Try different datetime formats
	formats := []string{
		"2006-01-02_15-04-05",
		"2006-01-02_15-04",
		"2006-01-02",
		"20060102_150405",
		"20060102",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateTimeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse datetime: %s", dateTimeStr)
}

func categorizeBackups(backups []backup) (keep []backup, delete []backup) {
	now := time.Now()

	// Calculate time boundaries
	startOfLastWeek := startOfWeek(now.AddDate(0, 0, -7))
	startOfLastMonth := startOfMonth(now.AddDate(0, -1, 0))

	// Track kept backups by month and day for deduplication
	keptDays := make(map[string]bool)
	keptMonths := make(map[string]bool)

	for _, b := range backups {
		var shouldKeep bool

		if b.datetime.After(startOfLastWeek) || b.datetime.Equal(startOfLastWeek) {
			// This week or last week: keep all
			shouldKeep = true
		} else if b.datetime.After(startOfLastMonth) || b.datetime.Equal(startOfLastMonth) {
			// This month or last month: keep one per day
			dayKey := b.datetime.Format("2006-01-02")
			if !keptDays[dayKey] {
				shouldKeep = true
				keptDays[dayKey] = true
			}
		} else {
			// Older: keep one per month
			monthKey := b.datetime.Format("2006-01")
			if !keptMonths[monthKey] {
				shouldKeep = true
				keptMonths[monthKey] = true
			}
		}

		if shouldKeep {
			keep = append(keep, b)
		} else {
			delete = append(delete, b)
		}
	}

	return keep, delete
}

func startOfWeek(t time.Time) time.Time {
	// Start week on Monday
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, -(weekday - 1))
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func askForConfirmation() bool {
	fmt.Print("\nDo you want to proceed with deletion? (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}

func deleteBackup(ctx context.Context, client *s3.Client, key string) error {
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(FlagBucket),
		Key:    aws.String(key),
	})
	return err
}
