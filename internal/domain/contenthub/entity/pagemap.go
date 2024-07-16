package entity

import (
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

type PageMap struct {
	*ContentSpec

	// Main storage for all pages.
	*PageTrees

	Cache *valueobject.Cache

	PageBuilder *PageBuilder

	//TODO : add for the cascade in the future
	//assembleChanges *valueobject.WhatChanged

	Log loggers.Logger
}

func (m *PageMap) InsertResourceNode(key string, node *PageTreesNode) {
	tree := m.TreeResources

	commit := tree.Lock(true)
	defer commit()

	tree.InsertIntoValuesDimension(key, node)
}

func (m *PageMap) AddFi(f *valueobject.File) error {
	if f.IsDir() {
		return nil
	}

	ps, err := newPageSource(f, m.Cache)
	if err != nil {
		return err
	}

	key := ps.Path().Base()

	switch ps.BundleType {
	case valueobject.BundleTypeFile:
		m.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert resource file: %q", f.FileName())
			},
		))
		m.InsertResourceNode(key, newPageTreesNode(ps))

	case valueobject.BundleTypeContentResource:
		m.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert content resource: %q", f.FileName())
			},
		))
		p, err := m.PageBuilder.WithSource(ps).Build()
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}
		if p == nil {
			// Disabled page.
			return nil
		}

		m.InsertResourceNode(key, newPageTreesNode(p))

	default:
		m.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert bundle: %q", f.FileName())
			},
		))
		// A content file.
		p, err := m.PageBuilder.WithSource(ps).Build()
		if err != nil {
			return err
		}

		//TODO check pi changes
		m.TreePages.InsertWithLock(ps.Path().Base(), newPageTreesNode(p))

	}
	return nil
}

func (m *PageMap) Assemble() error {
	if err := m.assembleStructurePages(); err != nil {
		return err
	}

	if err := m.applyAggregates(); err != nil {
		return err
	}

	if err := m.cleanPages(); err != nil {
		return err
	}

	if err := m.assembleTerms(); err != nil {
		return err
	}

	// TODO: apply aggregates cascade and dates to taxonomy and terms

	return nil
}

func (m *PageMap) assembleTerms() error {
	if err := m.PageBuilder.Term.Assemble(m.TreePages, m.PageBuilder); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) cleanPages() error {
	// TODO: clean all the draft, expired, scheduled in the future pages
	return nil
}

func (m *PageMap) applyAggregates() error {

	// TODO
	// Apply cascade Aggregates to pages
	// Apply Dates to home, section, and meta changed pages
	// Apply cascade to source page
	// Use linked list to connect all the cascade in the same path

	// Restore all the changed pages in the cache

	return nil
}

func (m *PageMap) assembleStructurePages() error {

	if err := m.addMissingTaxonomies(); err != nil {
		return err
	}

	if err := m.PageBuilder.Section.Assemble(m.TreePages, m.PageBuilder); err != nil {
		return err
	}

	if err := m.addMissingStandalone(); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) addMissingTaxonomies() error {
	tree := m.TreePages

	commit := tree.Lock(true)
	defer commit()

	if err := m.PageBuilder.Taxonomy.Assemble(tree, m.PageBuilder); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) addMissingStandalone() error {
	tree := m.TreePages

	commit := tree.Lock(true)
	defer commit()

	if err := m.PageBuilder.Standalone.Assemble(tree, m.PageBuilder); err != nil {
		return err
	}

	return nil
}
