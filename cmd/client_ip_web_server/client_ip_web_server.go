package client_ip_web_server

import (
	"fmt"
	"net/http"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "client-ip-web-server",
	Short: "Webserver which prints out a client IP",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		server()
	},
}

func server() {
	http.HandleFunc("/", handler)
	port := ":8000"
	fmt.Println("Listen on 0.0.0.0:8000, see http://127.0.0.1:8000")
	http.ListenAndServe(port, nil)
}

func getIP(r *http.Request) string {
	// Try getting IP from headers (useful when behind a proxy)
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		// Fallback to remote address
		ip = r.RemoteAddr
	}
	return ip
}

func handler(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	fmt.Fprintf(w, "%s\n", ip)
}
