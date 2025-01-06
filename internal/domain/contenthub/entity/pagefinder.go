package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
	"path"
	"path/filepath"
	"strings"
)

type PageFinder struct {
	Fs   contenthub.FsService
	home contenthub.Page

	PageMapper *PageMap
}

func (pf *PageFinder) GetPageFromPath(langIndex int, path string) (contenthub.Page, error) {
	p := paths.Parse(files.ComponentFolderContent, path)

	tree := pf.PageMapper.TreePages.Shape(0, langIndex)
	n := tree.Get(p.Base())

	if n != nil {
		ps, found := n.getPage()
		if !found {
			return valueobject.NilPage, nil
		}

		return ps, nil
	}

	return nil, nil
}

func (pf *PageFinder) GetPageRef(context contenthub.Page, ref string, home contenthub.Page) (contenthub.Page, error) {
	pf.home = home
	n, err := pf.getContentNode(context, true, ref)
	if err != nil {
		return nil, err
	}

	if n != nil {
		ps, found := n.getPage()
		if !found {
			return valueobject.NilPage, nil
		}

		return ps, nil
	}
	return nil, nil
}

const defaultContentExt = ".md"

func (pf *PageFinder) getContentNode(context contenthub.Page, isReflink bool, ref string) (*PageTreesNode, error) {
	ref = paths.ToSlashTrimTrailing(ref)
	inRef := ref
	if ref == "" {
		ref = "/"
	}

	if paths.HasExt(ref) {
		return pf.getContentNodeForRef(context, isReflink, true, inRef, ref)
	}

	// We are always looking for a content file and having an extension greatly simplifies the code that follows,
	// even in the case where the extension does not match this one.
	if ref == "/" {
		if n, err := pf.getContentNodeForRef(context, isReflink, false, inRef, "/_index"+defaultContentExt); n != nil || err != nil {
			return n, err
		}
	} else if strings.HasSuffix(ref, "/index") {
		if n, err := pf.getContentNodeForRef(context, isReflink, false, inRef, ref+"/index"+defaultContentExt); n != nil || err != nil {
			return n, err
		}
		if n, err := pf.getContentNodeForRef(context, isReflink, false, inRef, ref+defaultContentExt); n != nil || err != nil {
			return n, err
		}
	} else {
		if n, err := pf.getContentNodeForRef(context, isReflink, false, inRef, ref+defaultContentExt); n != nil || err != nil {
			return n, err
		}
	}

	return nil, nil
}

func (pf *PageFinder) getContentNodeForRef(context contenthub.Page, isReflink, hadExtension bool, inRef, ref string) (*PageTreesNode, error) {
	contentPathParser := paths.NewPathParser()

	if context != nil && !strings.HasPrefix(ref, "/") {
		// Branch pages: /mysection, "./mypage" => /mysection/mypage
		// Regular pages: /mysection/mypage.md, Path=/mysection/mypage, "./someotherpage" => /mysection/mypage/../someotherpage
		// Regular leaf bundles: /mysection/mypage/index.md, Path=/mysection/mypage, "./someotherpage" => /mysection/mypage/../someotherpage
		// Given the above, for regular pages we use the containing folder.
		var baseDir string
		if pi := context.Paths(); pi != nil {
			if pi.IsBranchBundle() || (hadExtension && strings.HasPrefix(ref, "../")) {
				baseDir = pi.Dir()
			} else {
				baseDir = pi.ContainerDir()
			}
		}

		// Try the page-relative path first.
		rel := path.Join(baseDir, ref)

		relPath, _ := contentPathParser.ParseBaseAndBaseNameNoIdentifier(files.ComponentFolderContent, rel)

		n, err := pf.getContentNodeFromPath(relPath, ref)
		if n != nil || err != nil {
			return n, err
		}

		// Try to look for a reverse lookup the specific file, because it has extension
		if hadExtension && context.PageFile() != nil {
			if n, err := pf.getContentNodeFromRefReverseLookup(inRef, context.PageFile().FileInfo()); n != nil || err != nil {
				return n, err
			}
		}

	}

	if strings.HasPrefix(ref, ".") {
		// Page relative, no need to look further.
		return nil, nil
	}

	relPath, nameNoIdentifier := contentPathParser.ParseBaseAndBaseNameNoIdentifier(files.ComponentFolderContent, ref)

	n, err := pf.getContentNodeFromPath(relPath, ref)

	if n != nil || err != nil {
		return n, err
	}

	if hadExtension && pf.home != nil && pf.home.PageFile() != nil {
		if n, err := pf.getContentNodeFromRefReverseLookup(inRef, pf.home.PageFile().FileInfo()); n != nil || err != nil {
			return n, err
		}
	}

	var doSimpleLookup bool
	if isReflink || context == nil {
		slashCount := strings.Count(inRef, "/")
		if slashCount <= 1 {
			doSimpleLookup = slashCount == 0 || ref[0] == '/'
		}
	}

	if !doSimpleLookup {
		return nil, nil
	}

	n = pf.PageMapper.pageReverseIndex.Get(nameNoIdentifier)
	if n == ambiguousContentNode {
		return nil, fmt.Errorf("page reference %q is ambiguous", inRef)
	}

	return n, nil
}

func (pf *PageFinder) getContentNodeFromRefReverseLookup(ref string, fi fs.FileMetaInfo) (*PageTreesNode, error) {
	dir := fi.FileName()
	if !fi.IsDir() {
		dir = filepath.Dir(fi.FileName())
	}

	realFilename := filepath.Join(dir, ref)

	pcs, err := pf.Fs.ReverseLookupContent(realFilename, true)
	if err != nil {
		return nil, err
	}

	// There may be multiple matches, but we will only use the first one.
	for _, pc := range pcs {
		pi := paths.Parse(pc.GetComponent(), pc.GetPath())
		if n := pf.PageMapper.TreePages.Get(pi.Base()); n != nil {
			return n, nil
		}
	}
	return nil, nil
}

func (pf *PageFinder) getContentNodeFromPath(s string, ref string) (*PageTreesNode, error) {
	n := pf.PageMapper.TreePages.Get(s)
	if n != nil {
		return n, nil
	}

	return nil, nil
}
