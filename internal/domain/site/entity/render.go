package entity

import "sync"

func pageRenderer(s *Site, pages <-chan *Page, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range pages {

	}
}
