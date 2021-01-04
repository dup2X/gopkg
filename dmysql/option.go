// Package dmysql ...
package dmysql

import (
	"time"

	"github.com/dup2X/gopkg/logger"
)

type connectionOption struct {
	usr              string
	passwd           string
	dbname           string
	charset          string
	loc              string
	columnsWithAlias bool
	autoCommit       bool
	parseTime        bool
	dialTimeout      time.Duration
	readTimeout      time.Duration
	writeTimeout     time.Duration
	debug            bool
	log              logger.Logger
}

type option struct {
	serviceName string
	poolSize    int
	maxConnSize int
	mode        AcquireConnMode
	waitTimeout time.Duration
	copt        *connectionOption
	keepSilent  bool
	disfEnable  bool
	log         logger.Logger
}

// Option option func
type Option func(o *option)

// WithPoolSize set pool size
func WithPoolSize(size int) Option {
	return func(o *option) {
		o.poolSize = size
	}
}

// WithMaxConnSize set max pool size
func WithMaxConnSize(size int) Option {
	return func(o *option) {
		o.maxConnSize = size
	}
}

// WithDialTimeout set conn timeout
func WithDialTimeout(t time.Duration) Option {
	return func(o *option) {
		o.copt.dialTimeout = t
	}
}

// WithReadTimeout set read timeout
func WithReadTimeout(t time.Duration) Option {
	return func(o *option) {
		o.copt.readTimeout = t
	}
}

// WithWriteTimeout set write timeout
func WithWriteTimeout(t time.Duration) Option {
	return func(o *option) {
		o.copt.writeTimeout = t
	}
}

// WithDebug set debug model
func WithDebug(debug bool) Option {
	return func(o *option) {
		o.copt.debug = debug
	}
}

// WithLoc set loc option
func WithLoc(loc string) Option {
	return func(o *option) {
		o.copt.loc = loc
	}
}

// WithColumnsWithAlias set columns
func WithColumnsWithAlias(colWithAlias bool) Option {
	return func(o *option) {
		o.copt.columnsWithAlias = colWithAlias
	}
}

// WithParseTime set parse time
func WithParseTime(parseTime bool) Option {
	return func(o *option) {
		o.copt.parseTime = parseTime
	}
}

// WithAutoCommit set auto_commit
func WithAutoCommit(auto bool) Option {
	return func(o *option) {
		o.copt.autoCommit = auto
	}
}

// WithKeepSilent ignore unusable host when init conn_pool
func WithKeepSilent(silent bool) Option {
	return func(o *option) {
		o.keepSilent = silent
	}
}

// WithLogger set logger
func WithLogger(log logger.Logger) Option {
	return func(o *option) {
		o.log = log
	}
}

// WithDisfServiceName set disf service name and enable disf
func WithDisfServiceName(sn string, disfEnable bool) Option {
	return func(o *option) {
		o.serviceName = sn
		o.disfEnable = disfEnable
	}
}

// WithAcquireConnMode ...
func WithAcquireConnMode(mode AcquireConnMode) Option {
	return func(o *option) {
		o.mode = mode
	}
}
