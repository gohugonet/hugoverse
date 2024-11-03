package entity

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/gohugonet/hugoverse/pkg/herrors"
)

func (c *Content) PreviewTarget(contentType, id, status string) (string, string, error) {
	content, err := c.getContent(contentType, id)
	if err != nil {
		return "", "", err
	}

	if site, ok := content.(*valueobject.Site); ok {
		dir, _, err := c.Hugo.tempDir(site.Title, c.Repo.UserDataDir())
		c.Log.Debugf("temp dir: %s", dir)

		if err != nil {
			c.Log.Errorf("failed to get temp dir: %v", err)
			return "", "", err
		}

		writer := newWriter(c.Log)
		go writer.startDumpFiles()

		writer.files <- c.Hugo.goModFile(dir)

		confFile, previewPubDir, err := c.sitePreviewConfigFile(site, dir)
		if err != nil {
			c.Log.Errorf("failed to get site config file: %v", err)
			writer.close()
			return "", "", err
		}
		writer.files <- confFile

		if err := c.writeSitePosts(site.ID, dir, writer.files); err != nil {
			c.Log.Errorf("failed to write site posts: %v", err)
			writer.close()
			return "", "", err
		}

		writer.close()

		err = <-writer.errs
		if err != nil {
			return "", "", fmt.Errorf("failed to render pages: %w", herrors.ImproveIfNilPointer(err))
		}

		return dir, previewPubDir, nil
	}

	return "", "", errors.New("only site could be built")

}

func (c *Content) sitePreviewConfigFile(site *valueobject.Site, dir string) (*valueobject.File, string, error) {
	previewDir, previewPubDir, err := c.Hugo.previewDir()
	if err != nil {
		c.Log.Errorf("failed to get preview dir: %v", err)
		return nil, "", err
	}

	baseURL, err := c.Hugo.previewPath(previewDir)
	if err != nil {
		c.Log.Errorf("failed to get preview path: %v", err)
		return nil, "", err
	}
	site.BaseURL = baseURL

	f, err := c.Hugo.siteConfigFile(site, dir)
	return f, previewPubDir, err
}
