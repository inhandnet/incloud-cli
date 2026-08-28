// Package cmdutil helpers for validating positional arguments that are used
// as MongoDB-style ObjectIds in API paths.
//
// The gateway routes requests by matching the shape of path segments. A
// non-ObjectId value (e.g. a typo, a serial number passed where a device id
// was expected) doesn't match any route and falls through to a generic
// "404 page not found" handler — which looks identical to a legitimate
// "resource not found" response but means something different ("you asked
// for a route that doesn't exist" vs. "the resource doesn't exist"). These
// validators catch the malformed case locally, before a request is even
// sent, with an error message that tells the caller (human or AI) exactly
// what was wrong and how to find the right id.
package cmdutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// objectIDPattern matches a MongoDB-style ObjectId: 24 hex characters.
var objectIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

// objectIDExample is used in error messages to show the expected shape.
const objectIDExample = "507f1f77bcf86cd799439011"

// IsObjectID reports whether s has the shape of a 24-character hex ObjectId.
// It only checks the shape, not whether the id actually exists.
func IsObjectID(s string) bool {
	return objectIDPattern.MatchString(s)
}

// FindCmd fills a lookup-command template (containing exactly one %s) in
// with value, producing a ready-to-run command, e.g.:
//
//	FindCmd("incloud device list -q %s", "17572") -> "incloud device list -q 17572"
func FindCmd(template, value string) string {
	return fmt.Sprintf(template, value)
}

// objectIDError builds a self-diagnosing error for a malformed id: it names
// the offending value, shows the expected format with an example, and gives
// a copy-pasteable command to find the correct id.
func objectIDError(resource, value, findCmd string) error {
	return fmt.Errorf(
		"invalid %s %q: expected a 24-character hex ObjectId (e.g. %s); run '%s' to find the %s",
		resource, value, objectIDExample, findCmd, resource,
	)
}

// ObjectIDArgs returns a cobra.PositionalArgs that first applies argCount
// (the command's existing arity check, e.g. cobra.ExactArgs(1)) and then
// validates that the argument at position pos (0-indexed) is a
// 24-character hex ObjectId.
//
// resource names the id in error messages (e.g. "device id"); findCmdTemplate
// is a command template with a single %s, filled in with the offending
// value, that helps locate the correct id (e.g. "incloud device list -q %s").
func ObjectIDArgs(argCount cobra.PositionalArgs, pos int, resource, findCmdTemplate string) cobra.PositionalArgs {
	return ObjectIDArgsFunc(argCount, pos, resource, func(args []string) string {
		return FindCmd(findCmdTemplate, args[pos])
	})
}

// ObjectIDArgsFunc is like ObjectIDArgs but lets the caller build the find
// command from the full argument list. Use this to scope the hint by an
// already-validated parent id (e.g. args[0], a group or device id) instead
// of the offending value itself — e.g. for a project id that can only be
// looked up within its parent group:
//
//	ObjectIDArgsFunc(cobra.ExactArgs(2), 1, "project id", func(args []string) string {
//	    return fmt.Sprintf("incloud device group project list %s", args[0])
//	})
func ObjectIDArgsFunc(argCount cobra.PositionalArgs, pos int, resource string, findCmd func(args []string) string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if argCount != nil {
			if err := argCount(cmd, args); err != nil {
				return err
			}
		}
		if pos >= len(args) {
			return nil
		}
		v := args[pos]
		if !IsObjectID(v) {
			return objectIDError(resource, v, findCmd(args))
		}
		return nil
	}
}

// ObjectIDCSVArgs is like ObjectIDArgs but the argument at pos is a
// comma-separated list of ids (e.g. "id1,id2,id3"); every element must be a
// valid ObjectId.
func ObjectIDCSVArgs(argCount cobra.PositionalArgs, pos int, resource, findCmdTemplate string) cobra.PositionalArgs {
	return ObjectIDCSVArgsFunc(argCount, pos, resource, func(_ []string, v string) string {
		return FindCmd(findCmdTemplate, v)
	})
}

// ObjectIDCSVArgsFunc is like ObjectIDCSVArgs but lets the caller build the
// find command from the full argument list and the offending element, e.g.
// to scope it by an already-validated parent id (args[0]) rather than the
// offending value itself.
func ObjectIDCSVArgsFunc(argCount cobra.PositionalArgs, pos int, resource string, findCmd func(args []string, value string) string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if argCount != nil {
			if err := argCount(cmd, args); err != nil {
				return err
			}
		}
		if pos >= len(args) {
			return nil
		}
		for _, v := range strings.Split(args[pos], ",") {
			if !IsObjectID(v) {
				return objectIDError(resource, v, findCmd(args, v))
			}
		}
		return nil
	}
}

// ObjectIDArgsAll validates that every positional argument is an ObjectId.
// Use it for variadic commands where each argument is an id of the same
// resource (e.g. "asset delete <id> [id...]").
func ObjectIDArgsAll(argCount cobra.PositionalArgs, resource, findCmdTemplate string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if argCount != nil {
			if err := argCount(cmd, args); err != nil {
				return err
			}
		}
		for _, v := range args {
			if !IsObjectID(v) {
				return objectIDError(resource, v, FindCmd(findCmdTemplate, v))
			}
		}
		return nil
	}
}
