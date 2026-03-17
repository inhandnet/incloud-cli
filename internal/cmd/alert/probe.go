package alert

import (
	"net/url"
	"strconv"
)

// applyProbeParams adds AlertProbe filter parameters to the query string.
func applyProbeParams(q url.Values, after, before, status string, priority *int, device, group string, types []string, ack, query string) {
	if after != "" {
		q.Set("from", after)
	}
	if before != "" {
		q.Set("to", before)
	}
	if status != "" {
		q.Set("status", status)
	}
	if priority != nil {
		q.Set("priority", strconv.Itoa(*priority))
	}
	if device != "" {
		q.Set("deviceId", device)
	}
	if group != "" {
		q.Set("deviceGroupId", group)
	}
	for _, t := range types {
		q.Add("type", t)
	}
	if ack != "" {
		q.Set("ack", ack)
	}
	if query != "" {
		q.Set("entityName", query)
	}
}
