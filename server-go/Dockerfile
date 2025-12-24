# 使用多阶段构建来减小镜像大小
# 第一阶段：构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的构建依赖
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY server-go/go.mod server-go/go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY server-go/ .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 复制配置文件和数据库初始化脚本
COPY --from=builder /app/.env* . 2>/dev/null || true
COPY --from=builder /app/init.sql . 2>/dev/null || true

# 暴露端口
EXPOSE 8080

# 创建非 root 用户运行应用
RUN adduser -D -s /bin/sh appuser && chown -R appuser:appuser /root/
USER appuser

# 运行应用
CMD ["./server"]