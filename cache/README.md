# cache
> cache是使用lru4算法实现的KV内存缓存组件,lru4是对lru算法进行淘汰策略变种产生的算法，使用较为简单。

## example
```
  // See more in local_test.go
  key := []byte("k1")
  val := []byte("v1")
  cache := NewLocal()
  cache.Set(key,val)
  val = cache.Get(key)
  cache.Del(key)
  cache.FlushAll()
```

## bench_mark
```
BenchmarkSet-4          10000000           223 ns/op
BenchmarkGet-4          20000000           113 ns/op
BenchmarkDel-4          20000000            76.6 ns/op
BenchmarkFlushAll-4     30000000            57.8 ns/op
```
