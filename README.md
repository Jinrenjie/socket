# 使用方式

### 打包

#### Linux 平台

```shell
# 在Linux平台下打包到 Mac OS 平台
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o socket main.go
# 在Linux平台下打包到 Windows 平台
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o socket main.go
```
#### Mac OS
```shell
# 在Mac OS平台下打包到 Linux 平台
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o socket main.go
# 在Mac OS平台下打包到 Windows 平台
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o socket main.go
```

### 配置
在可执行文件同级目录创建 `.socket.yaml` 配置文件
```yaml
web:
  host: 127.0.0.1
  port: 8000
  name: George
  pass: George@1994

# 运行时Api前缀
api:
  prefix: /api

# Redis 缓存配置
redis:
  host: 127.0.0.1
  port: 6379
  pass:
  db: 1

# Kafka 配置
kafka:
  broker: 115.28.7.104:9092
  topic: websocketlogs

# Socket 路由
socket:
  prefix: /

# 运行时进程ID
runtime:
  pid: socket.lock

# SSL加密证书
ssl:
  cert: ssl/fullchain.cer
  key: ssl/betterde.com.key

# log
log:
  mongodb:
  host: 127.0.0.1
  port: 27017
  timeout: 600
  username:
  password:
  database: webim
  collection: websocketlogs
```

`web`配置的`name`、和`pass`暂时并未用到，后续如果需要提供管理UI时会用到！

### 管理命令

```shell
Georges-iMac:socket zuber-imac$ socket
Instant Messaging service based on Golang implementation

Usage:
  socket [command]

Available Commands:
  help        Help about any command
  start       Start Web Socket service
  stop        Stop Web Socket service
  version     Display this service version

Flags:
  -h, --help   help for socket

Use "socket [command] --help" for more information about a command.
```
### 注意

redis 的`db`不要与其他业务逻辑数据混用，否则重启服务时将导致数被清空！

### Api

* GET   /api/check/:id              #检测用户是否在线
* GET   /api/connections?id=xxxx    #获取所有连接信息（redis内的数据无法进行分页所以，尽量使用id精确查询），或根据ID获取指定用户的连接信息
* POST  /api/deliver/:id            #投递消息到指定用户ID

关于参数和返回值请参考：[API文档](http://apidoc.zuker.im/#/home/project/inside/api/list?groupID=-1&projectName=IM优化&projectID=14)

### 功能

- [x] 用户连接
- [x] 状态管理
- [x] 消息投递
- [x] SSL支持
- [] 集成`consul`服务管理
- [] 消息存储
- [] 未读消息