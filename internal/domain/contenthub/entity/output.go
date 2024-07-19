package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/output"
)

type Output struct {
	targets []*valueobject.Target

	baseName string

	source   *Source
	pageKind string
}

func (o *Output) Build() error {
	for _, of := range o.outputFormats() {
		if err := o.buildTargets(of); err != nil {
			return err
		}
	}
	return nil
}

func (o *Output) buildTargets(f output.Format) error {
	pb := valueobject.GetPagePathBuilder(f)
	defer valueobject.PutPagePathBuilder(pb)

	pb.FullSuffix = f.MediaType.FirstSuffix.FullSuffix
	pb.IsUgly = f.Ugly // default false
	pb.BaseNameSameAsType = !o.source.IsBundle() && o.baseName != "" && o.baseName == f.BaseName

	switch f {
	case output.HTTPStatusHTMLFormat, output.SitemapFormat:
		pb.NoSubResources = true
	}

	if o.source.BundleType.IsBranchBundle() {
		if o.pageKind != valueobject.KindHome {
			pb.Add(o.source.Path().Dir())
		}

		pb.Add(f.BaseName + pb.FullSuffix)
	} else {
		if dir := o.source.Path().ContainerDir(); dir != "" {
			pb.Add(dir)
		}

		bn := o.baseName
		if o.baseName == "" {
			bn = o.source.Path().BaseNameNoIdentifier()
		}
		pb.Add(bn)

		pb.Add(f.BaseName + pb.FullSuffix)
	}

	if pb.IsHtmlIndex() {
		pb.LinkUpperOffset = 1
	}
	pb.PrefixPath = o.source.Identity.Language()

	return nil
}

func (o *Output) setBasename() {
	switch o.pageKind {
	case valueobject.KindStatus404:
		o.baseName = output.HTTPStatusHTMLFormat.BaseName
	case valueobject.KindSitemap:
		o.baseName = output.SitemapFormat.BaseName
	default:
		o.baseName = o.source.Path().BaseNameNoIdentifier()
	}
}

func (o *Output) outputFormats() output.Formats {
	var outputFormats output.Formats
	switch o.pageKind {
	case valueobject.KindStatus404:
		outputFormats = output.Formats{output.HTTPStatusHTMLFormat}
	case valueobject.KindSitemap:
		outputFormats = output.Formats{output.SitemapFormat}
	default:
		d := o.defaultOutputFormats()
		for _, v := range d[o.pageKind] {
			f, _ := allFormats().GetByName(v)
			outputFormats = append(outputFormats, f)
		}
	}

	return outputFormats
}

func (o *Output) defaultOutputFormats() map[string][]string {
	allFormats := allFormats()

	htmlOut, _ := allFormats.GetByName(output.HTMLFormat.Name)

	defaultListTypes := []string{htmlOut.Name}

	return map[string][]string{
		valueobject.KindPage:     {htmlOut.Name},
		valueobject.KindHome:     defaultListTypes,
		valueobject.KindSection:  defaultListTypes,
		valueobject.KindTerm:     defaultListTypes,
		valueobject.KindTaxonomy: defaultListTypes,
	}
}

func allFormats() output.Formats {
	return output.DefaultFormats
}
