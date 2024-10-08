package application

import (
	"context"
	"errors"
	"fmt"
	oapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	netlify "github.com/netlify/open-api/go/porcelain"
	ooapicontext "github.com/netlify/open-api/go/porcelain/context"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const accessToken = "nfp_AV9PCfK1QkGKeCSMJDokFN4167auccQq420d"
const siteID = "8f4e867c-b981-4959-8054-86c3f1f321e1"

type netlifyConfig struct {
	AuthToken     string `envconfig:"auth_token" required:"true"`
	SiteID        string `envconfig:"site_id" required:"true"`
	Directory     string `required:"true"`
	Draft         bool   `default:"true"`
	DeployMessage string `default:""`
	LogLevel      string `default:"warn"`
	LogFormat     string `default:"text"`
}

func DeployToNetlify(target string) error {
	c := &netlifyConfig{
		AuthToken:     accessToken,
		SiteID:        siteID,
		Directory:     path.Join(target, "public"),
		Draft:         false,
		DeployMessage: "Deployed from hugoverse",
		LogLevel:      "debug",
		LogFormat:     "text",
	}

	info, err := os.Stat(c.Directory)

	if os.IsNotExist(err) {
		return errors.New("file not exist")
	}

	if !info.IsDir() {
		return errors.New("target is not a directory")
	}

	logger := logrus.New()
	if err := setupLogging(c, logger); err != nil {
		logger.Fatal(err)
	}

	// Netlify setup
	client := setupNetlifyClient()
	ctx := setupContext(c, logger)

	// Deploy site
	resp, err := client.DoDeploy(ctx, &netlify.DeployOptions{
		SiteID:  c.SiteID,
		Dir:     c.Directory,
		IsDraft: c.Draft,
		Title:   c.DeployMessage,
	}, nil)
	if err != nil {
		logger.Fatalf("failed to deploy site: %s", err)
		return err
	}

	// Print the site URL
	if resp.DeploySslURL != "" {
		logger.Println("Deployed site: " + resp.DeploySslURL)
	} else if resp.DeployURL != "" {
		logger.Println("Deployed site: " + resp.DeployURL)
	}

	return nil
}

func setupLogging(c *netlifyConfig, logger *logrus.Logger) error {
	logLevel, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %s", err)
	}
	logger.SetLevel(logLevel)

	switch c.LogFormat {
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	default:
		logger.Warnf("invalid log format: %s", c.LogFormat)
	}
	return nil
}

func setupNetlifyClient() *netlify.Netlify {
	formats := strfmt.NewFormats()
	return netlify.NewHTTPClient(formats)
}

func setupContext(c *netlifyConfig, logger *logrus.Logger) ooapicontext.Context {
	ctx := ooapicontext.WithLogger(context.Background(), logger.WithFields(logrus.Fields{
		"source": "netlify",
	}))
	return ooapicontext.WithAuthInfo(ctx, oapiclient.BearerToken(c.AuthToken))
}
