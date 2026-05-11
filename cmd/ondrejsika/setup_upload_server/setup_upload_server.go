package setup_upload_server

import (
	"log"
	"os"

	"github.com/sikalabs/slr/cmd/ondrejsika"
	"github.com/spf13/cobra"
)

func init() {
	ondrejsika.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "setup-upload-server",
	Short: "Configure SLR_UPLOAD_SERVER_ORIGIN to the default upload server",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		setupUploadServer()
	},
}

func setupUploadServer() {
	origin := "https://slr-upload-server-public.panda.k8s.sl.zone"
	tmp := "/etc/SLR_UPLOAD_SERVER_ORIGIN"

	if err := os.WriteFile(tmp, []byte(origin+"\n"), 0644); err != nil {
		log.Fatal(err)
	}
}
