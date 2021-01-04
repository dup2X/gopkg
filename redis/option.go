// Package redis defined redis_client
package redis

import (
	"context"
	"time"
)

type option struct {
	// codis cluster name，业务含义上的
	clusterName string
	// key prefix
	prefix string
	// disf 的 serviceName
	serviceName string

	// nodemgr 节点状态更新周期
	workerCycle int
	// nodemgr 健康节点投票阈值
	healthyThreshold int64
	// nodemgr 故障恢复时间
	maxCooldownTime int64
	// nodemgr 最小可用度保护
	minHealthyRatio float64

	// connection pool size
	poolSize int64
	// retry times when err occurred
	maxTryTimes int64
	// max connection-numbers
	maxConn int64
	// ignore errs of initPool
	keepSilent bool

	// Acquire connection mode
	// i) block ii) unbloc iii) block with timeout
	mode         AcquireConnMode
	readTimeout  time.Duration
	writeTimeout time.Duration
	connTimeout  time.Duration
	// wait duration when pool has no available conn
	waitTimeout   time.Duration
	disfEnable    bool
	nodemgrEnable bool

	// Callback : cmd cost err
	statFunc func(ctx context.Context, cmd string, cost time.Duration, err error) error
	slaFuse  bool

	db int
}

// Option 动态参数配置，使 NewManger 支持变参，不设定的参数使用默认值
type Option func(o *option)

//SetStatFunc 是统计命令字符串、时延、redis操作错误的闭包
func SetStatFunc(sFunc func(ctx context.Context, cmd string, cost time.Duration, err error) error) Option {
	return func(o *option) {
		o.statFunc = sFunc
	}
}

// Prefix 是 redis key 默认的 Prefix
func Prefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix
	}
}

func SetDB(db int) Option {
	return func(o *option) {
		o.db = db
	}
}

// SetWaitTimeout 等待连接池中的连接的超时
func SetWaitTimeout(wtTimeout time.Duration) Option {
	return func(o *option) {
		o.waitTimeout = wtTimeout
	}
}

// SetReadTimeout 读取数据超时
func SetReadTimeout(rTimeout time.Duration) Option {
	return func(o *option) {
		o.readTimeout = rTimeout
	}
}

// SetWriteTimeout 写数据超时
func SetWriteTimeout(wTimeout time.Duration) Option {
	return func(o *option) {
		o.writeTimeout = wTimeout
	}
}

// SetConnectTimeout 连接超时
func SetConnectTimeout(cTimeout time.Duration) Option {
	return func(o *option) {
		o.connTimeout = cTimeout
	}
}

// SetPoolSize 设置初始连接池大小
// 叫 initial pool size 更合适
func SetPoolSize(poolSize int64) Option {
	return func(o *option) {
		o.poolSize = poolSize
	}
}

// SetMaxConn 设置连接池内的连接上限
func SetMaxConn(max int64) Option {
	return func(o *option) {
		o.maxConn = max
	}
}

// SetAcquireConnMode 设置获取连接模式：阻塞、超时等待、直接返回
func SetAcquireConnMode(mode AcquireConnMode) Option {
	return func(o *option) {
		o.mode = mode
	}
}

// SetClusterName 设置 codis 集群的名字
func SetClusterName(cn string) Option {
	return func(o *option) {
		o.clusterName = cn
	}
}

// SetKeepSilent 设置静默模式，初始化如果有host连不上则忽略
func SetKeepSilent(silent bool) Option {
	return func(o *option) {
		o.keepSilent = silent
	}
}

// EnableDisf 启用disf
func EnableDisf() Option {
	return func(o *option) {
		o.disfEnable = true
	}
}

// EnableNodemgr 启用nodemgr
func EnableNodemgr() Option {
	return func(o *option) {
		o.nodemgrEnable = true
	}
}

// DisfServiceName 设置disf对应的service name
func DisfServiceName(sn string) Option {
	return func(o *option) {
		o.serviceName = sn
	}
}

// SetWorkerCycle 设置nodemgr 节点状态更新周期
func SetWorkerCycle(cycle int) Option {
	return func(o *option) {
		o.workerCycle = cycle
	}
}

//SetHealthyThreshold 设置nodemgr 健康节点投票阈值
func SetHealthyThreshold(threshold int64) Option {
	return func(o *option) {
		o.healthyThreshold = threshold
	}
}

//SetMaxCooldownTime 设置nodemgr 故障恢复时间
func SetMaxCooldownTime(time int64) Option {
	return func(o *option) {
		o.maxCooldownTime = time
	}
}

//SetMinHealthyRatio 设置nodemgr 最小可用度保护
func SetMinHealthyRatio(ratio float64) Option {
	return func(o *option) {
		o.minHealthyRatio = ratio
	}
}

// SetSLAFuse 设置是否启用SLA熔断
func SetSLAFuse(enable bool) Option {
	return func(o *option) {
		o.slaFuse = enable
	}
}
