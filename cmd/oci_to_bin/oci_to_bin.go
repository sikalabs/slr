package oci_to_bin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "oci-to-bin <image> <path>",
	Short: "Download a binary from an OCI artifact in a container registry",
	Long: `Download a binary from an OCI artifact created by "slr bin-to-oci"
(or "oras push <image> <file>:application/octet-stream").

The artifact must have exactly one layer, which is saved to <path>
as an executable.

Registry credentials are taken from the Docker config (docker login).`,
	Example: `  slr oci-to-bin ttl.sh/slr-bin:wip-1 /tmp/slr`,
	Args:    cobra.ExactArgs(2),
	Run: func(c *cobra.Command, args []string) {
		err := ociToBin(c.Context(), args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}
	},
}

func ociToBin(ctx context.Context, image, path string) error {
	ref, err := name.ParseReference(image)
	if err != nil {
		return err
	}

	opts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	desc, err := remote.Get(ref, opts...)
	if err != nil {
		return err
	}
	var manifest v1.Manifest
	if err := json.Unmarshal(desc.Manifest, &manifest); err != nil {
		return err
	}
	if len(manifest.Layers) != 1 {
		return fmt.Errorf("artifact %s must have exactly one layer, has %d", image, len(manifest.Layers))
	}

	layer, err := remote.Layer(ref.Context().Digest(manifest.Layers[0].Digest.String()), opts...)
	if err != nil {
		return err
	}
	rc, err := layer.Compressed()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Download to a temp file in the target directory and rename it at the end,
	// so a running binary is never overwritten partially
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, rc); err != nil {
		return err
	}
	if err := tmp.Chmod(0755); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Saved %s to %s\n", image, path)
	return nil
}
