# redis

## example

```go
    addrs := []string{":6379"}
    auth := ""
    redis := NewManager(addrs, auth, Prefix("test"), SetReadTimeout(time.Second*2),
         SetWriteTimeout(time.Second*2), SetPoolSize(256), SetNamespace("test"))
    redis.Get("aaa")
```

## 多集群使用方法

参考 `_example` 内的 `multi_cluster_example.go`
