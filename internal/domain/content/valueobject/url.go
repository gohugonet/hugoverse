package valueobject

import (
	"fmt"
	"net/url"
)

func extractTypeAndID(urlString string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	// Get the query parameters
	queryParams := parsedURL.Query()

	// Extract the 'type' and 'id' parameters
	postType := queryParams.Get("type")
	postID := queryParams.Get("id")

	// Combine them into the desired format
	result := fmt.Sprintf("%s:%s", postType, postID)

	return result, nil
}
