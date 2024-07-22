package entity

import (
	"fmt"
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
	o.setBasename()

	for _, of := range o.outputFormats() {
		switch o.pageKind {
		case valueobject.KindStatus404, valueobject.KindSitemap:
			if err := o.buildStandalone(of); err != nil {
				return err
			}
		case valueobject.KindHome, valueobject.KindSection, valueobject.KindTerm, valueobject.KindTaxonomy:
			if err := o.buildBrunch(of); err != nil {
				return err
			}
		case valueobject.KindPage:
			if err := o.buildPage(of); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown page kind: %s", o.pageKind)
		}
	}
	return nil
}

func (o *Output) buildBrunch(f output.Format) error {
	if o.pageKind == valueobject.KindHome {
		return o.buildHome(f)
	}

	return o.buildSection(f)
}

func (o *Output) buildSection(f output.Format) error {
	pb := valueobject.GetPagePathBuilder(f)
	defer valueobject.PutPagePathBuilder(pb)

	pb.FullSuffix = f.MediaType.FirstSuffix.FullSuffix
	pb.Add(o.source.Path().Dir())
	pb.Add(f.BaseName + pb.FullSuffix)
	if pb.IsHtmlIndex() {
		pb.LinkUpperOffset = 1
	}

	pb.Sanitize()
	target := &valueobject.Target{
		Prefix:                "",
		FilePath:              pb.PathFile(),
		SubResourceBaseTarget: pb.PathDir(),
	}
	o.targets = append(o.targets, target)

	return nil
}

func (o *Output) buildHome(f output.Format) error {
	pb := valueobject.GetPagePathBuilder(f)
	defer valueobject.PutPagePathBuilder(pb)

	pb.FullSuffix = f.MediaType.FirstSuffix.FullSuffix
	pb.Add(f.BaseName + pb.FullSuffix)
	if pb.IsHtmlIndex() {
		pb.LinkUpperOffset = 1
	}

	pb.Sanitize()
	target := &valueobject.Target{
		Prefix:                "",
		FilePath:              pb.PathFile(),
		SubResourceBaseTarget: pb.PathDir(),
	}
	o.targets = append(o.targets, target)

	return nil
}

func (o *Output) buildStandalone(f output.Format) error {
	pb := valueobject.GetPagePathBuilder(f)
	defer valueobject.PutPagePathBuilder(pb)

	pb.FullSuffix = f.MediaType.FirstSuffix.FullSuffix
	pb.IsUgly = true
	pb.BaseNameSameAsType = !o.source.IsBundle() && o.baseName != "" && o.baseName == f.BaseName
	pb.NoSubResources = true

	if dir := o.source.Path().Dir(); dir != "" {
		pb.Add(dir)
	}
	if o.baseName != "" {
		pb.Add(o.baseName)
	}
	pb.ConcatLast(pb.FullSuffix)

	pb.Sanitize()
	target := &valueobject.Target{
		Prefix:                o.source.Identity.Language(),
		FilePath:              pb.PathFile(),
		SubResourceBaseTarget: pb.PathDir(),
	}

	o.targets = append(o.targets, target)
	return nil
}

func (o *Output) buildPage(f output.Format) error {
	pb := valueobject.GetPagePathBuilder(f)
	defer valueobject.PutPagePathBuilder(pb)

	pb.FullSuffix = f.MediaType.FirstSuffix.FullSuffix
	pb.IsUgly = f.Ugly // default false
	pb.BaseNameSameAsType = !o.source.IsBundle() && o.baseName != "" && o.baseName == f.BaseName

	if dir := o.source.Path().ContainerDir(); dir != "" {
		pb.Add(dir)
	}

	bn := o.baseName
	if o.baseName == "" {
		return fmt.Errorf("no base name")
	}
	pb.Add(bn)

	pb.Add(f.BaseName + pb.FullSuffix)

	if pb.IsHtmlIndex() {
		// TODO, target file not care about html index
		pb.LinkUpperOffset = 1
	}

	pb.Sanitize()
	target := &valueobject.Target{
		Prefix:                o.source.Identity.Language(),
		FilePath:              pb.PathFile(),
		SubResourceBaseTarget: pb.PathDir(),
	}

	o.targets = append(o.targets, target)

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
