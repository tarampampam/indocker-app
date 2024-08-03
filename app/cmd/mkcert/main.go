package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"github.com/urfave/cli/v3"
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
func run() error { //nolint:funlen,gocognit,gocyclo
	var (
		emailFlag = cli.StringFlag{
			Name:     "email",
			Usage:    "Email address for important account notifications",
			Sources:  cli.EnvVars("EMAIL"),
			Required: true,
			OnlyOnce: true,
			Validator: func(s string) error {
				if s == "" {
					return errors.New("email address cannot be empty")
				}

				if !strings.Contains(s, "@") {
					return fmt.Errorf("invalid email address: %s", s)
				}

				return nil
			},
		}
		// Create a token (https://dash.cloudflare.com/profile/api-tokens) with the following permissions:
		// - Zone:Zone:Read
		// - Zone:DNS:Edit
		// Zone Resources: Include -- Specific zone -- <your-root-domain>
		apiKeyFlag = cli.StringFlag{
			Name:     "api-key",
			Usage:    "Cloudflare API key",
			Sources:  cli.EnvVars("API_KEY"),
			Required: true,
			OnlyOnce: true,
			Validator: func(s string) error {
				if s == "" {
					return errors.New("API key cannot be empty")
				}

				return nil
			},
		}
		productionFlag = cli.BoolFlag{
			Name:     "production",
			Usage:    "Use the production Let's Encrypt server; otherwise, the staging server will be used",
			Value:    false,
			Sources:  cli.EnvVars("PRODUCTION"),
			OnlyOnce: true,
		}
		outCertFlag = cli.StringFlag{
			Name:     "out-cert",
			Usage:    "File to write certificate to",
			Sources:  cli.EnvVars("OUT_CERT_FILE"),
			Value:    "certs/fullchain.pem",
			OnlyOnce: true,
			Validator: func(s string) error {
				if s == "" {
					return errors.New("file path cannot be empty")
				}

				if stat, err := os.Stat(s); err != nil {
					if os.IsNotExist(err) {
						return nil // file does not exist
					}

					return fmt.Errorf("failed to check if output file %s exists: %w", s, err)
				} else {
					if stat.IsDir() {
						return fmt.Errorf("%s is a directory", s)
					}

					return errors.New("output file already exists (if you wish to overwrite it, please delete it first)")
				}
			},
		}
		outKeyFlag = cli.StringFlag{
			Name:      "out-key",
			Usage:     "File to write private key to",
			Sources:   cli.EnvVars("OUT_KEY_FILE"),
			Value:     "certs/privkey.pem",
			OnlyOnce:  true,
			Validator: outCertFlag.Validator, // reuse the validator
		}
	)

	return (&cli.Command{
		Usage: "Domain certificate creator",
		Action: func(ctx context.Context, c *cli.Command) error {
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
				TTL:                200,             //nolint:mnd // the TTL of the TXT record used for the DNS challenge
				PropagationTimeout: time.Minute,     // maximum waiting time for DNS propagation
				PollingInterval:    2 * time.Second, //nolint:mnd // time between DNS propagation check
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

			log.Infof("Certificate and key saved to %s and %s\n", outCert, outKey)

			return nil
		},
		Flags: []cli.Flag{
			&emailFlag,
			&apiKeyFlag,
			&productionFlag,
			&outCertFlag,
			&outKeyFlag,
		},
	}).Run(context.Background(), os.Args)
}

type user struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *user) GetEmail() string                        { return u.Email }
func (u *user) GetRegistration() *registration.Resource { return u.Registration }
func (u *user) GetPrivateKey() crypto.PrivateKey        { return u.key }
