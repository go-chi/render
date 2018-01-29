// +build !jsoniter

package json

import "encoding/json"

var (
	Marshal    = json.Marshal
	NewDecoder = json.NewDecoder
	NewEncoder = json.NewEncoder
)
