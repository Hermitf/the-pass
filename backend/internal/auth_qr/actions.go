package authqr

import (
	"context"
)

// mergeMetadata 将给定的 kv 合并进票据的 Metadata
func mergeMetadata(dst map[string]string, src map[string]string) map[string]string {
	// src 为空或为 nil 均直接返回原始目标
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = make(map[string]string, len(src))
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// MarkScanned 将票据状态从 pending 推进到 scanned，写入可选的元数据
func (s *Store) MarkScanned(ctx context.Context, id string, meta map[string]string) (*Ticket, error) {
	return s.UpdateTicket(ctx, id, func(t *Ticket) error {
		switch t.Status {
		case TicketStatusPending:
			t.Status = TicketStatusScanned
			t.Metadata = mergeMetadata(t.Metadata, meta)
			return nil
		case TicketStatusScanned:
			// 幂等：已是 scanned 直接合并元数据
			t.Metadata = mergeMetadata(t.Metadata, meta)
			return nil
		default:
			// 已确认/已拒绝不允许回退
			return ErrTicketExpired
		}
	})
}

// Confirm 将票据状态从 scanned 推进到 confirmed，并绑定用户信息
func (s *Store) Confirm(ctx context.Context, id string, userID int64, userType string, meta map[string]string) (*Ticket, error) {
	return s.UpdateTicket(ctx, id, func(t *Ticket) error {
		if t.Status != TicketStatusScanned {
			return ErrTicketExpired
		}
		t.Status = TicketStatusConfirmed
		t.UserID = userID
		t.UserType = userType
		t.Metadata = mergeMetadata(t.Metadata, meta)
		return nil
	})
}

// Reject 将票据置为 rejected，可附带原因到 Metadata
func (s *Store) Reject(ctx context.Context, id string, reason string, meta map[string]string) (*Ticket, error) {
	return s.UpdateTicket(ctx, id, func(t *Ticket) error {
		switch t.Status {
		case TicketStatusPending, TicketStatusScanned:
			t.Status = TicketStatusRejected
			m := map[string]string{"reject_reason": reason}
			t.Metadata = mergeMetadata(t.Metadata, m)
			t.Metadata = mergeMetadata(t.Metadata, meta)
			return nil
		default:
			return ErrTicketExpired
		}
	})
}
