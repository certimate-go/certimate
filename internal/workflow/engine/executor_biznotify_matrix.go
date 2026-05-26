package engine

import (
	"context"
	"log/slog"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/notifier/providers/matrix"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type matrixSessionAccessRepository interface {
	MergeMatrixSessionIntoAccess(ctx context.Context, accessID, sessionToken, deviceID string) error
}

func (ne *bizNotifyNodeExecutor) persistMatrixSessionIfNeeded(execCtx *NodeExecutionContext, provider string, accessConfig map[string]any) {
	if provider != string(domain.NotificationProviderTypeMatrix) {
		return
	}
	if execCtx == nil || accessConfig == nil {
		return
	}

	credentials := domain.AccessConfigForMatrix{}
	if err := xmaps.Populate(accessConfig, &credentials); err != nil {
		return
	}

	cfg := &matrix.NotifierConfig{
		HomeserverUrl:      credentials.HomeserverUrl,
		AuthMode:           credentials.AuthMode,
		AccessToken:        credentials.AccessToken,
		SessionAccessToken: credentials.SessionAccessToken,
		SessionDeviceId:    credentials.SessionDeviceId,
		UserId:             credentials.UserId,
		Password:           credentials.Password,
		RoomId:             credentials.RoomId,
	}

	sess, ok := matrix.TakePendingSession(cfg)
	if !ok || sess.AccessToken == "" {
		return
	}

	accessID := ""
	if execCtx.Node != nil {
		accessID = execCtx.Node.Data.Config.AsBizNotify().ProviderAccessId
	}
	if accessID == "" {
		return
	}

	saver, ok := ne.accessRepo.(matrixSessionAccessRepository)
	if !ok {
		return
	}
	if err := saver.MergeMatrixSessionIntoAccess(execCtx.Context(), accessID, sess.AccessToken, sess.DeviceID); err != nil {
		ne.logger.Warn(
			"matrix: could not persist session to access",
			slog.String("accessId", accessID),
			slog.String("error", err.Error()),
		)
		return
	}

	ne.logger.Info(
		"matrix: session saved to access credentials",
		slog.String("accessId", accessID),
		slog.String("deviceId", sess.DeviceID),
	)
}
