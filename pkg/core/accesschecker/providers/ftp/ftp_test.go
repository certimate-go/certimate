package ftp

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
)

const ftpPassword = "ftp-pass-secret"

func TestChecker_Check_ConfigInvalid(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
	}{
		{"empty config", map[string]any{}},
		{"missing host", map[string]any{"username": "alice", "password": ftpPassword}},
		{"blank host", map[string]any{"host": "  ", "username": "alice", "password": ftpPassword}},
		{"missing username", map[string]any{"host": "127.0.0.1", "password": ftpPassword}},
		{"missing password", map[string]any{"host": "127.0.0.1", "username": "alice"}},
		{"port out of range", map[string]any{"host": "127.0.0.1", "port": 70000, "username": "alice", "password": ftpPassword}},
		{"malformed port type", map[string]any{"host": "127.0.0.1", "port": "twenty-one", "username": "alice", "password": ftpPassword}},
		{"line break in password", map[string]any{"host": "127.0.0.1", "username": "alice", "password": "bad\r\nPASS"}},
	}

	checker := &checker{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := checker.Check(context.Background(), c.config)
			if !errors.Is(err, core.ErrAccessCheckConfigInvalid) {
				t.Fatalf("expected ErrAccessCheckConfigInvalid, got %v", err)
			}
			if strings.Contains(err.Error(), ftpPassword) {
				t.Fatalf("config error leaked password: %v", err)
			}
		})
	}
}

func TestChecker_Check_LoginPasses(t *testing.T) {
	addr := startFTP(t, true)
	host, port := splitAddr(t, addr)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"host":     host,
		"port":     port,
		"username": "alice",
		"password": ftpPassword,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusPassed, "", core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusPassed)
	assertDiagnostics(t, result, net.JoinHostPort(host, strconv.Itoa(port)), "credential accepted")
	assertNoSecrets(t, result, ftpPassword)
}

func TestChecker_Check_LoginRejected(t *testing.T) {
	addr := startFTP(t, false)
	host, port := splitAddr(t, addr)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"host":     host,
		"port":     port,
		"username": "alice",
		"password": ftpPassword,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonAuthRejected, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusFailed)
	assertDiagnostics(t, result, net.JoinHostPort(host, strconv.Itoa(port)), "login rejected")
	assertNoSecrets(t, result, ftpPassword)
}

func TestChecker_Check_RefusedPort(t *testing.T) {
	addr := closedAddr(t)
	host, port := splitAddr(t, addr)

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"host":     host,
		"port":     port,
		"username": "alice",
		"password": ftpPassword,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, ftpPassword)
}

func TestChecker_Check_OmittedPortUses21(t *testing.T) {
	conn, dialErr := net.DialTimeout("tcp", "127.0.0.1:21", 200*time.Millisecond)
	if dialErr == nil {
		conn.Close()
		t.Skip("127.0.0.1:21 is already accepting connections")
	}

	result, err := (&checker{}).Check(context.Background(), map[string]any{
		"host":     "127.0.0.1",
		"username": "alice",
		"password": ftpPassword,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusFailed, core.AccessCheckReasonConnectFailed, core.AccessCheckItemStatusFailed, core.AccessCheckItemStatusSkipped)
	assertDiagnostics(t, result, "127.0.0.1:21")
	assertNoSecrets(t, result, ftpPassword)
}

func TestChecker_Check_TimeoutIsInconclusive(t *testing.T) {
	addr := startMute(t)
	host, port := splitAddr(t, addr)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	result, err := (&checker{}).Check(ctx, map[string]any{
		"host":     host,
		"port":     port,
		"username": "alice",
		"password": ftpPassword,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertResult(t, result, core.AccessCheckStatusInconclusive, core.AccessCheckReasonTimeout, core.AccessCheckItemStatusSkipped, core.AccessCheckItemStatusSkipped)
	assertNoSecrets(t, result, ftpPassword)
}

func startFTP(t *testing.T, acceptLogin bool) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveFTP(conn, acceptLogin)
		}
	}()

	return ln.Addr().String()
}

func serveFTP(conn net.Conn, acceptLogin bool) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = conn.Write([]byte("220 ready\r\n"))

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "USER "):
			if !acceptLogin {
				_, _ = conn.Write([]byte("530 login incorrect\r\n"))
				return
			}
			_, _ = conn.Write([]byte("331 password required\r\n"))
		case strings.HasPrefix(cmd, "PASS "):
			_, _ = conn.Write([]byte("230 logged in\r\n"))
		case strings.HasPrefix(cmd, "FEAT"):
			_, _ = conn.Write([]byte("502 FEAT not implemented\r\n"))
		case strings.HasPrefix(cmd, "TYPE"):
			_, _ = conn.Write([]byte("200 type set\r\n"))
		case strings.HasPrefix(cmd, "QUIT"):
			_, _ = conn.Write([]byte("221 bye\r\n"))
			return
		default:
			_, _ = conn.Write([]byte("502 not implemented\r\n"))
		}
	}
}

func startMute(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				time.Sleep(30 * time.Second)
			}()
		}
	}()

	return ln.Addr().String()
}

func closedAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func splitAddr(t *testing.T, addr string) (string, int) {
	t.Helper()

	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	return host, port
}

func assertResult(t *testing.T, got *core.AccessCheckResult, wantStatus string, wantReason string, wantConnect string, wantAuth string) {
	t.Helper()

	if got == nil {
		t.Fatal("expected a result, got nil")
	}
	if got.Status != wantStatus {
		t.Errorf("status = %q, want %q", got.Status, wantStatus)
	}
	if got.ReasonCode != wantReason {
		t.Errorf("reasonCode = %q, want %q", got.ReasonCode, wantReason)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items = %+v, want exactly connect and auth", got.Items)
	}
	if got.Items[0].Id != core.AccessCheckItemConnect || got.Items[0].Status != wantConnect {
		t.Errorf("connect item = %+v, want status %q", got.Items[0], wantConnect)
	}
	if got.Items[1].Id != core.AccessCheckItemAuth || got.Items[1].Status != wantAuth {
		t.Errorf("auth item = %+v, want status %q", got.Items[1], wantAuth)
	}
}

func assertDiagnostics(t *testing.T, got *core.AccessCheckResult, mustContain ...string) {
	t.Helper()

	joined := strings.Join(got.Diagnostics, "\n")
	for _, substr := range mustContain {
		if !strings.Contains(joined, substr) {
			t.Errorf("diagnostics missing %q: %v", substr, got.Diagnostics)
		}
	}
}

func assertNoSecrets(t *testing.T, got *core.AccessCheckResult, secrets ...string) {
	t.Helper()

	joined := strings.Join(got.Diagnostics, "\n")
	for _, secret := range secrets {
		if secret != "" && strings.Contains(joined, secret) {
			t.Errorf("diagnostics leaked secret material: %v", got.Diagnostics)
		}
	}
}
