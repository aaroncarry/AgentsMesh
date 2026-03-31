# Nginx Frontend Proxy Configuration

## 架构说明

新的部署架构在 Traefik 之前增加了 Nginx 作为前端代理:

```
Client (外部请求)
    ↓
Nginx (前端代理 - 端口 80, 9443)
    ↓
Traefik (内部路由 - 仅内部网络)
    ↓
Backend Services (backend, relay, web, web-admin)
```

## Nginx 的职责

### HTTP 流量 (端口 80)

1. **速率限制 (Rate Limiting)**
   - API 请求: 100 req/s (突发 20)
   - 一般请求: 50 req/s (突发 20)
   - 连接限制: 每 IP 20 并发连接

2. **路由规则**
   - `/api/*` → Traefik → Backend API (更严格的速率限制)
   - `/relay/*` → Traefik → Relay Server (WebSocket, 长连接超时 7 天)
   - `/*` → Traefik → Web Frontend (默认路由)

3. **性能优化**
   - Gzip 压缩
   - 连接复用 (keepalive)
   - 请求缓冲
   - 最大上传大小: 100MB

4. **安全功能**
   - 安全响应头 (X-Frame-Options, X-Content-Type-Options, X-XSS-Protection)
   - 客户端真实 IP 传递 (X-Real-IP, X-Forwarded-For)

### gRPC 流量 (端口 9443)

- **TCP 透传**: gRPC mTLS 流量直接透传到 Traefik → Backend
- **长连接**: 超时设置为 7 天,适合持久化 gRPC stream
- **无 HTTP 处理**: 使用 nginx stream 模块,保证 mTLS 握手直达 Backend

## 配置文件结构

```
nginx/
├── nginx.conf          # 主配置文件
└── README.md          # 本文档
```

## 使用新配置部署

### 方式 1: 使用新的 docker-compose 文件

```bash
cd deploy/selfhost

# 复制环境变量配置
cp .env.example .env

# 编辑 .env 文件,设置 SERVER_HOST 等参数
vim .env

# 使用新的 compose 文件启动
docker compose -f docker-compose-new.yml up -d
```

### 方式 2: 更新现有部署

如果你已经在使用 `docker-compose.yml`,可以:

1. 备份当前配置:
```bash
cp docker-compose.yml docker-compose.yml.backup
```

2. 使用新配置:
```bash
cp docker-compose-new.yml docker-compose.yml
```

3. 如果需要更新 Traefik 配置 (可选):
```bash
cp traefik/traefik-new.yml traefik/traefik.yml
```

4. 重启服务:
```bash
docker compose down
docker compose up -d
```

## 端口说明

| 服务 | 端口 | 说明 |
|-----|------|------|
| Nginx | 80 (外部) | HTTP 入口 |
| Nginx | 9443 (外部) | gRPC mTLS 入口 |
| Traefik | 80 (内部) | 仅 Nginx 可访问 |
| Traefik | 9443 (内部) | 仅 Nginx 可访问 |
| MinIO Console | 9001 (外部) | 可选,管理界面 |
| MinIO API | 9000 (外部) | S3 API |

## 日志查看

```bash
# Nginx 访问日志
docker compose logs -f nginx

# Nginx HTTP 访问日志 (详细)
docker exec agentsmesh-nginx-1 tail -f /var/log/nginx/access.log

# Nginx gRPC 访问日志
docker exec agentsmesh-nginx-1 tail -f /var/log/nginx/grpc_access.log

# Nginx 错误日志
docker exec agentsmesh-nginx-1 tail -f /var/log/nginx/error.log

# Traefik 日志
docker compose logs -f traefik
```

## 健康检查

```bash
# Nginx 健康检查
curl http://localhost/nginx-health

# 完整链路测试
curl http://localhost/health        # → Nginx → Traefik → Backend
```

## 性能调优

### 调整速率限制

编辑 `nginx/nginx.conf`:

```nginx
# 调整 API 速率限制 (默认 100 req/s)
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=200r/s;

# 调整一般请求速率限制 (默认 50 req/s)
limit_req_zone $binary_remote_addr zone=general_limit:10m rate=100r/s;
```

### 调整并发连接数

```nginx
# Worker 连接数 (默认 2048)
events {
    worker_connections 4096;
}

# 单 IP 并发连接限制 (默认 20)
location / {
    limit_conn conn_limit 50;
}
```

### 调整上传大小

```nginx
# 最大上传大小 (默认 100MB)
client_max_body_size 500M;
```

## 故障排查

### 1. 502 Bad Gateway

可能原因:
- Traefik 未启动
- Docker 网络问题

检查:
```bash
docker compose ps traefik
docker compose logs traefik
```

### 2. 速率限制触发 (503 Service Temporarily Unavailable)

检查日志:
```bash
docker exec agentsmesh-nginx-1 grep "limiting requests" /var/log/nginx/error.log
```

临时解决: 调高速率限制或增加 burst 值

### 3. WebSocket 连接断开

检查 `/relay` 路径的超时设置:
```bash
docker exec agentsmesh-nginx-1 nginx -T | grep -A 10 "location /relay"
```

## 安全建议

1. **生产环境启用 HTTPS**: 在 Nginx 中配置 SSL/TLS 证书
2. **限制 MinIO 端口**: 将 MinIO Console (9001) 仅绑定到 localhost
3. **配置防火墙**: 仅开放必要的端口 (80, 443, 9443)
4. **定期更新**: 保持 Nginx 镜像版本更新

## 迁移回原架构

如果需要移除 Nginx,恢复为 Traefik 直接暴露:

```bash
docker compose down
cp docker-compose.yml.backup docker-compose.yml
docker compose up -d
```

