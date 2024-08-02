package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"github.com/urfave/cli/v2"
)

var defaultDomains = func() []string { //nolint:gochecknoglobals
	const rootDomain = "indocker.app"

	var domains = make([]string, 0)

	for _, subDomain := range []string{
		"*", "*.app", "*.apps", "*.www", "*.http", "*.mail", "*.m", "*.go", "*.static", "*.img", "*.media",
		"*.admin", "*.api", "*.back", "*.backend", "*.front", "*.frontend", "*.srv", "*.service", "*.dev",
		"*.db", "*.test", "*.demo", "*.alpha", "*.beta", "*.x-docker",
	} {
		domains = append(domains, strings.Join([]string{subDomain, rootDomain}, "."))
	}

	return domains
}()

// main CLI application entrypoint.
func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())

		os.Exit(1)
	}
}

// run this CLI application.
func run() error { //nolint:funlen
	var (
		emailFlag = cli.StringFlag{
			Name:     "email",
			Usage:    "Email address for important account notifications",
			EnvVars:  []string{"EMAIL"},
			Required: true,
		}
		// Create a token (https://dash.cloudflare.com/profile/api-tokens) with the following permissions:
		// - Zone:Zone:Read
		// - Zone:DNS:Edit
		// Zone Resources: Include -- Specific zone -- <your-root-domain>
		apiKeyFlag = cli.StringFlag{
			Name:     "api-key",
			Usage:    "Cloudflare API key",
			EnvVars:  []string{"API_KEY"},
			Required: true,
		}
		productionFlag = cli.BoolFlag{
			Name:    "production",
			Usage:   "Use the production Let's Encrypt server; otherwise, the staging server will be used",
			Value:   false,
			EnvVars: []string{"PRODUCTION"},
		}
		outCertFlag = cli.StringFlag{
			Name:    "out-cert",
			Usage:   "File to write certificate to",
			EnvVars: []string{"OUT_CERT_FILE"},
			Value:   "certs/fullchain.pem",
		}
		outKeyFlag = cli.StringFlag{
			Name:    "out-key",
			Usage:   "File to write private key to",
			EnvVars: []string{"OUT_KEY_FILE"},
			Value:   "certs/privkey.pem",
		}
	)

	return (&cli.App{
		Usage: "Domain certificate creator",
		Action: func(c *cli.Context) error {
			var (
				email        = c.String(emailFlag.Name)
				key          = c.String(apiKeyFlag.Name)
				isProduction = c.Bool(productionFlag.Name)
				outCert      = c.String(outCertFlag.Name)
				outKey       = c.String(outKeyFlag.Name)
			)

			privateKey, privKeyErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if privKeyErr != nil {
				return fmt.Errorf("failed to generate private key: %w", privKeyErr)
			}

			var (
				usr    = &user{Email: email, key: privateKey}
				config = lego.NewConfig(usr)
			)

			if isProduction {
				config.CADirURL = lego.LEDirectoryProduction
			} else {
				config.CADirURL = lego.LEDirectoryStaging
			}

			client, clientErr := lego.NewClient(config)
			if clientErr != nil {
				return fmt.Errorf("failed to create Let's Encrypt client: %w", clientErr)
			}

			// create and configure challenge provider
			dnsProvider, providerErr := cloudflare.NewDNSProviderConfig(&cloudflare.Config{
				AuthEmail:          email,           // account email address
				AuthToken:          key,             // API token with DNS:Edit permission
				TTL:                200,             //nolint:gomnd // the TTL of the TXT record used for the DNS challenge
				PropagationTimeout: time.Minute,     // maximum waiting time for DNS propagation
				PollingInterval:    2 * time.Second, //nolint:gomnd // time between DNS propagation check
			})
			if providerErr != nil {
				return fmt.Errorf("failed to create Cloudflare DNS provider: %w", providerErr)
			}

			// use the DNS challenge provider
			if err := client.Challenge.SetDNS01Provider(dnsProvider); err != nil {
				return err
			}

			// new users will need to register
			reg, regERr := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
			if regERr != nil {
				return fmt.Errorf("failed to register user: %w", regERr)
			}

			usr.Registration = reg

			// obtain certificates
			certificates, obtainingErr := client.Certificate.Obtain(certificate.ObtainRequest{
				Domains: defaultDomains,
				Bundle:  true,
			})
			if obtainingErr != nil {
				return fmt.Errorf("failed to obtain certificate: %w", obtainingErr)
			}

			const fileMode = 0o600

			if err := os.WriteFile(outCert, certificates.Certificate, fileMode); err != nil {
				return fmt.Errorf("failed to write certificate to file: %w", err)
			}

			if err := os.WriteFile(outKey, certificates.PrivateKey, fileMode); err != nil {
				return fmt.Errorf("failed to write private key to file: %w", err)
			}

			_, _ = fmt.Fprintf(os.Stdout, "Certificate and key saved to %s and %s\n", outCert, outKey)

			return nil
		},
		Flags: []cli.Flag{
			&emailFlag,
			&apiKeyFlag,
			&productionFlag,
			&outCertFlag,
			&outKeyFlag,
		},
	}).Run(os.Args)
}

type user struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *user) GetEmail() string                        { return u.Email }
func (u *user) GetRegistration() *registration.Resource { return u.Registration }
func (u *user) GetPrivateKey() crypto.PrivateKey        { return u.key }
