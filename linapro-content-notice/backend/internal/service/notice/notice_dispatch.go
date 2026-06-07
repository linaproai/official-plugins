// notice_dispatch.go implements the publication fan-out helpers that forward
// newly published notices into the host inbox pipeline through the notify
// bridge exposed to source plugins.

package notice

import (
	"context"
	"strconv"

	"lina-core/pkg/plugin/capability/notifycap"
)

// Plugin-local notify category codes. The host notify service treats category
// codes as opaque sender-declared strings; this plugin owns its own category
// vocabulary and registers matching translations under
// `notify.category.{code}.label` / `.color` in its own manifest/i18n bundles.
const (
	// noticeCategoryCodeNotice is the opaque category code for general notice messages dispatched by this plugin.
	noticeCategoryCodeNotice notifycap.CategoryCode = "notice"
	// noticeCategoryCodeAnnouncement is the opaque category code for announcement messages dispatched by this plugin.
	noticeCategoryCodeAnnouncement notifycap.CategoryCode = "announcement"
)

// dispatchPublishedNotice delivers one published notice into the unified inbox
// pipeline after the notice record is persisted.
func (s *serviceImpl) dispatchPublishedNotice(
	ctx context.Context,
	noticeID int64,
	title string,
	content string,
	noticeType int,
	senderUserID int64,
) error {
	_, err := s.notifySvc.Send(ctx, s.capabilityContext(ctx, "notice.publish"), notifycap.SendInput{
		SourceType:   notifycap.SourceTypeNotice,
		SourceID:     strconv.FormatInt(noticeID, 10),
		Title:        title,
		Content:      content,
		Category:     s.noticeTypeToCategoryCode(noticeType),
		SenderUserID: senderUserID,
	})
	return err
}

// noticeTypeToCategoryCode maps plugin-owned notice types to plugin-owned
// notify inbox category codes.
func (s *serviceImpl) noticeTypeToCategoryCode(noticeType int) notifycap.CategoryCode {
	switch noticeType {
	case NoticeTypeAnnouncement:
		return noticeCategoryCodeAnnouncement
	default:
		return noticeCategoryCodeNotice
	}
}
