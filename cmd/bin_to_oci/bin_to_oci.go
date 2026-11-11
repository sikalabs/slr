package bin_to_oci

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

const ArtifactType = "application/vnd.sikalabs.bin.v1"

func init() {
	root.Cmd.AddCommand(Cmd)
}

var Cmd = &cobra.Command{
	Use:   "bin-to-oci <path> <image>",
	Short: "Push a binary to a container registry as an OCI artifact",
	Long: `Push a binary to a container registry as an OCI artifact.

The artifact has an empty config and a single application/octet-stream
layer with the file name in the org.opencontainers.image.title annotation,
so it can be pulled also by "oras pull".

Registry credentials are taken from the Docker config (docker login).`,
	Example: `  slr bin-to-oci ./slr ttl.sh/slr-bin:wip-1`,
	Args:    cobra.ExactArgs(2),
	Run: func(c *cobra.Command, args []string) {
		err := binToOCI(c.Context(), args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}
	},
}

func binToOCI(ctx context.Context, path, image string) error {
	ref, err := name.ParseReference(image)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	opts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	config := static.NewLayer([]byte("{}"), types.OCIEmptyJSON)
	layer := static.NewLayer(data, "application/octet-stream")
	for _, blob := range []v1.Layer{config, layer} {
		if err := remote.WriteLayer(ref.Context(), blob, opts...); err != nil {
			return err
		}
	}

	configDesc, err := descriptor(config)
	if err != nil {
		return err
	}
	layerDesc, err := descriptor(layer)
	if err != nil {
		return err
	}
	layerDesc.Annotations = map[string]string{
		"org.opencontainers.image.title": filepath.Base(path),
	}

	manifest, err := json.Marshal(v1.Manifest{
		SchemaVersion: 2,
		MediaType:     types.OCIManifestSchema1,
		ArtifactType:  ArtifactType,
		Config:        *configDesc,
		Layers:        []v1.Descriptor{*layerDesc},
	})
	if err != nil {
		return err
	}

	if err := remote.Put(ref, rawManifest(manifest), opts...); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Pushed %s to %s\n", path, ref)
	return nil
}

func descriptor(l v1.Layer) (*v1.Descriptor, error) {
	mt, err := l.MediaType()
	if err != nil {
		return nil, err
	}
	digest, err := l.Digest()
	if err != nil {
		return nil, err
	}
	size, err := l.Size()
	if err != nil {
		return nil, err
	}
	return &v1.Descriptor{MediaType: mt, Digest: digest, Size: size}, nil
}

type rawManifest []byte

func (m rawManifest) RawManifest() ([]byte, error)        { return m, nil }
func (m rawManifest) MediaType() (types.MediaType, error) { return types.OCIManifestSchema1, nil }
