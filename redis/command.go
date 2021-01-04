// Package redis defined redis_client
package redis

// redis 命令的封装，防止手写出错
const (
	commandAuth   = "AUTH"
	commandPing   = "PING"
	commandExpire = "EXPIRE"
	commandExists = "EXISTS"
	commandDel    = "DEL"
	commandSet    = "SET"
	commandSetEx  = "SETEX"
	commandSetNx  = "SETNX"
	commandGet    = "GET"
	commandIncr   = "INCR"
	commandIncrBy = "INCRBY"
	commandDecr   = "DECR"
	commandDecrBy = "DECRBY"
	commandMGet   = "MGET"
	commandMSet   = "MSET"

	commandHGet    = "HGET"
	commandHGetAll = "HGETALL"
	commandHSet    = "HSET"
	commandHKeys   = "HKEYS"
	commandHDel    = "HDEL"
	commandHExists = "HEXISTS"
	commandHIncrBy = "HINCRBY"
	commandHMGet   = "HMGET"
	commandHMSet   = "HMSET"
	commandHLen    = "HLEN"

	commandLPush  = "LPUSH"
	commandLPop   = "LPOP"
	commandRPush  = "RPUSH"
	commandRPop   = "RPOP"
	commandBLPop  = "BLPOP"
	commandBRPop  = "BRPOP"
	commandLLen   = "LLEN"
	commandLRange = "LRANGE"
	commandLRem   = "LREM"
	commandLTrim  = "LTRIM"

	commandZAdd             = "ZADD"
	commandZRangeByScore    = "ZRANGEBYSCORE"
	commandZRemRangeByScore = "ZREMRANGEBYSCORE"

	commandSAdd      = "SADD"
	commandSCard     = "SCARD"
	commandSIsMember = "SISMEMBER"
	commandSMembers  = "SMEMBERS"
	commandSRem      = "SREM"
	commandSUnion    = "SUNION"
)
