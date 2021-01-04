# httpsvr #
> 添加IDL约束条件, 约束接口的声明为func XXXX(ctx context.Context, req RequestDefine) (resp ResponseDefine, err error)

> 支持自定义的请求解析与相应编码函数

## 几点建议 ##
 - 建议不要使用form, 这种无类型的数据提交方式适合弱类型语言与宽校验接口
 - 敏感或核心接口不要使用GET，原因可以自己查下

 以上建议不采纳导致挖坑的后果需要自己填上。

## How to Use ##
参考_example即可


## FAQ ##
1、为什么要使用这个签名?

前面已经解释，这个共识也是和taowen等各方达成一致的结果。


2、我不想写IDL，如何支持？

自定义输入输出结构体，实现两个方法即可。用不了2分钟


3、为什么不支持标准的HTTP Handler的签名？

参考第一条


4、middleware不够灵活

middleware 用到的场景到底有多少？
这个组件的目标是简单高效可扩展，不是灵活随意搞,所以在各个方面都有限制实现的路径，但是整体支持功能扩展。


5、如果要加特性或者其他怎么搞？

很简单，ServerOption 新增配置即可，对应用方透明



