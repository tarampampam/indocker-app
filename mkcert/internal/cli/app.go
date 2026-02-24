package cli

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"github.com/urfave/cli/v3"
)

//go:generate go run app_generate.go

type app struct {
	c *cli.Command

	options struct {
		email         string
		apiKey        string
		isProduction  bool
		outArchiveDir string
	}
}

// NewApp creates new console application.
func NewApp() *cli.Command { //nolint:funlen
	var cliApp app

	cliApp.c = &cli.Command{
		Usage: "Domain certificate creator",
		Action: func(_ context.Context, c *cli.Command) error {
			{ // validate options (cli flag values)
				switch email := cliApp.options.email; {
				case email == "":
					return fmt.Errorf("email address cannot be empty")
				case !strings.Contains(email, "@"):
					return fmt.Errorf("invalid email address: %s", email)
				}

				switch apiKey := cliApp.options.apiKey; {
				case apiKey == "":
					return fmt.Errorf("API key cannot be empty")
				case len(apiKey) < 10: //nolint:mnd
					return fmt.Errorf("API key is too short")
				}

				if outDir := cliApp.options.outArchiveDir; outDir == "" {
					return fmt.Errorf("output directory path cannot be empty")
				} else {
					if stat, statErr := os.Stat(outDir); statErr != nil {
						return fmt.Errorf("failed to check if output directory %s exists: %w", outDir, statErr)
					} else if !stat.IsDir() {
						return fmt.Errorf("%s is not a directory", outDir)
					}
				}
			}

			var cert, certErr = cliApp.obtainCert()
			if certErr != nil {
				return fmt.Errorf("failed to obtain certificate: %w", certErr)
			}

			var dest = path.Join(cliApp.options.outArchiveDir, "archive.tar.gz")

			if err := cliApp.writeArchive(*cert, dest); err != nil {
				return fmt.Errorf("failed to write archive: %w", err)
			}

			log.Infof("Certificate archive has been written to %s", dest)

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "email",
				Usage:       "email",
				Sources:     cli.EnvVars("EMAIL"),
				Destination: &cliApp.options.email,
				OnlyOnce:    true,
				Config:      cli.StringConfig{TrimSpace: true},
			},
			// Create a token (https://dash.cloudflare.com/profile/api-tokens) with the following permissions:
			// - Zone:Zone:Read
			// - Zone:DNS:Edit
			// Zone Resources: Include -- Specific zone -- <your-root-domain>
			&cli.StringFlag{
				Name:        "api-key",
				Usage:       "cloudflare API key (required)",
				Sources:     cli.EnvVars("API_KEY"),
				Destination: &cliApp.options.apiKey,
				OnlyOnce:    true,
				Config:      cli.StringConfig{TrimSpace: true},
			},
			&cli.BoolFlag{
				Name:        "production",
				Usage:       "use the production Let's Encrypt server; otherwise, the staging server will be used",
				Value:       false,
				Destination: &cliApp.options.isProduction,
				Sources:     cli.EnvVars("PRODUCTION"),
				OnlyOnce:    true,
			},
			&cli.StringFlag{
				Name:        "out-archive-dir",
				Usage:       "directory to write archive to (required)",
				Sources:     cli.EnvVars("OUT_ARCHIVE_DIR"),
				Value:       ".",
				Destination: &cliApp.options.outArchiveDir,
				OnlyOnce:    true,
				Config:      cli.StringConfig{TrimSpace: true},
			},
		},
	}

	return cliApp.c
}

func (*app) domainsList() []string {
	const rootDomain = "indocker.app"

	var domains = make([]string, 0) //nolint:prealloc

	for _, subDomain := range []string{
		"*", "*.app", "*.apps", "*.www", "*.http", "*.mail", "*.m", "*.go", "*.static", "*.img", "*.media",
		"*.admin", "*.api", "*.back", "*.backend", "*.front", "*.frontend", "*.srv", "*.service", "*.dev",
		"*.db", "*.test", "*.demo", "*.alpha", "*.beta", "*.x-docker",
	} {
		domains = append(domains, strings.Join([]string{subDomain, rootDomain}, "."))
	}

	return domains
}

func (app *app) obtainCert() (*certificate.Resource, error) {
	privateKey, privateKeyErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if privateKeyErr != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", privateKeyErr)
	}

	var (
		usr    = &user{Email: app.options.email, key: privateKey}
		config = lego.NewConfig(usr)
	)

	if app.options.isProduction {
		config.CADirURL = lego.LEDirectoryProduction
	} else {
		config.CADirURL = lego.LEDirectoryStaging
	}

	client, clientErr := lego.NewClient(config)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to create Let's Encrypt client: %w", clientErr)
	}

	// create and configure challenge provider
	dnsProvider, providerErr := cloudflare.NewDNSProviderConfig(&cloudflare.Config{
		AuthEmail:          app.options.email,  // account email address
		AuthToken:          app.options.apiKey, // API token with DNS:Edit permission
		TTL:                200,                //nolint:mnd // the TTL of the TXT record used for the DNS challenge
		PropagationTimeout: time.Minute,        // maximum waiting time for DNS propagation
		PollingInterval:    2 * time.Second,    //nolint:mnd // time between DNS propagation check
	})
	if providerErr != nil {
		return nil, fmt.Errorf("failed to create Cloudflare DNS provider: %w", providerErr)
	}

	// use the DNS challenge provider
	if err := client.Challenge.SetDNS01Provider(dnsProvider); err != nil {
		return nil, fmt.Errorf("failed to set DNS challenge provider: %w", err)
	}

	//	return &certificate.Resource{ // FIXME: JUST FOR A TEST
	//		PrivateKey:  []byte("foo"),
	//		Certificate: []byte("bar"),
	//	}, nil

	// new users will need to register
	reg, regERr := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if regERr != nil {
		return nil, fmt.Errorf("failed to register user: %w", regERr)
	}

	usr.Registration = reg

	// obtain certificates
	certificates, obtainingErr := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: app.domainsList(),
		Bundle:  true,
	})
	if obtainingErr != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", obtainingErr)
	}

	return certificates, nil
}

func (*app) writeArchive(cert certificate.Resource, filepath string) error {
	var fd, fdErr = os.Create(filepath)
	if fdErr != nil {
		return fmt.Errorf("failed to create archive file: %w", fdErr)
	}

	defer func() { _ = fd.Close() }()

	var (
		gz, _ = gzip.NewWriterLevel(fd, gzip.BestCompression)
		tr    = tar.NewWriter(gz)
		now   = time.Now()
	)

	defer func() { _ = gz.Close(); _ = tr.Close() }()

	for fileName, content := range map[string][]byte{
		"privkey.pem":   cert.PrivateKey,
		"fullchain.pem": cert.Certificate,
	} {
		if err := tr.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     fileName,
			Size:     int64(len(content)),
			Mode:     0o600, //nolint:mnd
			ModTime:  now,
		}); err != nil {
			return err
		}

		if wrote, err := tr.Write(content); err != nil {
			return err
		} else if wrote != len(content) {
			return fmt.Errorf("failed to write %s file content (wrote bytes count mismatch)", fileName)
		}

		if err := tr.Flush(); err != nil {
			return err
		}
	}

	return nil
}

type user struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *user) GetEmail() string                        { return u.Email }
func (u *user) GetRegistration() *registration.Resource { return u.Registration }
func (u *user) GetPrivateKey() crypto.PrivateKey        { return u.key }
