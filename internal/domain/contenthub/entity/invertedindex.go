package entity

import (
	"context"
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/collections"
	"github.com/mdfriday/hugoverse/pkg/compare"
	"math"
	"sort"
	"sync"
	"time"
)

var (
	zeroDate = time.Time{}

	// DefaultConfig is the default related config.
	DefaultConfig = Config{
		Threshold: 80,
		Indices: contenthub.IndicesConfig{
			valueobject.IndexConfig{IndexName: "keywords", IndexWeight: 100, IndexType: valueobject.TypeBasic},
			valueobject.IndexConfig{IndexName: "date", IndexWeight: 10, IndexType: valueobject.TypeBasic},
		},
	}
)

var validTypes = map[string]bool{
	valueobject.TypeBasic:     true,
	valueobject.TypeFragments: true,
}

// InvertedIndex holds an inverted index, also sometimes named posting list, which
// lists, for every possible search term, the documents that contain that term.
type InvertedIndex struct {
	cfg   Config
	index map[string]map[contenthub.Keyword][]contenthub.Document
	// Counts the number of documents added to each index.
	indexDocCount map[string]int

	minWeight int
	maxWeight int

	// No modifications after this is set.
	finalized bool
}

// NewInvertedIndex creates a new InvertedIndex.
// Documents to index must be added in Add.
func NewInvertedIndex(cfg Config) *InvertedIndex {
	idx := &InvertedIndex{index: make(map[string]map[contenthub.Keyword][]contenthub.Document), indexDocCount: make(map[string]int), cfg: cfg}
	for _, conf := range cfg.Indices {
		idx.index[conf.Name()] = make(map[contenthub.Keyword][]contenthub.Document)
		if conf.Weight() < idx.minWeight {
			// By default, the weight scale starts at 0, but we allow
			// negative weights.
			idx.minWeight = conf.Weight()
		}
		if conf.Weight() > idx.maxWeight {
			idx.maxWeight = conf.Weight()
		}
	}
	return idx
}

// Search finds the documents matching any of the keywords in the given indices
// against query options in opts.
// The resulting document set will be sorted according to number of matches
// and the index weights, and any matches with a rank below the configured
// threshold (normalize to 0..100) will be removed.
// If an index name is provided, only that index will be queried.
func (idx *InvertedIndex) Search(ctx context.Context, opts valueobject.SearchOpts) ([]contenthub.Document, error) {
	var (
		queryElements []queryElement
		configs       contenthub.IndicesConfig
	)

	if len(opts.Indices) == 0 {
		configs = idx.cfg.Indices
	} else {
		configs = make(contenthub.IndicesConfig, len(opts.Indices))
		for i, indexName := range opts.Indices {
			cfg, found := idx.getIndexCfg(indexName)
			if !found {
				return nil, fmt.Errorf("index %q not found", indexName)
			}
			configs[i] = cfg
		}
	}

	for _, cfg := range configs {
		var keywords []contenthub.Keyword
		if opts.Document != nil {
			k, err := opts.Document.RelatedKeywords(cfg)
			if err != nil {
				return nil, err
			}
			keywords = append(keywords, k...)
		}
		if cfg.Type() == valueobject.TypeFragments {
			fmt.Println("TODO: fragments 1")
		}
		queryElements = append(queryElements, newQueryElement(cfg.Name(), keywords...))
	}
	for _, slice := range opts.NamedSlices {
		var keywords []contenthub.Keyword
		key := slice.KeyString()
		if key == "" {
			return nil, fmt.Errorf("index %q not valid", slice.Key)
		}
		conf, found := idx.getIndexCfg(key)
		if !found {
			return nil, fmt.Errorf("index %q not found", key)
		}

		for _, val := range slice.Values {
			k, err := conf.ToKeywords(val)
			if err != nil {
				return nil, err
			}
			keywords = append(keywords, k...)
		}
		queryElements = append(queryElements, newQueryElement(conf.Name(), keywords...))
	}

	if opts.Document != nil {
		return idx.searchDate(ctx, opts.Document, opts.Document.PublishDate(), queryElements...)
	}
	return idx.search(ctx, queryElements...)
}

func (idx *InvertedIndex) getIndexCfg(name string) (contenthub.IndexConfig, bool) {
	for _, conf := range idx.cfg.Indices {
		if conf.Name() == name {
			return conf, true
		}
	}

	return valueobject.IndexConfig{}, false
}

func (idx *InvertedIndex) search(ctx context.Context, query ...queryElement) ([]contenthub.Document, error) {
	return idx.searchDate(ctx, nil, zeroDate, query...)
}

