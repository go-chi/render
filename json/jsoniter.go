// +build jsoniter

package json

import jsoniter "github.com/json-iterator/go"

var (
	json       = jsoniter.ConfigCompatibleWithStandardLibrary
	Marshal    = json.Marshal
	NewDecoder = json.NewDecoder
	NewEncoder = json.NewEncoder
)
