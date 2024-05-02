package valueobject

import "sync"

type PublishOnce struct {
	PublisherInit sync.Once
	PublisherErr  error
}
