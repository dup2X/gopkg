# config 
配置接入组件

## 目标
提供统一的配置接入接口, 所有项目统一格式和入口，降低沟通和理解成本

## 接口说明
Configer 定义了配置的主要接口，GetSettingXXX：获取指定section下的key的值
Section 定义了一个配置的Section，通常作为一个逻辑上的模块配置，避免一些混淆。
每个模块维护自己section中的配置，避免模块个性化配置被共享。
GetXXXXMust: 获取指定key的value，如果获取失败，则以默认值填充。

## 实现
目前支持类ini配置，简单的toml配置，所有配置建议拆分成kv模式，方便配置管理和服务发现的接入。
复杂格式的配置，会增加管理和扩展成本。

## 举例
test.conf
```
[logger]
type = stdout
level = DEBUG

[read-mysql]
host = 127.0.0.1
port = 3306
db = test
pool_size = 64
```

