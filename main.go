package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type options struct {
	baseURL     string
	path        string
	port        int
	dialTimeout time.Duration
	readTimeout time.Duration
	data        map[string]string
	insecureTLS bool
}

func main() {
	opts, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "argument error: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}

	if err := run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() (options, error) {
	var opts options

	flag.StringVar(&opts.baseURL, "url", "", "WebSocket base URL (e.g. ws://localhost:8080)")
	flag.StringVar(&opts.path, "path", "", "WebSocket path (e.g. /ws)")
	flag.IntVar(&opts.port, "port", 0, "Port to override in the WebSocket URL (optional)")
	flag.DurationVar(&opts.dialTimeout, "dial-timeout", 10*time.Second, "How long to wait when establishing the connection")
	flag.DurationVar(&opts.readTimeout, "read-timeout", 10*time.Second, "How long to wait for responses after sending (0 waits indefinitely)")
	flag.BoolVar(&opts.insecureTLS, "insecure-skip-verify", false, "Skip TLS certificate verification (for wss://; testing only)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s -url ws://host -path /ws [-port 8080] [-insecure-skip-verify] Name=Value [More=Data]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if opts.baseURL == "" {
		return opts, fmt.Errorf("-url is required")
	}
	if opts.path == "" {
		return opts, fmt.Errorf("-path is required")
	}

	opts.data = make(map[string]string)
	for _, arg := range flag.Args() {
		if !strings.Contains(arg, "=") {
			return opts, fmt.Errorf("invalid data %q (want Name=Value)", arg)
		}
		parts := strings.SplitN(arg, "=", 2)
		if strings.TrimSpace(parts[0]) == "" {
			return opts, fmt.Errorf("missing name in %q", arg)
		}
		opts.data[parts[0]] = parts[1]
	}

	return opts, nil
}

func run(opts options) error {
	fullURL, err := buildURL(opts.baseURL, opts.path, opts.port)
	if err != nil {
		return err
	}
	if opts.insecureTLS && !strings.HasPrefix(fullURL, "wss://") {
		return fmt.Errorf("-insecure-skip-verify is only valid with wss:// URLs")
	}

	payload, err := json.Marshal(opts.data)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: opts.dialTimeout,
	}
	if strings.HasPrefix(fullURL, "wss://") {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: opts.insecureTLS} //nolint:gosec // optional override for testing
	}

	conn, resp, err := dialer.Dial(fullURL, nil)
	if err != nil {
		return fmt.Errorf("dial %s: %w", fullURL, err)
	}
	defer conn.Close()

	if resp != nil {
		fmt.Fprintf(os.Stderr, "connected: %s\n", resp.Status)
	}

	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	fmt.Printf("sent: %s\n", payload)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				// The read loop exits on normal close or any read error.
				fmt.Fprintf(os.Stderr, "read finished: %v\n", err)
				return
			}
			printMessage(msg)
		}
	}()

	if opts.readTimeout > 0 {
		select {
		case <-done:
		case <-time.After(opts.readTimeout):
			fmt.Fprintf(os.Stderr, "no more messages within %s; closing connection\n", opts.readTimeout)
			_ = conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "timeout"),
				time.Now().Add(time.Second),
			)
			<-done
		}
	} else {
		<-done
	}

	return nil
}

func buildURL(rawURL, path string, port int) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("url must include scheme, e.g. ws://host or wss://host")
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return "", fmt.Errorf("unsupported scheme %q (use ws:// or wss://)", u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("url must include host")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u.Path = path
	if port > 0 {
		u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(port))
	}
	return u.String(), nil
}

func printMessage(msg []byte) {
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, msg, "", "  "); err == nil {
		fmt.Printf("recv:\n%s\n", formatted.String())
		return
	}
	fmt.Printf("recv: %s\n", msg)
}
