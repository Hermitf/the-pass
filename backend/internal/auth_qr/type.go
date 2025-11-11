// Package authqr 定义扫码登录票据的核心类型和状态常量。
package authqr

import (
	"errors"
	"time"
)

// TicketStatus 表示扫码登录票据的状态机。
// 状态流转：
// pending(初始) -> scanned(已扫码) -> confirmed(已确认) -> (被 Delete 或 TTL 过期回收)
//
//	└-> rejected(拒绝/取消)
//
// 说明：
// 1) confirmed 与 rejected 为终态，前端应停止轮询；
// 2) 过期不显式写入状态，读取时直接返回 ErrTicketExpired；
// 3) 状态推进只允许单向，不支持回退；
// 4) 通过 UpdateTicket + mutate 回调实现原子校验与变更。
type TicketStatus string

const (
	// TicketStatusPending 初始状态，等待用户扫码。
	TicketStatusPending TicketStatus = "pending"
	// TicketStatusScanned 用户已扫码，待移动端确认。
	TicketStatusScanned TicketStatus = "scanned"
	// TicketStatusConfirmed 移动端确认登录，待 PC 端领取。
	TicketStatusConfirmed TicketStatus = "confirmed"
	// TicketStatusRejected 用户拒绝或票据失效。
	TicketStatusRejected TicketStatus = "rejected"
)

var (
	// ErrTicketNotFound 表示票据不存在或已被清理（可能已删除或过期被 Redis 回收）。
	ErrTicketNotFound = errors.New("票据不存在或已失效")
	// ErrTicketExpired 表示票据已过期，不可继续使用（与 NotFound 区分：key 仍存在但逻辑上失效）。
	ErrTicketExpired = errors.New("票据已过期")
)

// Ticket 承载扫码登录票据的状态数据。
type Ticket struct {
	ID        string            `json:"id"`
	Status    TicketStatus      `json:"status"`
	ExpiresAt time.Time         `json:"expires_at"`
	UserID    int64             `json:"user_id,omitempty"`
	UserType  string            `json:"user_type,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AllowPolling 判断票据是否允许 PC 端继续轮询。
func (t Ticket) AllowPolling() bool {
	// 仅在等待扫码或已扫码阶段允许 PC 端持续拉取状态；
	// confirmed/rejected/过期应终止轮询，避免无意义的请求。
	switch t.Status {
	case TicketStatusPending, TicketStatusScanned:
		return true
	default:
		return false
	}
}
