package db

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// PutItem updates a single k/v in the item object
func PutItem(key string, value interface{}, item Item) error {
	kv := make(map[string]interface{})

	o, err := ItemAll(item)
	if err != nil {
		return err
	}

	if o == nil {
		o, err = marshalItem(item)
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(o, &kv)
	if err != nil {
		return err
	}

	// set k/v from params to decoded map
	kv[key] = value

	data := make(url.Values)
	for k, v := range kv {
		switch v.(type) {
		case string:
			data.Set(k, v.(string))

		case []string:
			vv := v.([]string)
			for i := range vv {
				data.Add(k, vv[i])
			}

		default:
			data.Set(k, fmt.Sprintf("%v", v))
		}
	}

	err = SetItem(data, item)
	if err != nil {
		return err
	}

	return nil
}
