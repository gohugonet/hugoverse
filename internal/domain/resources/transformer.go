package resources

import (
	"github.com/gohugonet/hugoverse/pkg/identity"
	"io"
)

type DependenceSvc interface {
	DependencyManager() identity.Manager
}

type PublishSvc interface {
	PublishContentToTarget(content, target string) error
	OpenPublishFileForWriting(relTargetPath string) (io.WriteCloser, error)
}
