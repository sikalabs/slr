package migrate_mysql_database

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var (
	FlagSource string
	FlagTarget string
)

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagSource, "source", "", "Source MySQL URL, e.g. mysql://user:pass@host:port/database (required)")
	Cmd.Flags().StringVar(&FlagTarget, "target", "", "Target MySQL URL, e.g. mysql://user:pass@host:port/database (required)")
	Cmd.MarkFlagRequired("source")
	Cmd.MarkFlagRequired("target")
}

var Cmd = &cobra.Command{
	Use:   "migrate-mysql-database",
	Short: "Migrate a MySQL database from a source to a target server",
	Long: `Migrate a MySQL database from a source server/database to a target server/database.

It streams the output of "mysqldump" for the source database directly into
"mysql" for the target database, so both the "mysqldump" and "mysql" binaries
must be available on PATH.`,
	Args: cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		if err := migrate(FlagSource, FlagTarget); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

type mysqlConn struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func parseMySQLURL(raw string) (*mysqlConn, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid MySQL URL: %w", err)
	}
	if u.Scheme != "mysql" {
		return nil, fmt.Errorf("invalid MySQL URL scheme %q, expected \"mysql\"", u.Scheme)
	}
	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		return nil, fmt.Errorf("MySQL URL must include a database name, e.g. mysql://user:pass@host:port/database")
	}
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("MySQL URL must include a host")
	}
	port := u.Port()
	if port == "" {
		port = "3306"
	}
	password, _ := u.User.Password()
	return &mysqlConn{
		Host:     host,
		Port:     port,
		User:     u.User.Username(),
		Password: password,
		Database: database,
	}, nil
}

func migrate(sourceURL, targetURL string) error {
	source, err := parseMySQLURL(sourceURL)
	if err != nil {
		return fmt.Errorf("source: %w", err)
	}
	target, err := parseMySQLURL(targetURL)
	if err != nil {
		return fmt.Errorf("target: %w", err)
	}

	if _, err := exec.LookPath("mysqldump"); err != nil {
		return fmt.Errorf("mysqldump binary not found in PATH: %w", err)
	}
	if _, err := exec.LookPath("mysql"); err != nil {
		return fmt.Errorf("mysql binary not found in PATH: %w", err)
	}

	fmt.Printf("Migrating database %q from %s:%s to %q on %s:%s...\n",
		source.Database, source.Host, source.Port, target.Database, target.Host, target.Port)

	dumpCmd := exec.Command("mysqldump",
		"-h", source.Host,
		"-P", source.Port,
		"-u", source.User,
		"--single-transaction",
		"--routines",
		"--triggers",
		source.Database,
	)
	dumpCmd.Env = append(os.Environ(), "MYSQL_PWD="+source.Password)
	dumpCmd.Stderr = os.Stderr

	restoreCmd := exec.Command("mysql",
		"-h", target.Host,
		"-P", target.Port,
		"-u", target.User,
		target.Database,
	)
	restoreCmd.Env = append(os.Environ(), "MYSQL_PWD="+target.Password)
	restoreCmd.Stderr = os.Stderr
	restoreCmd.Stdout = os.Stdout

	pipe, err := dumpCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe between mysqldump and mysql: %w", err)
	}
	restoreCmd.Stdin = pipe

	if err := restoreCmd.Start(); err != nil {
		return fmt.Errorf("failed to start mysql: %w", err)
	}
	if err := dumpCmd.Run(); err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}
	if err := restoreCmd.Wait(); err != nil {
		return fmt.Errorf("mysql restore failed: %w", err)
	}

	fmt.Println("Migration completed successfully!")
	return nil
}
