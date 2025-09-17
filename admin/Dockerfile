# 构建阶段
#FROM dockerpull.pw/library/node:18-alpine AS builder
FROM docker.1panel.live/library/node:18-alpine AS builder

WORKDIR /app

# 复制package文件
COPY package*.json ./

# 安装依赖
RUN npm config set registry https://registry.npmmirror.com
RUN npm install

# 复制源码
COPY . .

# 构建生产版本
RUN npm run build

# 运行阶段 - 复制文件到volume
#FROM dockerpull.pw/library/alpine:latest
FROM docker.1panel.live/library/alpine:latest

# 复制构建产物
COPY --from=builder /app/build /app/dist

# 直接复制文件到volume并退出
CMD ["sh", "-c", "cp -r /app/dist/* /app/build/ && echo 'Files copied, exiting' && exit 0"]
