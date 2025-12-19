# 使用官方 Go 镜像作为构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ipmitool .

# 使用 Alpine 作为运行阶段（更小的镜像）
FROM alpine:latest

# 安装 ipmitool 和必要的运行时依赖
RUN apk add --no-cache \
    ipmitool \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/ipmitool .

# 复制配置文件（如果存在）
COPY config.txt ./

# 更改文件所有者
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./ipmitool"]

