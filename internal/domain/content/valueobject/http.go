package valueobject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

func FetchAPIContent(contentType string, query string) ([]byte, error) {
	// Create the URL with the parameters
	fullURL := fmt.Sprintf("http://127.0.0.1:1314/api/search?type=%s&q=%s", url.QueryEscape(contentType), url.QueryEscape(query))

	// Send the GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	// Read the response body
	return io.ReadAll(resp.Body)
}

func postAPIContent(data map[string]interface{}) (string, error) {
	// Convert the map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshalling data: %v", err)
	}

	// Send the POST request
	resp, err := http.Post("http://127.0.0.1:1314/admin/edit", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check for non-OK status codes
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	// Get the "Location" header from the response
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header found in response")
	}

	// Return the modified URL
	return location, nil
}

func structToMap(obj interface{}) (map[string]interface{}, error) {
	// Ensure that obj is a struct
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %T", obj)
	}

	// Create a map to hold field names and values
	result := make(map[string]interface{})

	// Iterate over the struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i) // Get the field metadata
		fieldName := field.Name    // Get the field name

		// Use the JSON tag if available, otherwise use the struct field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = fieldName
		}

		// Store the field value in the map
		result[jsonTag] = v.Field(i).Interface()
	}

	return result, nil
}

func mapToYAML(data map[string]any) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}
