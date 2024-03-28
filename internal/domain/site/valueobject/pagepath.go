package valueobject

import (
	"fmt"
)

func NewPagePaths(ofs Formats, kind string, sec []string, dir, basename string) (PagePaths, error) {
	targetPathDescriptor, err := createTargetPathDescriptor(kind, sec, dir, basename)
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
