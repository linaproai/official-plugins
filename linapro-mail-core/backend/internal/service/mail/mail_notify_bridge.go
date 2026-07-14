// This file adapts mail Service to host notifycap.EmailDelivery so notify
// channel_type=email can send through mail-core without knowing SMTP plugins.

package mail

import (
	"context"

	"lina-core/pkg/plugin/capability/notifycap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
)

// notifyEmailBridge adapts Service to notifycap.EmailDelivery.
type notifyEmailBridge struct {
	mailSvc Service
}

// Deliver implements notifycap.EmailDelivery.
func (b notifyEmailBridge) Deliver(ctx context.Context, in notifycap.EmailDeliveryInput) (*notifycap.EmailDeliveryResult, error) {
	if b.mailSvc == nil {
		return nil, nil
	}
	result, err := b.mailSvc.Send(ctx, mailcap.SendInput{
		AccountID: in.AccountID,
		Message: mailcap.MailMessage{
			To:       in.To,
			Subject:  in.Subject,
			TextBody: in.Content,
		},
	})
	if err != nil {
		return nil, err
	}
	out := &notifycap.EmailDeliveryResult{}
	if result != nil {
		out.ProviderMessageID = result.MessageID
	}
	return out, nil
}

// ProvideNotifyEmailDelivery registers this mail service as the host email delivery bridge.
func ProvideNotifyEmailDelivery(mailSvc Service) error {
	return notifycap.ProvideEmailDelivery(notifyEmailBridge{mailSvc: mailSvc})
}
