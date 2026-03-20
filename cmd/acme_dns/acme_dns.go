package acme_dns

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/acmedns"
	"github.com/go-acme/lego/v4/registration"
	vault "github.com/hashicorp/vault/api"
	"github.com/nrdcg/goacmedns"
	goacmedns_storage "github.com/nrdcg/goacmedns/storage"

	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var FlagDomains []string
var FlagEmail string
var FlagAcmeDNSAPIBase string
var FlagAccountUsername string
var FlagAccountPassword string
var FlagAccountFullDomain string
var FlagAccountSubDomain string
var FlagCertFile string
var FlagKeyFile string
var FlagVaultAddr string
var FlagVaultToken string
var FlagVaultPath string

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringSliceVarP(&FlagDomains, "domains", "d", nil, "Domains to obtain certificate for")
	Cmd.Flags().StringVarP(&FlagEmail, "email", "e", "", "Email address for ACME registration")
	Cmd.Flags().StringVar(&FlagAcmeDNSAPIBase, "acme-dns-api", "https://auth.acme-dns.io", "ACME DNS API base URL")
	Cmd.Flags().StringVar(&FlagAccountUsername, "account-username", "", "ACME DNS account username")
	Cmd.Flags().StringVar(&FlagAccountPassword, "account-password", "", "ACME DNS account password")
	Cmd.Flags().StringVar(&FlagAccountFullDomain, "account-full-domain", "", "ACME DNS account full domain")
	Cmd.Flags().StringVar(&FlagAccountSubDomain, "account-sub-domain", "", "ACME DNS account sub domain")
	Cmd.Flags().StringVar(&FlagCertFile, "cert-file", "cert.crt", "Output certificate file path")
	Cmd.Flags().StringVar(&FlagKeyFile, "key-file", "cert.key", "Output private key file path")
	Cmd.Flags().StringVar(&FlagVaultAddr, "vault-addr", "", "Vault server address (e.g. https://vault.example.com)")
	Cmd.Flags().StringVar(&FlagVaultToken, "vault-token", "", "Vault token")
	Cmd.Flags().StringVar(&FlagVaultPath, "vault-path", "", "Vault KV2 path to store certificate and key (e.g. secret/data/certs/mysite)")
	_ = Cmd.MarkFlagRequired("domains")
	_ = Cmd.MarkFlagRequired("email")
	_ = Cmd.MarkFlagRequired("account-username")
	_ = Cmd.MarkFlagRequired("account-password")
	_ = Cmd.MarkFlagRequired("account-full-domain")
	_ = Cmd.MarkFlagRequired("account-sub-domain")
}

var Cmd = &cobra.Command{
	Use:   "acme-dns",
	Short: "Obtain ACME certificate using ACME DNS challenge with acme-dns provider",
	Args:  cobra.NoArgs,
	Run: func(c *cobra.Command, args []string) {
		acme_dns(
			FlagDomains,
			FlagEmail,
			FlagAcmeDNSAPIBase,
			FlagAccountUsername,
			FlagAccountPassword,
			FlagAccountFullDomain,
			FlagAccountSubDomain,
			FlagCertFile,
			FlagKeyFile,
			FlagVaultAddr,
			FlagVaultToken,
			FlagVaultPath,
		)
	},
}

// User implements registration.User for lego.
type User struct {
	Email        string
	Registration *registration.Resource
	key          *ecdsa.PrivateKey
}

func (u *User) GetEmail() string                        { return u.Email }
func (u *User) GetRegistration() *registration.Resource { return u.Registration }
func (u *User) GetPrivateKey() crypto.PrivateKey        { return u.key }

// memoryStorage is an in-memory implementation of goacmedns.Storage
// pre-populated with account data passed directly in code.
type memoryStorage struct {
	accounts map[string]goacmedns.Account
}

func newMemoryStorage(accounts map[string]goacmedns.Account) *memoryStorage {
	return &memoryStorage{accounts: accounts}
}

func (m *memoryStorage) Save(_ context.Context) error { return nil }

func (m *memoryStorage) Put(_ context.Context, domain string, account goacmedns.Account) error {
	m.accounts[domain] = account
	return nil
}

func (m *memoryStorage) Fetch(_ context.Context, domain string) (goacmedns.Account, error) {
	account, ok := m.accounts[domain]
	if !ok {
		return goacmedns.Account{}, goacmedns_storage.ErrDomainNotFound
	}
	return account, nil
}

func (m *memoryStorage) FetchAll(_ context.Context) (map[string]goacmedns.Account, error) {
	return m.accounts, nil
}

func acme_dns(
	domains []string,
	email string,
	acmeDNSAPIBase string,
	accountUsername string,
	accountPassword string,
	accountFullDomain string,
	accountSubDomain string,
	certFile string,
	keyFile string,
	vaultAddr string,
	vaultToken string,
	vaultPath string,
) {
	account := goacmedns.Account{
		Username:   accountUsername,
		Password:   accountPassword,
		FullDomain: accountFullDomain,
		SubDomain:  accountSubDomain,
	}
	accounts := make(map[string]goacmedns.Account, len(domains))
	for _, d := range domains {
		accounts[d] = account
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	user := &User{Email: email, key: privateKey}

	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryProduction
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	goacmednsClient, err := goacmedns.NewClient(acmeDNSAPIBase)
	if err != nil {
		log.Fatal(err)
	}

	store := newMemoryStorage(accounts)

	// NewDNSProviderClient is used here to pass a custom in-memory storage.
	//nolint:staticcheck
	provider, err := acmedns.NewDNSProviderClient(goacmednsClient, store)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Challenge.SetDNS01Provider(provider,
		dns01.AddRecursiveNameservers([]string{"8.8.8.8:53"}),
	)
	if err != nil {
		log.Fatal(err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		var cnameErr acmedns.ErrCNAMERequired
		if errors.As(err, &cnameErr) {
			log.Fatalf("CNAME setup required for %q:\n  %s CNAME %s.\nAdd this record and re-run.", cnameErr.Domain, cnameErr.FQDN, cnameErr.Target)
		}
		log.Fatal(err)
	}

	if err := os.WriteFile(certFile, cert.Certificate, 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(keyFile, cert.PrivateKey, 0600); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Certificate obtained for: %s\n", strings.Join(domains, ", "))
	fmt.Printf("Certificate URL: %s\n", cert.CertURL)
	fmt.Printf("Certificate saved to: %s\n", certFile)
	fmt.Printf("Private key saved to: %s\n", keyFile)

	if vaultAddr != "" && vaultToken != "" && vaultPath != "" {
		vaultConfig := vault.DefaultConfig()
		vaultConfig.Address = vaultAddr
		vaultClient, err := vault.NewClient(vaultConfig)
		if err != nil {
			log.Fatal(err)
		}
		vaultClient.SetToken(vaultToken)

		// vaultPath format: <mount>/<secret-path>, e.g. "secret/certs/mysite"
		parts := strings.SplitN(vaultPath, "/", 2)
		if len(parts) != 2 {
			log.Fatalf("--vault-path must be in format <mount>/<secret-path>, got: %s", vaultPath)
		}
		vaultMount, vaultSecretPath := parts[0], parts[1]

		_, err = vaultClient.KVv2(vaultMount).Put(context.Background(), vaultSecretPath, map[string]interface{}{
			"tls.crt": string(cert.Certificate),
			"tls.key": string(cert.PrivateKey),
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Certificate stored in Vault at: %s\n", vaultPath)
	}
}
