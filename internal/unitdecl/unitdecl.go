// Package unitdecl holds the unit declarations for CLI commands.
//
// A declaration here is an audit record. It states: someone read this command's
// payload and confirmed the unit of every time-valued field in it, so those
// fields may be renamed to carry their unit into -o json / -o yaml / --jq
// output. A command that is not declared in this file makes no unit assertion
// and its structured output is passed through untouched.
//
// Keeping every declaration in one file is the point: the list is meant to be
// read top to bottom during review. Declarations are exported symbols rather
// than strings keyed by command name so that naming a command that does not
// exist fails to compile — the IM-3180 spec itself misspelled
// "device uplink perf" as "device uplink-perf" in four places and survived
// several rounds of review, which a string-keyed lookup would have shipped.
//
// Before adding a command here:
//
//   - Find the field's origin. For latency / jitter the authority is the device
//     firmware: nezha-agent/pkg/message/message.go:117 says
//     "latency, us, set value to -1 if timeout". Do not infer a unit from a
//     platform DTO, a portal chart, or an existing document — those were the
//     sources that produced the misreadings this file exists to prevent.
//   - Confirm no service in the path rescales the value. The uplink path does
//     not (InterfaceListener -> Uplink.from -> flux Uplink are all straight
//     assignments), but a path that divides somewhere needs a different unit
//     suffix, or no declaration at all.
//   - Check that the same field name does not carry two different units inside
//     one payload. Granularity stops at the command, so that case cannot be
//     expressed here.
//   - Only annotate a sentinel that a device-side or protocol source states
//     explicitly. -1 (timeout) is the only sentinel on this link; there is no
//     clamp ceiling, so values such as 2000000 microseconds are real
//     measurements and annotating them would mask a genuine fault.
package unitdecl

import "github.com/inhandnet/incloud-cli/internal/iostreams"

// latencyJitter renames the microsecond-valued latency / jitter fields and maps
// the -1 timeout sentinel to null plus a status field.
//
// jitter gets the same -1 handling as latency: the device interaction protocol
// uses one convention for both, and a negative jitter is physically impossible.
//
// Full words, not "Us": the abbreviation collides with US, the country code,
// which appears frequently in these payloads.
var latencyJitter = iostreams.FieldRewrites{
	"latency": {
		To:        "latencyMicroseconds",
		StatusKey: "latencyStatus",
		Timeout:   true,
	},
	"jitter": {
		To:        "jitterMicroseconds",
		StatusKey: "jitterStatus",
		Timeout:   true,
	},
}

// offlineDuration renames the second-valued offline duration fields.
// No sentinel values are known for these.
var offlineDuration = iostreams.FieldRewrites{
	"totalOfflineDuration": {To: "totalOfflineDurationSeconds"},
	"avgOfflineDuration":   {To: "avgOfflineDurationSeconds"},
	"maxOfflineDuration":   {To: "maxOfflineDurationSeconds"},
}

// The declared commands. Each entry is one audited payload.
var (
	// DeviceUplink covers `incloud device uplink`, whose payload carries
	// latency / jitter per uplink as reported by the device.
	DeviceUplink = latencyJitter
	// DeviceUplinkGet covers `incloud device uplink get`, which returns a single
	// entry of the same payload as DeviceUplink.
	DeviceUplinkGet = latencyJitter
	// DeviceUplinkPerf covers `incloud device uplink perf`. Its structured
	// output is columnar: field names live in "columns" and samples in "values".
	DeviceUplinkPerf = latencyJitter
	// DeviceInterface covers `incloud device interface`, which nests the same
	// per-interface latency / jitter one level deeper.
	DeviceInterface = latencyJitter
	// DeviceLogMqtt covers `incloud device log mqtt`, where the fields appear
	// inside the echoed device message payloads.
	DeviceLogMqtt = latencyJitter
	// OverviewOffline covers `incloud overview offline`, whose offline durations
	// are seconds (nezha-core aggregates them from second-resolution offline
	// records; no sentinel is defined for them).
	OverviewOffline = offlineDuration
)
