package tunnel

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/xtaci/smux"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

const (
	defaultNgrokPort     = 4443
	msgNewVisitorConn    = "NewVisitorConn"
	msgNewVisitorConnRes = "NewVisitorConnResp"
)

// forwardOptions holds configuration for local port forwarding.
type forwardOptions struct {
	localPort int
	tunnelID  string
	token     string
	ngrokAddr string
}

func NewCmdTunnelForward(f *factory.Factory) *cobra.Command {
	var (
		localPort int
		ngrokPort int
		token     string
	)

	cmd := &cobra.Command{
		Use:   "forward <tunnel-id>",
		Short: "Forward a tunnel to a local port",
		Long: `Forward an existing tunnel to a local TCP port.

Connects to the tunnel via TLS and starts a local TCP listener.
Connect to the local port with ssh, telnet, curl, or any TCP client.

Works with tunnels created by any command: open-cli, open-web, oobm connect, etc.

Press Ctrl+C to stop.`,
		Example: `  # Forward a tunnel to a random local port
  incloud tunnel forward nhcqr3rzpxqfiviqdu3c7x4o

  # Forward to a specific port
  incloud tunnel forward nhcqr3rzpxqfiviqdu3c7x4o --port 2222

  # Forward with token (required when server has JWT enabled)
  incloud tunnel forward nhcqr3rzpxqfiviqdu3c7x4o --token <jwt>

  # Use with OOBM: connect a resource, then forward its tunnel
  incloud oobm connect <resource-id> --service ssh:22:cli -o json
  incloud tunnel forward <tunnelId-from-output> --token <token-from-output> --port 2222
  ssh root@localhost -p 2222`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			opts := &forwardOptions{
				localPort: localPort,
				tunnelID:  args[0],
				token:     token,
				ngrokAddr: fmt.Sprintf("%s:%d", actx.NgrokHost(), ngrokPort),
			}
			return runForward(f, opts)
		},
	}

	cmd.Flags().IntVarP(&localPort, "port", "p", 0, "Local port to listen on (0 = random)")
	cmd.Flags().StringVar(&token, "token", "", "Auth token for the tunnel (from tunnel creation response)")
	cmd.Flags().IntVar(&ngrokPort, "ngrok-port", defaultNgrokPort, "Ngrok TCP proxy port")

	return cmd
}

// runForward establishes a multiplexed visitor connection to the ngrok server
// and forwards local TCP connections as smux streams.
func runForward(f *factory.Factory, opts *forwardOptions) error {
	session, err := dialMuxSession(opts.ngrokAddr, opts.tunnelID, opts.token)
	if err != nil {
		return err
	}
	defer session.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.localPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	actualPort := listener.Addr().(*net.TCPAddr).Port
	fmt.Fprintf(f.IO.Out, "Listening on 127.0.0.1:%d\n", actualPort)
	fmt.Fprintf(f.IO.ErrOut, "Forwarding tunnel %s\n", opts.tunnelID)
	fmt.Fprintf(f.IO.ErrOut, "Press Ctrl+C to stop\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		fmt.Fprintf(f.IO.ErrOut, "\nShutting down...\n")
		cancel()
		listener.Close()
	}()

	// Server sends a notification stream when tunnel closes. AcceptStream
	// blocks until the server opens a stream (close notification) or the
	// session dies (network failure). Either way, we shut down.
	go func() {
		stream, err := session.AcceptStream()
		if err != nil {
			if ctx.Err() == nil {
				fmt.Fprintf(f.IO.ErrOut, "Tunnel closed\n")
				cancel()
				listener.Close()
			}
			return
		}
		defer stream.Close()
		// Read the close notification message from server
		if notifyMsg, err := readNgrokMsg(stream); err == nil {
			var closePayload struct {
				Error string `json:"Error"`
			}
			if json.Unmarshal(notifyMsg.Payload, &closePayload) == nil && closePayload.Error != "" {
				fmt.Fprintf(f.IO.ErrOut, "Tunnel closed: %s\n", closePayload.Error)
			} else {
				fmt.Fprintf(f.IO.ErrOut, "Tunnel closed by server\n")
			}
		} else {
			fmt.Fprintf(f.IO.ErrOut, "Tunnel closed\n")
		}
		if ctx.Err() == nil {
			cancel()
			listener.Close()
		}
	}()

	var wg sync.WaitGroup
	for {
		localConn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			fmt.Fprintf(f.IO.ErrOut, "Accept error: %v\n", err)
			continue
		}

		wg.Go(func() {
			if err := forwardViaStream(session, localConn); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "Stream error: %v\n", err)
				if session.IsClosed() {
					cancel()
					listener.Close()
				}
			}
		})
	}

	wg.Wait()
	return nil
}

