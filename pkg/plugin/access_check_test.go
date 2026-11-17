package plugin

import (
	"context"
	"log/slog"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/certimate-go/certimate/pkg/plugin/proto"
)

type accessCheckPlugin struct {
	*recordingPlugin
	checkImpl  func(ctx context.Context, req *AccessCheckRequest, logger *slog.Logger) (*AccessCheckResult, error)
	checkCalls int
	checkReq   *AccessCheckRequest
}

func (f *accessCheckPlugin) CheckAccess(ctx context.Context, req *AccessCheckRequest, logger *slog.Logger) (*AccessCheckResult, error) {
	f.checkCalls++
	f.checkReq = req
	if f.checkImpl != nil {
		return f.checkImpl(ctx, req, logger)
	}
	return &AccessCheckResult{Status: AccessCheckStatusPassed}, nil
}

func passingAccessCheckResult() *AccessCheckResult {
	return &AccessCheckResult{
		Status:      AccessCheckStatusFailed,
		ReasonCode:  "auth_rejected",
		Items:       []AccessCheckItem{{ID: "connect", Status: AccessCheckItemStatusPassed}, {ID: "auth", Status: AccessCheckItemStatusFailed}},
		Diagnostics: []string{"remote said no"},
	}
}

func TestAccessCheck_RoundTripCarriesFullResult(t *testing.T) {
	impl := &accessCheckPlugin{recordingPlugin: &recordingPlugin{metadata: testMetadata(CapabilityDeploy)}}
	impl.checkImpl = func(context.Context, *AccessCheckRequest, *slog.Logger) (*AccessCheckResult, error) {
		return passingAccessCheckResult(), nil
	}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	res, err := client.CheckAccess(context.Background(), &AccessCheckRequest{AccessConfigJSON: `{"apiKey":"secret-value"}`})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	want := passingAccessCheckResult()
	if res.Status != want.Status || res.ReasonCode != want.ReasonCode || len(res.Items) != 2 || res.Items[1].Status != AccessCheckItemStatusFailed {
		t.Fatalf("result mismatch: %+v", res)
	}
	if len(res.Diagnostics) != 1 || res.Diagnostics[0] != "remote said no" {
		t.Fatalf("diagnostics mismatch: %+v", res.Diagnostics)
	}
	if impl.checkReq.AccessConfigJSON != `{"apiKey":"secret-value"}` {
		t.Fatalf("request payload mismatch: %+v", impl.checkReq)
	}
}

func TestAccessCheck_MetadataFlagFollowsInterface(t *testing.T) {
	withCheck := &accessCheckPlugin{recordingPlugin: &recordingPlugin{metadata: testMetadata(CapabilityDeploy)}}
	client, cleanup := newRoundTrip(t, withCheck)
	defer cleanup()
	meta, err := client.GetMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if !meta.SupportsAccessCheck {
		t.Fatalf("metadata must advertise access check support when the plugin implements AccessChecker: %+v", meta)
	}

	withoutCheck := &recordingPlugin{metadata: testMetadata(CapabilityDeploy)}
	client2, cleanup2 := newRoundTrip(t, withoutCheck)
	defer cleanup2()
	meta2, err := client2.GetMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if meta2.SupportsAccessCheck {
		t.Fatalf("metadata must not advertise access check support without the interface: %+v", meta2)
	}
}

func TestAccessCheck_UnimplementedWithoutInterface(t *testing.T) {
	impl := &recordingPlugin{metadata: testMetadata(CapabilityDeploy)}
	client, cleanup := newRoundTrip(t, impl)
	defer cleanup()

	_, err := client.CheckAccess(context.Background(), &AccessCheckRequest{AccessConfigJSON: `{}`})
	if err == nil {
		t.Fatal("expected unimplemented for plugins without AccessChecker")
	}
	if status.Code(err) != codes.Unimplemented {
		t.Fatalf("expected codes.Unimplemented, got %v", err)
	}
	if impl.processCalls != 0 {
		t.Fatalf("CheckAccess must never fall back to Process, saw %d process calls", impl.processCalls)
	}
}

func TestAccessCheck_ServerRejectsInvalidResults(t *testing.T) {
	cases := []struct {
		name string
		res  *AccessCheckResult
	}{
		{name: "unknown status", res: &AccessCheckResult{Status: "green"}},
		{name: "bad item status", res: &AccessCheckResult{Status: AccessCheckStatusPassed, Items: []AccessCheckItem{{ID: "auth", Status: "maybe"}}}},
		{name: "item without id", res: &AccessCheckResult{Status: AccessCheckStatusPassed, Items: []AccessCheckItem{{ID: "", Status: AccessCheckItemStatusPassed}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			impl := &accessCheckPlugin{recordingPlugin: &recordingPlugin{metadata: testMetadata(CapabilityDeploy)}}
			impl.checkImpl = func(context.Context, *AccessCheckRequest, *slog.Logger) (*AccessCheckResult, error) {
				return tc.res, nil
			}
			client, cleanup := newRoundTrip(t, impl)
			defer cleanup()
			if _, err := client.CheckAccess(context.Background(), &AccessCheckRequest{AccessConfigJSON: `{}`}); err == nil {
				t.Fatalf("expected rejection for %s", tc.name)
			}
		})
	}
}

func TestAccessCheck_ClientRejectsInvalidWireResults(t *testing.T) {
	rogue := &rogueServer{check: func(context.Context, *proto.CheckAccessRequest) (*proto.CheckAccessResponse, error) {
		return &proto.CheckAccessResponse{Status: "green", Items: []*proto.CheckAccessItem{{Id: "auth", Status: "maybe"}}}, nil
	}}
	client, cleanup := newRogueRoundTrip(t, rogue)
	defer cleanup()
	if _, err := client.CheckAccess(context.Background(), &AccessCheckRequest{AccessConfigJSON: `{}`}); err == nil {
		t.Fatal("expected client to reject invalid wire result")
	}
}

func TestAccessCheck_ResultValidate(t *testing.T) {
	if err := (*AccessCheckResult)(nil).Validate(); err == nil {
		t.Fatal("nil result must be rejected")
	}
	if err := (&AccessCheckResult{Status: AccessCheckStatusInconclusive}).Validate(); err != nil {
		t.Fatalf("bare inconclusive result must pass: %v", err)
	}
	if err := (&AccessCheckResult{Status: ""}).Validate(); err == nil {
		t.Fatal("empty status must be rejected")
	}
}
