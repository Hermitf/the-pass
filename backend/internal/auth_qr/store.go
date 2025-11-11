package authqr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	ticketKeyPrefix  = "login:ticket:"
	ticketDefaultTTL = 2 * time.Minute
)

// Store 封装 Redis 操作，用于维护扫码登录票据的生命周期。
type Store struct {
	client *redis.Client
}

// NewStore 初始化票据存储实例。
func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

// ticketKey 统一票据 key 的命名规范，便于集中管理与调试。
func ticketKey(id string) string {
	return ticketKeyPrefix + id
}

// CreateTicket 写入新的票据记录，并返回票据主体。
// 流程概览：
// 1) 计算 TTL（为空则使用默认值）。
// 2) 生成唯一票据 ID，构造初始票据（status=pending）。
// 3) 序列化为 JSON。
// 4) 写入 Redis 并设置过期时间。
// 5) 返回票据主体（调用方可将 ID 编码到二维码中给前端使用）。
func (s *Store) CreateTicket(ctx context.Context, ttl time.Duration) (*Ticket, error) {
	if ttl <= 0 {
		ttl = ticketDefaultTTL
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	ticket := &Ticket{
		ID:        id,
		Status:    TicketStatusPending,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
		UpdatedAt: now,
	}

	payload, err := json.Marshal(ticket)
	if err != nil {
		return nil, fmt.Errorf("marshal ticket failed: %w", err)
	}

	if err := s.client.Set(ctx, ticketKey(id), payload, ttl).Err(); err != nil {
		return nil, fmt.Errorf("store ticket failed: %w", err)
	}

	return ticket, nil
}

// GetTicket 查询票据详情，并校验是否过期。
// 流程概览：
// 1) 从 Redis 读取指定票据。
// 2) 不存在则返回 ErrTicketNotFound。
// 3) 反序列化 JSON 为 Ticket 结构。
// 4) 比较当前时间与 ExpiresAt，过期则返回 ErrTicketExpired。
// 5) 返回票据详情给调用方（前端轮询或后端校验均可复用）。
func (s *Store) GetTicket(ctx context.Context, id string) (*Ticket, error) {
	data, err := s.client.Get(ctx, ticketKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	var ticket Ticket
	if err := json.Unmarshal(data, &ticket); err != nil {
		return nil, fmt.Errorf("unmarshal ticket failed: %w", err)
	}

	if time.Now().UTC().After(ticket.ExpiresAt) {
		return nil, ErrTicketExpired
	}

	return &ticket, nil
}

// UpdateTicket 提供事务安全的票据更新方法，由 mutate 回调更新状态和附加信息。
// 典型用途：
// - 扫码阶段：pending -> scanned，并可记录扫码设备信息到 Metadata。
// - 确认阶段：scanned -> confirmed，并绑定 UserID/UserType。
// - 拒绝/取消：scanned -> rejected（或 pending -> rejected）。
// 流程概览：
// 1) 使用 WATCH 监视 key，先读取当前票据快照。
// 2) 基础校验：不存在/过期直接返回错误。
// 3) 执行业务回调 mutate 更新票据字段（包含状态机校验）。
// 4) 刷新 UpdatedAt，序列化。
// 5) 以剩余 TTL 写回（不延长生命周期，避免“复活”过期票据）。
// 6) 提交事务并返回更新后的票据。
// 并发语义：若有并发写入导致事务失败，Exec 会返回错误，调用方可按需重试。
func (s *Store) UpdateTicket(ctx context.Context, id string, mutate func(t *Ticket) error) (*Ticket, error) {
	key := ticketKey(id)
	var updated *Ticket

	err := s.client.Watch(ctx, func(tx *redis.Tx) error {
		data, err := tx.Get(ctx, key).Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return ErrTicketNotFound
			}
			return err
		}

		var ticket Ticket
		if err := json.Unmarshal(data, &ticket); err != nil {
			return fmt.Errorf("unmarshal ticket failed: %w", err)
		}

		if time.Now().UTC().After(ticket.ExpiresAt) {
			return ErrTicketExpired
		}

		if err := mutate(&ticket); err != nil {
			return err
		}

		ticket.UpdatedAt = time.Now().UTC()
		payload, err := json.Marshal(&ticket)
		if err != nil {
			return fmt.Errorf("marshal ticket failed: %w", err)
		}

		remaining := time.Until(ticket.ExpiresAt)
		if remaining <= 0 {
			return ErrTicketExpired
		}

		pipe := tx.TxPipeline()
		pipe.Set(ctx, key, payload, remaining)
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}

		updated = &ticket
		return nil
	}, key)

	if err != nil {
		return nil, err
	}

	return updated, nil
}

// DeleteTicket 删除票据，用于登录完成或主动关闭流程。
func (s *Store) DeleteTicket(ctx context.Context, id string) error {
	return s.client.Del(ctx, ticketKey(id)).Err()
}
