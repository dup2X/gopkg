// Package logger ...

// +build !go1.7,!go1.8,!go1.9

package logger

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
)

var ctxStringFunc func(ctx context.Context) string

func init() {
	ctxStringFunc = nilCtxString
}

var nilCtxString = func(ctx context.Context) string { return "ctx_format=unset" }

// RegisterLogger ...
func RegisterLogger(obj Logger) {
	allLog = obj
}

// RegisterContextFormat ...
func RegisterContextFormat(fn func(ctx context.Context) string) {
	ctxStringFunc = fn
}

// Debug ...
func Debug(tag DLTag, args ...interface{}) {
	debug(append([]interface{}{tag, "||"}, args...)...)
}

// Debugf ...
func Debugf(ctx context.Context, tag DLTag, format string, args ...interface{}) {
	debugf("%s||%s||"+format, append([]interface{}{tag, ctxStringFunc(ctx)}, args...)...)
}

// Trace ...
func Trace(tag DLTag, args ...interface{}) {
	trace(append([]interface{}{tag, "||"}, args...)...)
}

// Tracef ...
func Tracef(ctx context.Context, tag DLTag, format string, args ...interface{}) {
	tracef("%s||%s||"+format, append([]interface{}{tag, ctxStringFunc(ctx)}, args...)...)
}

// Info ...
func Info(tag DLTag, args ...interface{}) {
	info(append([]interface{}{tag, "||"}, args...)...)
}

// Infof ...
func Infof(ctx context.Context, tag DLTag, format string, args ...interface{}) {
	infof("%s||%s||"+format, append([]interface{}{tag, ctxStringFunc(ctx)}, args...)...)
}

// Warn ...
func Warn(tag DLTag, args ...interface{}) {
	warn(append([]interface{}{tag, "||"}, args...)...)
}

// Warnf ...
func Warnf(ctx context.Context, tag DLTag, format string, args ...interface{}) {
	warnf("%s||%s||"+format, append([]interface{}{tag, ctxStringFunc(ctx)}, args...)...)
}

// Error ...
func Error(tag DLTag, args ...interface{}) {
	errorc(append([]interface{}{tag, "||"}, args...)...)
}

// Errorf ...
func Errorf(ctx context.Context, tag DLTag, format string, args ...interface{}) {
	errorf("%s||%s||"+format, append([]interface{}{tag, ctxStringFunc(ctx)}, args...)...)
}

// Fatal ...
func Fatal(tag DLTag, args ...interface{}) {
	fatal(append([]interface{}{tag, "||"}, args...)...)
}

// Fatalf ...
func Fatalf(ctx context.Context, tag DLTag, format string, args ...interface{}) {
	fatalf("%s||%s||"+format, append([]interface{}{tag, ctxStringFunc(ctx)}, args...)...)
}

// Public ...
func Public(ctx context.Context, key string, pairs map[string]interface{}) {
	if defaultOfflineFileLog == nil {
		return
	}
	var kvs []string
	kvs = append(kvs, key)
	kvs = append(kvs, "[public]("+genCallInfo(depth-5)+")")
	kvs = append(kvs, "timestamp="+time.Now().Format("2006-01-02 15:04:05"))
	kvs = append(kvs, ctxStringFunc(ctx))
	kvs = append(kvs, "opera_stat_key="+key)
	for k, v := range pairs {
		kvs = append(kvs, k+"="+fmt.Sprint(v))
	}
	defaultOfflineFileLog.track("public", strings.Join(kvs, "||"))
}
