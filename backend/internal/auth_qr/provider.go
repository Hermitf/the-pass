package authqr

import (
	"context"
	"time"
)

// TicketStore 定义扫码登录票据在存储层的最小操作集合。
// 典型调用链：
// 1) Web 端创建票据 CreateTicket，生成二维码供移动端扫码；
// 2) Web 端轮询 GetTicket，读取状态（pending/scanned/confirmed/rejected）；
// 3) 移动端扫码后调用 UpdateTicket 将状态置为 scanned，并可写入 Metadata；
// 4) 移动端确认后调用 UpdateTicket 将状态置为 confirmed，并绑定用户信息；
// 5) 登录完成后，服务端调用 DeleteTicket 清理票据（也可依赖 TTL 自动过期）。
type TicketStore interface {
	CreateTicket(ctx context.Context, ttl time.Duration) (*Ticket, error)
	GetTicket(ctx context.Context, id string) (*Ticket, error)
	UpdateTicket(ctx context.Context, id string, mutate func(t *Ticket) error) (*Ticket, error)
	DeleteTicket(ctx context.Context, id string) error
}

// 编译期断言：保证 Store 满足 TicketStore 接口，避免实现漂移导致的运行时错误。
var _ TicketStore = (*Store)(nil)
