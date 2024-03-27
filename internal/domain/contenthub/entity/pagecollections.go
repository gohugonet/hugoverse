package entity

// PageCollections contains the page collections for a site.
type PageCollections struct {
	PageMap *PageMap
}

type PageMap struct {
	*ContentMap
}
