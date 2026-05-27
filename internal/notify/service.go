package notify

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/domain/dtos"
)

const (
	testSubject = "[Certimate] Notification Testing"
	testMessage = "Welcome to use Certimate!"
)

type NotifyService struct {
	accessRepo accessRepository
}

func NewNotifyService(accessRepo accessRepository) *NotifyService {
	return &NotifyService{
		accessRepo: accessRepo,
	}
}

func (n *NotifyService) TestPush(ctx context.Context, req *dtos.NotifyTestPushReq) (*dtos.NotifyTestPushResp, error) {
	accessConfig := make(map[string]any)
	if access, err := n.accessRepo.GetById(ctx, req.AccessId); err != nil {
		return nil, fmt.Errorf("failed to get access #%s record: %w", req.AccessId, err)
	} else {
		if access.Reserve != "notif" {
			return nil, fmt.Errorf("access #%s is not available for notification", req.AccessId)
		}

		accessConfig = access.Config
	}

	notifier := NewClient()
	notifyReq := &SendNotificationRequest{
		Provider:               req.Provider,
		ProviderAccessConfig:   accessConfig,
		ProviderExtendedConfig: make(map[string]any),
		Subject:                testSubject,
		Message:                testMessage,
	}
	if _, err := notifier.SendNotification(ctx, notifyReq); err != nil {
		return nil, err
	}

	return &dtos.NotifyTestPushResp{}, nil
}
