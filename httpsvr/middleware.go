// Package httpsvr ...
package httpsvr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/dup2X/gopkg/ctxutil"
	"github.com/dup2X/gopkg/logger"
	"github.com/dup2X/gopkg/metrics"
)

//HandlerFunc 回调方法类型定义
type HandlerFunc func(ctx context.Context, r *http.Request, w http.ResponseWriter)

//Middleware 定义Middleware类型
type Middleware func(ctx context.Context, r *http.Request, w http.ResponseWriter, next HandlerFunc) HandlerFunc

//Recovery 捕获panic的通用处理方法
var Recovery = func(ctx context.Context, r *http.Request, w http.ResponseWriter, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, r *http.Request, w http.ResponseWriter) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server is busy."))
				stack := make([]byte, 2048)
				stack = stack[:runtime.Stack(stack, false)]

				f := "PANIC: %s\n%s"
				logger.Errorf(ctx, logger.DLTagUndefined, f, err, stack)
			}
		}()

		if next != nil {
			next(ctx, r, w)
		}
	}
}

//RecoveryWithMetric 捕获panic，记录metric
var RecoveryWithMetric = func(ctx context.Context, r *http.Request, w http.ResponseWriter, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, r *http.Request, w http.ResponseWriter) {
		defer func() {
			if err := recover(); err != nil {
				metrics.Add("panic", 1)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Server is busy."))
				stack := make([]byte, 1024)
				stack = stack[:runtime.Stack(stack, false)]

				f := "PANIC: %s\n%s"
				logger.Errorf(ctx, logger.DLTagUndefined, f, err, stack)
			}
		}()

		if next != nil {
			next(ctx, r, w)
		}
	}
}

//Access 处理下一个回调方法
var Access = func(ctx context.Context, r *http.Request, w http.ResponseWriter, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, r *http.Request, w http.ResponseWriter) {
		next(ctx, r, w)
	}
}

//Degrade 降级
var Degrade = func(ctx context.Context, r *http.Request, w http.ResponseWriter, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, r *http.Request, w http.ResponseWriter) {
	}
}

// Language ...
var Language = func(ctx context.Context, r *http.Request, w http.ResponseWriter, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, r *http.Request, w http.ResponseWriter) {
		r.ParseForm()
		lang := r.Form.Get(languageKey)
		if lang == "" {
			content := r.Header.Get(hintContentKey)
			v := make(map[string]interface{})
			if err := json.Unmarshal([]byte(content), &v); err != nil {
				logger.Infof(ctx, logger.DLTagUndefined, "_msg=hint_content not json,hintcontent=%s", content)
			}
			if lh, ok := v[languageKey]; ok {
				lang = fmt.Sprint(lh)
			}
		}
		if lang == "" {
			logger.Infof(ctx, logger.DLTagUndefined, "_msg=not found language")
			lang = localLanguage
		}
		// TODO filter invalid lang
		ctx = ctxutil.SetLang(ctx, lang)
		next(ctx, r, w)
	}
}