// dialMuxSession connects to the ngrok server, authenticates, and establishes
// a smux multiplexed session over the TLS connection. The smux session owns
// the underlying TLS connection and will close it when the session closes.
func dialMuxSession(ngrokAddr, tunnelID, token string) (_ *smux.Session, err error) {
	// ngrok tunnel port uses self-signed certificates
	tlsConn, err := tls.Dial("tcp", ngrokAddr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to ngrok: %w", err)
	}
	defer func() {
		if err != nil {
			tlsConn.Close()
		}
	}()

	payload := visitorConnPayload{Mux: true}
	if token != "" {
		payload.Token = token
	} else {
		payload.Key = tunnelID
	}
	if err = writeNgrokMsg(tlsConn, msgNewVisitorConn, &payload); err != nil {
		return nil, fmt.Errorf("visitor conn send: %w", err)
	}

	respMsg, err := readNgrokMsg(tlsConn)
	if err != nil {
		return nil, fmt.Errorf("visitor conn response: %w", err)
	}
	if respMsg.Type != msgNewVisitorConnRes {
		return nil, fmt.Errorf("unexpected response type: %s", respMsg.Type)
	}

	var resp visitorConnRespPayload
	if err = json.Unmarshal(respMsg.Payload, &resp); err != nil {
		return nil, fmt.Errorf("visitor conn response parse: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	smuxCfg := smux.DefaultConfig()
	smuxCfg.KeepAliveInterval = 5 * time.Second
	smuxCfg.KeepAliveTimeout = 15 * time.Second
	session, err := smux.Client(tlsConn, smuxCfg)
	if err != nil {
		return nil, fmt.Errorf("smux session: %w", err)
	}

	return session, nil
}

// forwardViaStream opens a smux stream and bridges it with the local connection.
// Returns an error if the stream cannot be opened (e.g., session closed).
func forwardViaStream(session *smux.Session, localConn net.Conn) error {
	defer localConn.Close()

	stream, err := session.OpenStream()
	if err != nil {
		return err
	}
	defer stream.Close()

	// Each goroutine closes both ends when its direction finishes,
	// which unblocks the other goroutine's io.Copy.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer stream.Close()
		defer localConn.Close()
		io.Copy(stream, localConn)
	}()
	go func() {
		defer wg.Done()
		defer stream.Close()
		defer localConn.Close()
		io.Copy(localConn, stream)
	}()
	wg.Wait()
	return nil
}

// ngrok binary message frame format:
// [8 bytes little-endian int64 length][JSON payload]
// JSON envelope: {"Type":"<msg_type>","Payload":{...}}

type ngrokMsg struct {
	Type    string          `json:"Type"`
	Payload json.RawMessage `json:"Payload"`
}

type visitorConnPayload struct {
	Key   string `json:"Key,omitempty"`
	Token string `json:"Token,omitempty"`
	Mux   bool   `json:"Mux,omitempty"`
}

type visitorConnRespPayload struct {
	Error string `json:"Error"`
}

func writeNgrokMsg(c net.Conn, msgType string, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	envelope := ngrokMsg{Type: msgType, Payload: payloadBytes}
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	var lenBuf [8]byte
	binary.LittleEndian.PutUint64(lenBuf[:], uint64(len(data)))
	if _, err := c.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err = c.Write(data)
	return err
}

func readNgrokMsg(c net.Conn) (*ngrokMsg, error) {
	var lenBuf [8]byte
	if _, err := io.ReadFull(c, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint64(lenBuf[:])
	if length > 1<<20 { // 1MB sanity limit
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(c, data); err != nil {
		return nil, err
	}
	var m ngrokMsg
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
