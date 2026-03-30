package api

import "github.com/tidwall/gjson"

// ResultIDName extracts _id and name from a standard {"result": {...}} API response.
func ResultIDName(body []byte) (id, name string) {
	result := gjson.GetBytes(body, "result")
	return result.Get("_id").String(), result.Get("name").String()
}
