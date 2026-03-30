package cmdutil

import (
	"fmt"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

// WriteCreated writes a "<resource> created" confirmation to stderr.
func WriteCreated(f *factory.Factory, resource string, body []byte) {
	id, name := api.ResultIDName(body)
	fmt.Fprintf(f.IO.ErrOut, "%s %q created. (id: %s)\n", resource, name, id)
}

// WriteUpdated writes a "<resource> updated" confirmation to stderr.
func WriteUpdated(f *factory.Factory, resource string, body []byte) {
	id, name := api.ResultIDName(body)
	fmt.Fprintf(f.IO.ErrOut, "%s %q (%s) updated.\n", resource, name, id)
}

// WriteDeleted writes a "<resource> deleted" confirmation to stderr.
func WriteDeleted(f *factory.Factory, resource, name, id string) {
	fmt.Fprintf(f.IO.ErrOut, "%s %q (%s) deleted.\n", resource, name, id)
}
