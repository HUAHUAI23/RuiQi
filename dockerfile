# 多阶段构建
# 阶段1: 构建Node.js前端
FROM node:23.10.0 AS frontend-builder
# 安装pnpm
RUN npm install -g pnpm@10.6.5
# 设置工作目录
WORKDIR /app
# 复制前端项目文件
COPY server/web/ ./
# 安装依赖并构建前端
RUN pnpm install
RUN pnpm build

# 阶段2: 构建Go后端
FROM golang:1.24.1-alpine AS backend-builder
# 设置Go环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
# 设置工作目录
WORKDIR /build
# 复制整个项目结构
COPY coraza-spoa/ ./coraza-spoa/
COPY pkg/ ./pkg/
COPY server/ ./server/
COPY go.work ./
COPY geo-ip/ ./geo-ip/
# 复制前端构建产物到正确位置
COPY --from=frontend-builder /app/dist ./server/web/dist
# 使用Go的工作区功能进行构建
RUN go work use ./coraza-spoa ./pkg ./server
RUN cd server && go build -o ../simple-waf-server main.go

# 阶段3: 最终镜像 - 使用Ubuntu 24.04并安装HAProxy 3.0
FROM ubuntu:24.04
# 避免交互式前端
ENV DEBIAN_FRONTEND=noninteractive
# 安装HAProxy 3.0
RUN apt-get update && \
    apt-get install -y --no-install-recommends software-properties-common ca-certificates && \
    add-apt-repository -y ppa:vbernat/haproxy-3.0 && \
    apt-get update && \
    apt-get install -y haproxy=3.0.* && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
# 创建必要的目录
WORKDIR /app
# 从构建器复制Go二进制文件
COPY --from=backend-builder /build/simple-waf-server .
# 复制前端构建产物
COPY --from=backend-builder /build/server/web/dist ./web/dist
# 复制Swagger文档文件
COPY --from=backend-builder /build/server/docs/ ./docs/

# 创建root用户家目录下的simple-waf目录并移动geo-ip文件夹进去
RUN mkdir -p /root/simple-waf
COPY --from=backend-builder /build/geo-ip/ /root/simple-waf/geo-ip/

# 设置运行权限
RUN chmod +x /app/simple-waf-server
# 设置环境变量
ENV GIN_MODE=release
# 暴露端口
EXPOSE 2333
# 运行应用
CMD ["/app/simple-waf-server"]