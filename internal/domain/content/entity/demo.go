package entity

import "fmt"

type Demo struct {
	Item

	Title      string `json:"title"`
	Artist     string `json:"artist"`
	Rating     int    `json:"rating"`
	Opinion    string `json:"opinion"`
	SpotifyUrl string `json:"spotify_url"`
}

// String defines how a Song is printed. Update it using more descriptive
// fields from the Song struct type
func (d *Demo) String() string {
	return fmt.Sprintf("Song: %s", d.UUID)
}
