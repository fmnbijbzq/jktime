local key = KEYS[1]
local cntKey = key..":cnt"
local val = ARGV[1]

local ttl = tonumber(redis.call("ttl", key))

-- -2是key不存在, ttl < 540 是发送了一个验证码，但是超过了1分钟
if ttl == -2 or ttl < 540 then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    return 0
-- -1 是存在key但是没有过期时间
elseif ttl == -1 then
    return -2
else
    -- 已经发送了一个验证码，但是还不到一分钟
    return -1
end
     
    