package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"github.com/urfave/cli/v2"
)

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string                        { return u.Email }
func (u *MyUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *MyUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

const rootDomain = "indocker.app"

var defaultDomains = []string{ //nolint:gochecknoglobals
	"*." + rootDomain,
	"*.app." + rootDomain,
	"*.apps." + rootDomain,
	"*.www." + rootDomain,
	"*.http." + rootDomain,
	"*.mail." + rootDomain,
	"*.m." + rootDomain,
	"*.go." + rootDomain,
	"*.static." + rootDomain,
	"*.img." + rootDomain,
	"*.media." + rootDomain,
	"*.admin." + rootDomain,
	"*.api." + rootDomain,
	"*.back." + rootDomain,
	"*.backend." + rootDomain,
	"*.front." + rootDomain,
	"*.frontend." + rootDomain,
	"*.srv." + rootDomain,
	"*.service." + rootDomain,
	"*.dev." + rootDomain,
	"*.db." + rootDomain,
	"*.test." + rootDomain,
	"*.demo." + rootDomain,
	"*.alpha." + rootDomain,
	"*.beta." + rootDomain,
	"*.x-docker." + rootDomain,
}

// exitFn is a function for application exiting.
var exitFn = os.Exit //nolint:gochecknoglobals

// main CLI application entrypoint.
func main() { exitFn(run()) }

// run this CLI application.
// Exit codes documentation: <https://tldp.org/LDP/abs/html/exitcodes.html>
func run() int { //nolint:funlen
	const (
		emailFlagName      = "email"
		apiKeyFlagName     = "api-key"
		domainsFlagName    = "domains"
		productionFlagName = "production"
		outCertFlagName    = "out-cert"
		outKeyFlagName     = "out-key"
	)

	var app = cli.App{
		Usage: "Domain certificate resolver",
		Action: func(c *cli.Context) error {
			var (
				email   = c.String(emailFlagName)
				key     = c.String(apiKeyFlagName)
				domains = c.StringSlice(domainsFlagName)
				prod    = c.Bool(productionFlagName)
				outCert = c.String(outCertFlagName)
				outKey  = c.String(outKeyFlagName)
			)

			// create a user (new accounts need an email and private key to start)
			privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				return err
			}

			var (
				user = &MyUser{
					Email: email,
					key:   privateKey,
				}
				config = lego.NewConfig(user)
			)

			if prod {
				config.CADirURL = lego.LEDirectoryProduction
			} else {
				config.CADirURL = lego.LEDirectoryStaging
			}

			// a client facilitates communication with the CA server.
			client, err := lego.NewClient(config)
			if err != nil {
				return err
			}

			// create and configure challenge provider
			dns, err := cloudflare.NewDNSProviderConfig(&cloudflare.Config{
				AuthEmail:          email,           // account email address
				AuthToken:          key,             // API token with DNS:Edit permission
				TTL:                200,             //nolint:gomnd // the TTL of the TXT record used for the DNS challenge
				PropagationTimeout: time.Minute,     // maximum waiting time for DNS propagation
				PollingInterval:    2 * time.Second, //nolint:gomnd // time between DNS propagation check
			})
			if err != nil {
				return err
			}

			// use the DNS challenge provider
			if err = client.Challenge.SetDNS01Provider(dns); err != nil {
				return err
			}

			// new users will need to register
			reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
			if err != nil {
				return err
			}
			user.Registration = reg

			// obtain certificates
			certificates, err := client.Certificate.Obtain(certificate.ObtainRequest{
				Domains: domains,
				Bundle:  true,
			})
			if err != nil {
				return err
			}

			const fileMode = 0o600

			if err = os.WriteFile(outCert, certificates.Certificate, fileMode); err != nil {
				return err
			}

			if err = os.WriteFile(outKey, certificates.PrivateKey, fileMode); err != nil {
				return err
			}

			log.Infof("Certificate and key saved to %s and %s", outCert, outKey)

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     emailFlagName,
				Usage:    "Email address for important account notifications",
				EnvVars:  []string{"EMAIL"},
				Required: true,
			},
			// Create a token (https://dash.cloudflare.com/profile/api-tokens) with the following permissions:
			// - Zone:Zone:Read
			// - Zone:DNS:Edit
			// Zone Resources: Include -- Specific zone -- <your-root-domain>
			&cli.StringFlag{
				Name:     apiKeyFlagName,
				Usage:    "Cloudflare API key",
				EnvVars:  []string{"API_KEY"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:    domainsFlagName,
				Usage:   "Domains to generate certificates for",
				EnvVars: []string{"DOMAINS"},
				Value:   cli.NewStringSlice(defaultDomains...),
			},
			&cli.BoolFlag{
				Name:    productionFlagName,
				Usage:   "Use production Let's Encrypt server (otherwise staging server is used)",
				EnvVars: []string{"PRODUCTION"},
			},
			&cli.StringFlag{
				Name:    outCertFlagName,
				Usage:   "File to write certificate to",
				EnvVars: []string{"OUT_CERT_FILE"},
				Value:   "certs/fullchain.pem",
			},
			&cli.StringFlag{
				Name:    outKeyFlagName,
				Usage:   "File to write private key to",
				EnvVars: []string{"OUT_KEY_FILE"},
				Value:   "certs/privkey.pem",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())

		return 1
	}

	return 0
}
