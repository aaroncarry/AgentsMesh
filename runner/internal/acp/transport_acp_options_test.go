package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"testing"
)

func TestACPTransportHandshakeSendsInitializedWhenEnabled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	defer stdinR.Close()
	defer stdinW.Close()
	defer stdoutR.Close()
	defer stdoutW.Close()

	transport := NewACPTransportWithOptions(EventCallbacks{}, slog.Default(), ACPTransportOptions{
		SendInitialized: true,
	})
	if err := transport.Initialize(ctx, stdinW, stdoutR, nil); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	go transport.ReadLoop(ctx)

	serverErr := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdinR)

		req, err := readJSONRPCLine(scanner)
		if err != nil {
			serverErr <- err
			return
		}
		if req.Method != "initialize" {
			serverErr <- fmt.Errorf("first method = %q, want initialize", req.Method)
			return
		}
		id, ok := req.GetID()
		if !ok {
			serverErr <- fmt.Errorf("initialize request missing numeric id")
			return
		}
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      &id,
			Result:  json.RawMessage(`{"protocolVersion":1}`),
		}
		if err := writeJSONRPCLine(stdoutW, resp); err != nil {
			serverErr <- err
			return
		}

		notif, err := readJSONRPCLine(scanner)
		if err != nil {
			serverErr <- err
			return
		}
		if notif.Method != "initialized" || notif.ID != nil {
			serverErr <- fmt.Errorf("second message = method %q id %v, want initialized notification", notif.Method, notif.ID)
			return
		}
		serverErr <- nil
	}()

	if _, err := transport.Handshake(ctx); err != nil {
		t.Fatalf("Handshake: %v", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("server: %v", err)
	}
}

func readJSONRPCLine(scanner *bufio.Scanner) (*JSONRPCMessage, error) {
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	var msg JSONRPCMessage
	if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func writeJSONRPCLine(w io.Writer, msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}
