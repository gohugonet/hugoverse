package application

import (
	"context"
	"errors"
	"fmt"
	oapiclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/netlify/open-api/v2/go/models"
	netlify "github.com/netlify/open-api/v2/go/porcelain"
	ooapicontext "github.com/netlify/open-api/v2/go/porcelain/context"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

type netlifyConfig struct {
	AuthToken     string `envconfig:"auth_token" required:"true"`
	SiteID        string `envconfig:"site_id" required:"true"`
	Directory     string `required:"true"`
	Draft         bool   `default:"true"`
	DeployMessage string `default:""`
	LogLevel      string `default:"warn"`
	LogFormat     string `default:"text"`
}

func DeployToNetlify(target string, deployment *valueobject.Deployment, domain *valueobject.Domain, token string) error {
	c := &netlifyConfig{
		AuthToken:     token,
		SiteID:        deployment.SiteID,
		Directory:     path.Join(target, "public"),
		Draft:         false,
		DeployMessage: "Deployed by MDFriday",
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

	fmt.Println("Deploying to Netlify...", deployment.SiteName, domain.FullDomain())

	// 检查 SiteID 是否为空
	if c.SiteID == "" {
		// 创建新 Netlify 站点
		newSite, err := client.CreateSite(ctx, &models.SiteSetup{
			Site: models.Site{
				//AccountSlug:  "admin-zbpioce",
				Name:         deployment.SiteName,
				CustomDomain: domain.FullDomain(),
				Ssl:          true,
			},
			SiteSetupAllOf1: models.SiteSetupAllOf1{},
		}, true) // 设置 configureDNS 为 true
		if err != nil {
			logger.Errorf("failed to create Netlify site: %s", err)
			return err
		}

		// 更新 SiteID
		c.SiteID = newSite.ID
		deployment.SiteID = newSite.ID
		deployment.Status = "deploying"
		logger.Println("Created new site with ID: " + c.SiteID)
	}

	// Deploy site
	resp, err := client.DoDeploy(ctx, &netlify.DeployOptions{
		SiteID:  c.SiteID,
		Dir:     c.Directory,
		IsDraft: c.Draft,
		Title:   c.DeployMessage,
	}, nil)
	if err != nil {
		logger.Errorf("failed to deploy site: %s", err)
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
