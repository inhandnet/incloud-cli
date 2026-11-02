package device

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const gateDeviceID = "507f1f77bcf86cd799439011"

// gateServer answers bulk-validate with validateStatus/validateBody and every other
// request with dataBody, recording the paths it saw.
func gateServer(t *testing.T, validateStatus int, validateBody, dataBody string) (*httptest.Server, *[]string) {
	t.Helper()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if r.URL.Path == "/api/v1/product-compatibilities/bulk-validate" {
			w.WriteHeader(validateStatus)
			w.Write([]byte(validateBody))
			return
		}
		w.Write([]byte(dataBody))
	}))
	t.Cleanup(srv.Close)
	return srv, &paths
}

func verdictBody(compatID, verdict string) string {
	return `{"result":[{"id":"` + gateDeviceID + `","compatibilities":{"` + compatID + `":` + verdict + `}}]}`
}

type gateCase struct {
	name     string
	compatID string
	dataPath string
	cmd      string
}

var gateCases = []gateCase{
	{"interface", "nezha_interface", "/api/v1/devices/" + gateDeviceID + "/interfaces", "interface"},
	{"uplink", "nezha_uplink", "/api/v1/devices/" + gateDeviceID + "/uplinks", "uplink"},
}

func execGated(t *testing.T, srv *httptest.Server, c gateCase, args ...string) string {
	t.Helper()
	f, _ := newTestFactory(t, srv.URL)
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	if c.cmd == "interface" {
		root.AddCommand(NewCmdInterface(f))
	} else {
		root.AddCommand(NewCmdUplink(f))
	}
	root.SetArgs(append([]string{c.cmd, gateDeviceID}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return f.IO.Out.(*bytes.Buffer).String()
}

func requested(paths []string, path string) bool {
	for _, p := range paths {
		if strings.HasSuffix(p, " "+path) {
			return true
		}
	}
	return false
}

func TestGateSupportedFetchesData(t *testing.T) {
	for _, c := range gateCases {
		t.Run(c.name, func(t *testing.T) {
			srv, paths := gateServer(t, http.StatusOK,
				verdictBody(c.compatID, `{"support":true}`), `{"result":[]}`)
			execGated(t, srv, c)
			if !requested(*paths, c.dataPath) {
				t.Errorf("expected %s to be requested, got %v", c.dataPath, *paths)
			}
		})
	}
}

func TestGateUnsupportedSkipsData(t *testing.T) {
	for _, c := range gateCases {
		t.Run(c.name, func(t *testing.T) {
			srv, paths := gateServer(t, http.StatusOK,
				verdictBody(c.compatID, `{"support":false,"firmwareUnknown":false}`), `{"result":[]}`)
			out := execGated(t, srv, c, "-o", "json")
			for _, p := range *paths {
				if strings.Contains(p, "/api/v1/devices/") {
					t.Errorf("expected no device data request, got %s", p)
				}
			}
			var got map[string]any
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("unmarshal %q: %v", out, err)
			}
			if got["supported"] != false || got["compatibility"] != c.compatID || got["deviceId"] != gateDeviceID {
				t.Errorf("unexpected notice: %v", got)
			}
			if got["message"] != "This product does not support this feature." {
				t.Errorf("unexpected message: %v", got["message"])
			}
			if _, ok := got["minVersion"]; ok {
				t.Errorf("minVersion should be absent: %v", got)
			}
		})
	}
}

func TestGateUnsupportedMinVersion(t *testing.T) {
	for _, c := range gateCases {
		t.Run(c.name, func(t *testing.T) {
			srv, paths := gateServer(t, http.StatusOK,
				verdictBody(c.compatID, `{"support":false,"minVersion":"V2.0.6"}`), `{"result":[]}`)
			out := execGated(t, srv, c, "-o", "table")
			if requested(*paths, c.dataPath) {
				t.Errorf("expected %s not to be requested", c.dataPath)
			}
			want := "The current firmware of this device does not support this feature, please upgrade its version to V2.0.6 or above.\n"
			if out != want {
				t.Errorf("table output = %q, want %q", out, want)
			}
		})
	}
}

func TestGateUnknownFetchesData(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
	}{
		"validate fails": {http.StatusForbidden, `{"error":"forbidden"}`},
		"key absent":     {http.StatusOK, `{"result":[{"id":"` + gateDeviceID + `","compatibilities":{}}]}`},
	}
	for _, c := range gateCases {
		for name, v := range cases {
			t.Run(c.name+"/"+name, func(t *testing.T) {
				srv, paths := gateServer(t, v.status, v.body, `{"result":[]}`)
				execGated(t, srv, c)
				if !requested(*paths, c.dataPath) {
					t.Errorf("expected %s to be requested, got %v", c.dataPath, *paths)
				}
			})
		}
	}
}

func TestGateUnsupportedInterfaceSkipsRefresh(t *testing.T) {
	srv, paths := gateServer(t, http.StatusOK,
		verdictBody("nezha_interface", `{"support":false}`), `{"result":[]}`)
	execGated(t, srv, gateCases[0], "--refresh")
	if requested(*paths, "/api/v1/devices/"+gateDeviceID+"/interfaces/refresh") {
		t.Errorf("refresh should not be triggered when unsupported: %v", *paths)
	}
}

func TestUnsupportedMessageFirmwareUnknown(t *testing.T) {
	msg := unsupportedMessage(&compatibilityVerdict{FirmwareUnknown: true, MinVersion: "V2.0.6"})
	if !strings.Contains(msg, "firmware version is unknown") {
		t.Errorf("unexpected message: %q", msg)
	}
}
