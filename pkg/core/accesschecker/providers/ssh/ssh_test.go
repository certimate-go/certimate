package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
)

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"empty config", map[string]any{}},
		{"missing host", map[string]any{"port": 22, "username": "root", "authMethod": "none"}},
		{"missing username", map[string]any{"host": "127.0.0.1", "port": 22, "authMethod": "none"}},
		{"port out of range", map[string]any{"host": "127.0.0.1", "port": 70000, "username": "root", "authMethod": "none"}},
		{"invalid authMethod", map[string]any{"host": "127.0.0.1", "port": 22, "username": "root", "authMethod": "agent"}},
		{"key method without key", map[string]any{"host": "127.0.0.1", "port": 22, "username": "root", "authMethod": "key"}},
		{"password method without password", map[string]any{"host": "127.0.0.1", "port": 22, "username": "root", "authMethod": "password"}},
		{"malformed port type", map[string]any{"host": "127.0.0.1", "port": "twenty-two", "username": "root", "authMethod": "none"}},
		{"jump server missing host", map[string]any{
			"host": "127.0.0.1", "port": 22, "username": "root", "authMethod": "none",
			"jumpServers": []map[string]any{{"port": 22, "username": "root", "authMethod": "none"}},
		}},
	}

	checker := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
		})
	}
}

func TestChecker_Check_ConnectFailedOnRefusedPort(t *testing.T) {
	addr := closedAddr(t)
	config := map[string]any{
		"host":       addrHost(addr),
		"port":       addrPort(addr),
		"username":   "test",
		"authMethod": "none",
	}

	result, err := (&checker{}).Check(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "failed" || result.ReasonCode != "connect_failed" {
		t.Fatalf("expected failed/connect_failed, got %s/%s", result.Status, result.ReasonCode)
	}
	if result.Items[0].Id != core.AccessCheckItemConnect || result.Items[0].Status != "failed" {
		t.Fatalf("expected connect item failed, got %+v", result.Items)
	}
	if result.Items[1].Id != core.AccessCheckItemAuth || result.Items[1].Status != "skipped" {
		t.Fatalf("expected auth item skipped, got %+v", result.Items)
	}
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	addr := muteSshServer(t)
	config := map[string]any{
		"host":       addrHost(addr),
		"port":       addrPort(addr),
		"username":   "test",
		"authMethod": "none",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	result, err := (&checker{}).Check(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "inconclusive" || result.ReasonCode != "timeout" {
		t.Fatalf("expected inconclusive/timeout, got %s/%s", result.Status, result.ReasonCode)
	}
	for _, item := range result.Items {
		if item.Status != "skipped" {
			t.Fatalf("expected all items skipped on timeout, got %+v", result.Items)
		}
	}
}

func TestChecker_Check_CancelIsInconclusive(t *testing.T) {
	addr := muteSshServer(t)
	config := map[string]any{
		"host":       addrHost(addr),
		"port":       addrPort(addr),
		"username":   "test",
		"authMethod": "none",
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	result, err := (&checker{}).Check(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "inconclusive" || result.ReasonCode != "canceled" {
		t.Fatalf("expected inconclusive/canceled, got %s/%s", result.Status, result.ReasonCode)
	}
}

func TestChecker_Check_InvalidKeyMaterialIsAuthRejected(t *testing.T) {
	addr := bannerSshServer(t)
	config := map[string]any{
		"host":       addrHost(addr),
		"port":       addrPort(addr),
		"username":   "test",
		"authMethod": "key",
		"key":        "-----BEGIN OPENSSH PRIVATE KEY-----\ngarbage\n-----END OPENSSH PRIVATE KEY-----\n",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := (&checker{}).Check(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "failed" || result.ReasonCode != "auth_rejected" {
		t.Fatalf("expected failed/auth_rejected, got %s/%s", result.Status, result.ReasonCode)
	}
	if result.Items[0].Id != core.AccessCheckItemConnect || result.Items[0].Status != "passed" {
		t.Fatalf("expected connect item passed, got %+v", result.Items)
	}
	if result.Items[1].Id != core.AccessCheckItemAuth || result.Items[1].Status != "failed" {
		t.Fatalf("expected auth item failed, got %+v", result.Items)
	}
	if strings.Contains(strings.Join(result.Diagnostics, " "), "garbage") {
		t.Fatalf("diagnostics leaked key material: %v", result.Diagnostics)
	}
}

func TestIsAuthFailure(t *testing.T) {
	authErr := fmt.Errorf("ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain")
	if !isAuthFailure(authErr) {
		t.Fatal("expected auth failure to be recognized")
	}

	otherErr := fmt.Errorf("ssh: handshake failed: ssh: no common algorithm for key exchange")
	if isAuthFailure(otherErr) {
		t.Fatal("expected handshake failure to not be recognized as auth failure")
	}

	if isAuthFailure(nil) {
		t.Fatal("expected nil error to not be an auth failure")
	}
}

func closedAddr(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	return addr
}

func muteSshServer(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.Copy(io.Discard, conn)
			}()
		}
	}()

	return listener.Addr().String()
}

func bannerSshServer(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				conn.Write([]byte("SSH-2.0-OpenSSH_9.6\r\n"))
				io.Copy(io.Discard, conn)
			}()
		}
	}()

	return listener.Addr().String()
}

func addrHost(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

func addrPort(addr string) int32 {
	_, port, _ := net.SplitHostPort(addr)
	var p int32
	fmt.Sscanf(port, "%d", &p)
	return p
}
