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

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
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
Connect to the local port with ssh, telnet, or any TCP client.

Press Ctrl+C to stop.`,
		Example: `  # Forward a tunnel to a random local port
  incloud tunnel forward nhcqr3rzpxqfiviqdu3c7x4o

  # Forward to a specific port
  incloud tunnel forward nhcqr3rzpxqfiviqdu3c7x4o --port 2222

  # Forward with token (required when server has JWT enabled)
  incloud tunnel forward nhcqr3rzpxqfiviqdu3c7x4o --token <jwt>`,
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
	cmd.Flags().IntVar(&ngrokPort, "ngrok-port", 4443, "Ngrok TCP proxy port")

	return cmd
}

// runForward starts a local TCP listener and forwards connections to the
// ngrok tunnel. It blocks until interrupted by SIGINT/SIGTERM.
func runForward(f *factory.Factory, opts *forwardOptions) error {
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
			handleForwardConn(f, localConn, opts.ngrokAddr, opts.tunnelID, opts.token)
		})
	}

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

// handleForwardConn bridges a local TCP connection to the ngrok server
// via the visitor connection protocol.
func handleForwardConn(f *factory.Factory, localConn net.Conn, ngrokAddr, tunnelID, token string) {
	defer localConn.Close()

	// ngrok tunnel port uses self-signed certificates
	remoteConn, err := tls.Dial("tcp", ngrokAddr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Fprintf(f.IO.ErrOut, "Failed to connect to ngrok: %v\n", err)
		return
	}
	defer remoteConn.Close()

	// Send NewVisitorConn message (prefer Token over Key for JWT auth)
	payload := visitorConnPayload{}
	if token != "" {
		payload.Token = token
	} else {
		payload.Key = tunnelID
	}
	if err := writeNgrokMsg(remoteConn, "NewVisitorConn", &payload); err != nil {
		fmt.Fprintf(f.IO.ErrOut, "Visitor conn send failed: %v\n", err)
		return
	}

	respMsg, err := readNgrokMsg(remoteConn)
	if err != nil {
		fmt.Fprintf(f.IO.ErrOut, "Visitor conn response failed: %v\n", err)
		return
	}
	if respMsg.Type != "NewVisitorConnResp" {
		fmt.Fprintf(f.IO.ErrOut, "Unexpected response type: %s\n", respMsg.Type)
		return
	}

	var resp visitorConnRespPayload
	if err := json.Unmarshal(respMsg.Payload, &resp); err != nil {
		fmt.Fprintf(f.IO.ErrOut, "Visitor conn response parse failed: %v\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Fprintf(f.IO.ErrOut, "Visitor conn rejected: %s\n", resp.Error)
		return
	}

	// Bidirectional proxy — connection is now a transparent TCP pipe to the device
	done := make(chan struct{}, 2)
	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	<-done
}
