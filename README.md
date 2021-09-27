# go-cron

go-cron 是一个分布式cron调度系统, 分为master和worker。

master负责cron的统一管理, 增删改查, 以及任务的日志查看, 强制终止任务, 以及服务发现

worker负责cron任务的变化监控, 调度, 分布式并发执行, 日志管理, 服务注册

节点状态管理使用etcd, 日志存储使用mongdb,

## 安装启动

### 物理机

1, 下载源码

```bash
git clone https://github.com/w1n6/go-cron.git 
```



2, 编译  

```bash
go mod download
```

3, 配置etcd及mongdb

```yaml
common:
  etcd:
    endpoints:
      - 0.0.0.0:23790	#etcd地址
    timeout: 1000 #连接超时 单位毫秒
  mongo:
    uri: mongodb://localhost,localhost:27018,localhost:27019 #这里使用了mongdb集群, 可依据环境自行配制
    timeout: 5000	#连接超时 单位毫秒
    database: cron	#日志数据库
    collection: log	
master:
  host: 0.0.0.0  #监听地址
  port: 8099	#监听端口
worker:
  maxlog: 10	# 当产生10条日志就同步到mongdb, 以提升效率
```



### Docker

1, 打包镜像

```bash
docker build -t go-cron .
```

2, 运行

master

```bash
docker run -d --name go-cron -p 8099:8099 -v ./defau;t.yml:/opt/go-cron/default.yml ./go-cron master -c settings/default.yml 
```

worker

```bash
docker run -d --name go-cron -p 8099:8099 -v ./default.yml:/opt/go-cron/default.yml ./go-cron worker -c settings/default.yml 
```

