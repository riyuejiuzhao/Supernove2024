# 使用官方的Golang镜像作为基础镜像
FROM golang:1.22 as builder

# 设置工作目录
WORKDIR /app

# 将Go模块文件复制到容器中
COPY go.mod ./
COPY go.sum ./

ARG svrName

# 下载依赖
RUN go mod download

# 复制源代码到容器中
COPY . .

# 编译Go应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd/${svrName}.go

# 使用一个新的、更小的基础镜像
FROM alpine:latest

# 将编译好的二进制文件从构建器镜像复制到运行时镜像
COPY --from=builder /app/myapp /app/myapp

# 设置工作目录
WORKDIR /app

# 设置容器启动时运行的命令
CMD ["./myapp"]
