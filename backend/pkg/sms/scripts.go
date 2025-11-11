package sms

import "github.com/redis/go-redis/v9"

// ---------- Lua 脚本（集中管理） ----------

// 删除窗口外记录 + 插入当前时间 + 计数 + 设置过期，并返回是否允许
var luaRateLimitScript = redis.NewScript(`
local zkey = KEYS[1]
local nowScore = ARGV[1]
local windowStartScore = ARGV[2]
local maxCount = tonumber(ARGV[3])
local expireSeconds = tonumber(ARGV[4])
redis.call('ZREMRANGEBYSCORE', zkey, '-inf', windowStartScore)
redis.call('ZADD', zkey, nowScore, nowScore)
local count = redis.call('ZCARD', zkey)
redis.call('EXPIRE', zkey, expireSeconds)
if count <= maxCount then
  return {1, count}
else
  return {0, count}
end
`)

// 每日计数自增；若无过期则设置至当天结束
var luaDailyIncrScript = redis.NewScript(`
local dkey = KEYS[1]
local expireSeconds = tonumber(ARGV[1])
local ttl = redis.call('TTL', dkey)
local count = redis.call('INCR', dkey)
if ttl == -1 then
  redis.call('EXPIRE', dkey, expireSeconds)
end
return count
`)

// 只读窗口统计：若可发送返回 {1,0}；若不可发送返回 {0, earliestNs}
var luaPeekRateScript = redis.NewScript(`
local zkey = KEYS[1]
local windowStartScore = ARGV[1]
local maxCount = tonumber(ARGV[2])
local nowScore = tonumber(ARGV[3])
local count = redis.call('ZCOUNT', zkey, windowStartScore, '+inf')
if count < maxCount then
	return {1, 0}
end
local res = redis.call('ZRANGEBYSCORE', zkey, windowStartScore, '+inf', 'WITHSCORES', 'LIMIT', 0, 1)
if res and #res >= 2 then
	local earliest = tonumber(res[2])
	return {0, earliest}
end
return {0, 0}
`)
