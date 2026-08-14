local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refillRate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "timestamp")
local tokens = tonumber(data[1])
local lastRefill = tonumber(data[2])

if tokens == nil or lastRefill == nil then
    tokens = capacity
    lastRefill = now
end

if #data > 0 then
    for i = 1, #data, 2 do
        if data[i] == "tokens" then tokens = tonumber(data[i+1]) end
        if data[i] == "timestamp" then lastRefill =  tonumber(data[i+1]) end
    end
end

local elapsed = math.max(0, now - lastRefill)
local refill = elapsed * refillRate

tokens = math.min(capacity, tokens + refill)
lastRefill = now

if tokens < requested then
    redis.call("HSET", key, "tokens", tokens, "timestamp", lastRefill)
    redis.call("EXPIRE", key, math.ceil(capacity / refillRate))
    return {0, math.floor(tokens)}
end

tokens = tokens - requested

local ttl = math.ceil((capacity - tokens) / refillRate)
if ttl < 1 then ttl = 1 end

redis.call("HSET", key, "tokens", tokens, "timestamp", now)
redis.call("EXPIRE", key, ttl)

return {1, math.floor(tokens)}