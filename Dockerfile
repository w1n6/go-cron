#构建镜像
FROM golang as builder
#编译路径
WORKDIR /usr/src/go-cron
#设置goproxy代理
ENV GOPROXY=https://goproxy.cn
#复制依赖文件
COPY ./go.mod ./
COPY ./go.sum ./
#下载依赖库
RUN go mod download
#拷贝代码
COPY . .
#编译项目
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o go-cron

#运行环境
FROM busybox as runner
#保证时区一致
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
#复制运行二进制文件及config,settings
COPY --from=builder /usr/src/go-cron /opt/go-cron
#移动到工作目录
WORKDIR /opt/go-cron