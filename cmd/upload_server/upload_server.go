package upload_server

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/sikalabs/slr/version"
	"github.com/spf13/cobra"
)

var FlagDataDir string
var FlagToken string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVar(&FlagDataDir, "data-dir", "", "Directory to store uploaded files (required)")
	Cmd.MarkFlagRequired("data-dir")
	Cmd.Flags().StringVar(&FlagToken, "token", "", "Optional token to require for uploads")
}

var Cmd = &cobra.Command{
	Use:   "upload-server",
	Short: "Simple file upload server",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		runServer()
	},
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func runServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(FlagDataDir))))

	fmt.Println("Listen on 0.0.0.0:8000, see http://127.0.0.1:8000")
	http.ListenAndServe(":8000", mux)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "slr upload-server, slr version %s\n", version.Version)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if FlagToken != "" {
		token := r.Header.Get("X-Upload-Token")
		if token != FlagToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error getting file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	unixTime := time.Now().Unix()
	randStr := randomString(8)
	dirName := fmt.Sprintf("%d_%s", unixTime, randStr)

	dirPath := filepath.Join(FlagDataDir, dirName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(dirPath, handler.Filename)
	f, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "/files/%s/%s\n", dirName, handler.Filename)
}
