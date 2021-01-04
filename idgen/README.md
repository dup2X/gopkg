# idgen
id生成器，生成uint64的ID和traceID/SpanID。

##  基本算法
id ＝ 39bit时间戳 + ip地址低16bit + 8bit序号

时间戳单位是10ms，序号是每10ms产生256个，所以每秒生成的id有25600个.

## 使用

``` go
    //生成TraceID
    traceID := GenTraceID()
    
    //生成SpanID
    spanID := GenSpanID()
    
    //生成ID
   generator := idgen.New("heheda~")
   id, err := generator.NextID()
```