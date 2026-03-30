package cmdutil

import (
	"bytes"
	"testing"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newTestFactory() (*factory.Factory, *bytes.Buffer) {
	errBuf := &bytes.Buffer{}
	return &factory.Factory{
		IO: &iostreams.IOStreams{
			ErrOut: errBuf,
		},
	}, errBuf
}

func TestWriteCreated(t *testing.T) {
	f, buf := newTestFactory()
	body := []byte(`{"result":{"_id":"abc123","name":"My Network"}}`)

	WriteCreated(f, "Connector network", body)

	want := "Connector network \"My Network\" created. (id: abc123)\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteUpdated(t *testing.T) {
	f, buf := newTestFactory()
	body := []byte(`{"result":{"_id":"abc123","name":"My Network"}}`)

	WriteUpdated(f, "Connector network", body)

	want := "Connector network \"My Network\" (abc123) updated.\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteDeleted(t *testing.T) {
	f, buf := newTestFactory()

	WriteDeleted(f, "Device", "router-01", "def456")

	want := "Device \"router-01\" (def456) deleted.\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