func (idx *InvertedIndex) searchDate(ctx context.Context, self contenthub.Document, upperDate time.Time, query ...queryElement) ([]contenthub.Document, error) {
	matchm := make(map[contenthub.Document]*rank, 200)
	defer func() {
		for _, r := range matchm {
			putRank(r)
		}
	}()

	applyDateFilter := !idx.cfg.IncludeNewer && !upperDate.IsZero()
	var fragmentsFilter collections.SortedStringSlice

	for _, el := range query {
		setm, found := idx.index[el.Index]
		if !found {
			return []contenthub.Document{}, fmt.Errorf("index for %q not found", el.Index)
		}

		config, found := idx.getIndexCfg(el.Index)
		if !found {
			return []contenthub.Document{}, fmt.Errorf("index config for %q not found", el.Index)
		}

		for _, kw := range el.Keywords {
			if docs, found := setm[kw]; found {
				for _, doc := range docs {
					if compare.Eq(doc, self) {
						continue
					}

					if applyDateFilter {
						// Exclude newer than the limit given
						if doc.PublishDate().After(upperDate) {
							continue
						}
					}

					if config.Type() == valueobject.TypeFragments && config.ApplyFilter() {
						if fkw, ok := kw.(valueobject.FragmentKeyword); ok {
							fragmentsFilter = append(fragmentsFilter, string(fkw))
						}
					}

					r, found := matchm[doc]
					if !found {
						r = getRank(doc, config.Weight())
						matchm[doc] = r
					} else {
						r.addWeight(config.Weight())
					}
				}
			}
		}
	}

	if len(matchm) == 0 {
		return []contenthub.Document{}, nil
	}

	matches := make(ranks, 0, 100)

	for _, v := range matchm {
		avgWeight := v.Weight / v.Matches
		weight := norm(avgWeight, idx.minWeight, idx.maxWeight)
		threshold := idx.cfg.Threshold / v.Matches

		if weight >= threshold {
			matches = append(matches, v)
		}
	}

	sort.Stable(matches)
	sort.Strings(fragmentsFilter)

	result := make([]contenthub.Document, len(matches))

	for i, m := range matches {
		result[i] = m.Doc

		if len(fragmentsFilter) > 0 {
			fmt.Println("TODO: fragments 2")
		}
	}

	return result, nil
}

// Add documents to the inverted index.
// The value must support == and !=.
func (idx *InvertedIndex) Add(ctx context.Context, docs ...contenthub.Document) error {
	if idx.finalized {
		panic("index is finalized")
	}
	var err error
	for _, config := range idx.cfg.Indices {
		if config.Weight() == 0 {
			// Disabled
			continue
		}
		setm := idx.index[config.Name()]

		for _, doc := range docs {
			var added bool
			var words []contenthub.Keyword
			words, err = doc.RelatedKeywords(config)
			if err != nil {
				continue
			}

			for _, keyword := range words {
				added = true
				setm[keyword] = append(setm[keyword], doc)
			}

			if config.Type() == valueobject.TypeFragments {
				fmt.Println("TODO: fragments 4")
			}

			if added {
				idx.indexDocCount[config.Name()]++
			}
		}
	}

	return err
}

func (idx *InvertedIndex) Finalize(ctx context.Context) error {
	if idx.finalized {
		return nil
	}

	for _, config := range idx.cfg.Indices {
		if config.CardinalityThreshold() == 0 {
			continue
		}
		setm := idx.index[config.Name()]
		if idx.indexDocCount[config.Name()] == 0 {
			continue
		}

		// Remove high cardinality terms.
		numDocs := idx.indexDocCount[config.Name()]
		for k, v := range setm {
			percentageWithKeyword := int(math.Ceil(float64(len(v)) / float64(numDocs) * 100))
			if percentageWithKeyword > config.CardinalityThreshold() {
				delete(setm, k)
			}
		}

	}

	idx.finalized = true

	return nil
}

// Config is the top level configuration element used to configure how to retrieve
// related content in Hugo.
type Config struct {
	// Only include matches >= threshold, a normalized rank between 0 and 100.
	Threshold int

	// To get stable "See also" sections we, by default, exclude newer related pages.
	IncludeNewer bool

	// Will lower case all string values and queries to the indices.
	// May get better results, but at a slight performance cost.
	ToLower bool

	Indices contenthub.IndicesConfig
}

// queryElement holds the index name and keywords that can be used to compose a
// search for related content.
type queryElement struct {
	Index    string
	Keywords []contenthub.Keyword
}

func newQueryElement(index string, keywords ...contenthub.Keyword) queryElement {
	return queryElement{Index: index, Keywords: keywords}
}

type ranks []*rank

type rank struct {
	Doc     contenthub.Document
	Weight  int
	Matches int
}

func (r *rank) addWeight(w int) {
	r.Weight += w
	r.Matches++
}

var rankPool = sync.Pool{
	New: func() interface{} {
		return &rank{}
	},
}

func getRank(doc contenthub.Document, weight int) *rank {
	r := rankPool.Get().(*rank)
	r.Doc = doc
	r.Weight = weight
	r.Matches = 1
	return r
}

func putRank(r *rank) {
	rankPool.Put(r)
}

func (r ranks) Len() int      { return len(r) }
func (r ranks) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r ranks) Less(i, j int) bool {
	if r[i].Weight == r[j].Weight {
		if r[i].Doc.PublishDate() == r[j].Doc.PublishDate() {
			return r[i].Doc.Name() < r[j].Doc.Name()
		}
		return r[i].Doc.PublishDate().After(r[j].Doc.PublishDate())
	}
	return r[i].Weight > r[j].Weight
}

// normalizes num to a number between 0 and 100.
func norm(num, min, max int) int {
	if min > max {
		panic("min > max")
	}
	return int(math.Floor((float64(num-min) / float64(max-min) * 100) + 0.5))
}
