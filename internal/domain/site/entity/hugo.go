package entity

func (p *Page) LinkTitle() string {
	return p.Title()
}

func (p *Page) Sites() *Sites {
	return &Sites{site: p.Site}
}

type Sites struct {
	site *Site
}

func (s *Sites) First() *Site {
	return s.site
}

func (s *Sites) Default() *Site {
	return s.First()
}

func (s *Site) IsMultilingual() bool {
	return s.IsMultiLingual()
}

func (s *Site) IsMultihost() bool {
	return false
}
