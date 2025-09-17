#FROM dockerpull.pw/library/golang:1.21-alpine AS builder
FROM docker.1panel.live/library/golang:1.21-alpine AS builder

# 设置Go代理和网络配置
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

WORKDIR /app

COPY . .

# 下载依赖并生成go.sum
RUN go mod tidy -x
RUN go mod verify

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 运行阶段
#FROM dockerpull.pw/library/alpine:latest
FROM docker.1panel.live/library/alpine:latest

WORKDIR /root/

# 从builder阶段复制二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
