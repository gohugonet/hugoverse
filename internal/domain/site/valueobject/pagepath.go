package valueobject

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/pkg/output"
)

func NewPagePaths(ofs output.Formats, page contenthub.PageInfo) (PagePaths, error) {
	targetPathDescriptor, err := createTargetPathDescriptor(page)
	if err != nil {
		return PagePaths{}, err
	}

	outputFormats := ofs
	if len(outputFormats) == 0 {
		fmt.Println("outputFormats is null", outputFormats)

		return PagePaths{}, nil
	}

	pageOutputFormats := make(OutputFormats, len(outputFormats))
	targets := make(map[string]TargetPathsHolder)

	for i, f := range outputFormats {
		desc := targetPathDescriptor
		desc.Type = f
		paths := createTargetPaths(desc)

		var relPermalink, permalink string

		pageOutputFormats[i] = NewOutputFormat(relPermalink, permalink, f)

		// Use the main format for permalinks, usually HTML.
		permalinksIndex := 0
		targets[f.Name] = TargetPathsHolder{
			Paths:        paths,
			OutputFormat: pageOutputFormats[permalinksIndex],
		}
	}

	return PagePaths{
		outputFormats:        pageOutputFormats,
		firstOutputFormat:    pageOutputFormats[0],
		TargetPaths:          targets,
		targetPathDescriptor: targetPathDescriptor,
	}, nil
}

type PagePaths struct {
	outputFormats     OutputFormats
	firstOutputFormat OutputFormat

	TargetPaths          map[string]TargetPathsHolder
	targetPathDescriptor TargetPathDescriptor
}

func (l PagePaths) OutputFormats() OutputFormats {
	return l.outputFormats
}
