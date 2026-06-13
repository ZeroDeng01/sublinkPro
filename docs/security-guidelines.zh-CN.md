[English](security-guidelines.md) | 简体中文

# 安全指南

本指南概述了 SublinkPro 开发和部署的安全注意事项、最佳实践和安全检查。

---

## 🔐 默认凭据

### ⚠️ 关键安全警告

**默认管理员凭据**：`admin / 123456`

**必须操作**：
- 首次登录后立即更改默认密码
- 使用强密码且唯一（至少 12 个字符）
- 考虑启用 MFA（多因素认证）

**对于生产部署**：
```bash
# 使用默认凭据登录
# 导航到：设置 → 安全 → 更改密码
# 或使用 API 以编程方式更改密码
```

**切勿**：
- 在生产环境中使用默认凭据
- 共享管理员凭据
- 以明文存储凭据
- 将凭据提交到版本控制

---

## 🔑 敏感配置

### 环境变量

这些变量包含敏感数据，必须受到保护：

#### JWT 密钥
```bash
SUBLINK_JWT_SECRET=<your-secret-key>
```

**用途**：签署用于身份验证的 JWT 令牌
**安全要求**：
- 至少 32 个字符
- 使用加密随机值
- 切勿跨实例重用
- 定期轮换

**生成强密钥**：
```bash
openssl rand -base64 32
```

#### API 加密密钥
```bash
SUBLINK_API_ENCRYPTION_KEY=<your-encryption-key>
```

**用途**：加密敏感 API 数据
**安全要求**：
- 必须恰好 32 字节（AES-256）
- 与 JWT 密钥分开保存
- 安全存储（推荐使用密钥管理器）

**生成**：
```bash
openssl rand -hex 32
```

#### MFA 重置密钥
```bash
SUBLINK_MFA_RESET_SECRET=<your-mfa-secret>
```

**用途**：紧急 MFA 重置功能
**安全要求**：
- 仅用于紧急恢复
- 限制系统管理员访问
- 审计所有使用此密钥的操作
- 每次使用后轮换（推荐）

### 多实例部署

运行多个实例时：

**关键**：所有实例必须使用**相同**的密钥：
- `SUBLINK_JWT_SECRET`
- `SUBLINK_API_ENCRYPTION_KEY`
- `SUBLINK_MFA_RESET_SECRET`

**原因**：JWT 令牌和加密数据必须在实例之间可移植。

**部署模式**：
```bash
# 从安全密钥管理器加载
export SUBLINK_JWT_SECRET=$(vault kv get -field=jwt_secret secret/sublinkpro)
export SUBLINK_API_ENCRYPTION_KEY=$(vault kv get -field=api_key secret/sublinkpro)
export SUBLINK_MFA_RESET_SECRET=$(vault kv get -field=mfa_secret secret/sublinkpro)
```

---

## 🗄️ 数据库安全

### SQLite

**默认位置**：`./db/sublink.db`

**安全注意事项**：
- 文件权限：`600`（仅所有者读写）
- 定期备份
- 静态加密（文件系统级别）
- 不推荐用于多实例部署

**安全权限**：
```bash
chmod 600 db/sublink.db
chown sublinkpro:sublinkpro db/sublink.db
```

### MySQL / PostgreSQL 迁移

**警告**：从 SQLite 迁移到 MySQL/PostgreSQL 需要手动重启。

**迁移检查清单**：
1. ✅ 备份 SQLite 数据库
2. ✅ 配置 MySQL/PostgreSQL 连接
3. ✅ 更新 `DSN` 配置
4. ✅ 重启应用程序
5. ✅ 验证数据完整性
6. ✅ 测试身份验证

**迁移期间的安全**：
- 对数据库连接使用 TLS
- 限制数据库用户权限
- 启用审计日志
- 验证无数据丢失

### 数据库连接安全

```bash
# MySQL 使用 TLS
DSN="username:password@tcp(host:3306)/dbname?tls=true"

# PostgreSQL 使用 SSL
DSN="postgres://username:password@host:5432/dbname?sslmode=require"
```

**切勿**：
- 将数据库端口暴露给公共互联网
- 使用 root 数据库账户
- 在配置文件中存储明文密码
- 在生产环境中禁用 SSL/TLS

---

## 🐳 Docker 安全

### 运行时目录

Docker 挂载这些包含敏感数据的目录：

```yaml
volumes:
  - ./db:/app/db           # 数据库和配置文件
  - ./template:/app/template   # 模板文件
  - ./logs:/app/logs       # 应用日志（可能包含敏感数据）
```

**安全要求**：
- 主机目录：`700` 或 `750` 权限
- 文件：`600` 或 `640` 权限
- 以非 root 用户运行容器
- 尽可能使用只读挂载

**安全的 docker-compose 示例**：
```yaml
version: '3.8'
services:
  sublinkpro:
    image: zerodeng/sublink-pro:latest
    user: "1000:1000"  # 非 root 用户
    volumes:
      - ./db:/app/db:rw
      - ./template:/app/template:ro  # 只读
      - ./logs:/app/logs:rw
    environment:
      - SUBLINK_JWT_SECRET=${SUBLINK_JWT_SECRET}
    secrets:
      - jwt_secret
      - api_key

secrets:
  jwt_secret:
    external: true
  api_key:
    external: true
```

### 容器安全

**最佳实践**：
- 使用特定版本标签，而非 `latest`
- 扫描镜像漏洞
- 启用 Docker 内容信任
- 限制容器资源
- 启用安全选项：

```yaml
security_opt:
  - no-new-privileges:true
  - seccomp:unconfined
read_only: true
tmpfs:
  - /tmp
```

---

## 🔒 MFA（多因素认证）

### 设置安全

**启用 MFA 时**：
- 立即生成备份码
- 安全存储备份码（密码管理器）
- 最终确定前测试 TOTP 代码生成
- 记录恢复过程

**用户操作**：
```
设置 → 安全 → 多因素认证
→ 扫描二维码 → 保存备份码 → 验证 TOTP
```

### 紧急恢复

**MFA 重置密钥**（`SUBLINK_MFA_RESET_SECRET`）：
- 仅用于紧急 MFA 禁用
- 需要系统管理员访问权限
- 审计所有使用
- 每次紧急使用后轮换

**恢复过程**：
1. 通过替代渠道验证用户身份
2. 使用 MFA 重置密钥禁用 MFA
3. 强制密码重置
4. 用户使用新代码重新启用 MFA
5. 轮换 MFA 重置密钥

---

## 🌐 网络安全

### 端口暴露

**默认端口**：`8000`

**建议**：
- 在生产环境中使用反向代理（nginx/Caddy）
- 在代理级别启用 HTTPS/TLS
- 限制对后端端口的直接访问
- 使用防火墙规则

**Nginx 示例**：
```nginx
server {
    listen 443 ssl http2;
    server_name sublink.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### API 安全

**身份验证**：
- 所有 `/api/v1/*` 端点需要身份验证
- 使用 `X-API-Key` 标头或 JWT 令牌
- 订阅端点 `/c/*` 使用基于令牌的认证

**速率限制**（推荐）：
```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

location /api/ {
    limit_req zone=api burst=20 nodelay;
    proxy_pass http://localhost:8000;
}
```

---

## 📝 日志和审计

### 日志安全

**日志位置**：`./logs/`

**安全注意事项**：
- 日志可能包含敏感数据（IP、用户代理、错误）
- 定期轮换日志
- 限制日志文件权限：`640`
- 编辑日志中的敏感数据

**日志轮换**：
```bash
# 设置 logrotate
cat > /etc/logrotate.d/sublinkpro <<EOF
/app/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 640 sublinkpro sublinkpro
}
EOF
```

### 审计事件

**监控**：
- 登录失败尝试
- MFA 更改
- 密码更改
- 管理员权限提升
- 批量数据导出
- 配置更改

---

## 🛡️ 变更安全检查清单

### 部署变更前

- [ ] 无硬编码凭据
- [ ] 日志中无敏感数据
- [ ] 无 SQL 注入漏洞
- [ ] 无 XSS 漏洞
- [ ] 身份验证/授权未更改或改进
- [ ] 存在输入验证
- [ ] 错误消息不泄露敏感信息
- [ ] 密钥使用环境变量
- [ ] 依赖项已更新和扫描
- [ ] 依赖项中无已知 CVE

### 代码审查安全重点

- [ ] 审查身份验证逻辑更改
- [ ] 验证权限检查
- [ ] 检查信息泄露
- [ ] 验证输入清理
- [ ] 确保安全默认值
- [ ] 检查时序攻击
- [ ] 验证加密操作
- [ ] 审查数据库查询注入

---

## 🚨 事件响应

### 安全事件检查清单

1. **识别**：发生了什么？
2. **遏制**：停止漏洞
3. **消除**：移除威胁
4. **恢复**：恢复服务
5. **总结**：记录和改进

### 立即行动

**如果凭据泄露**：
1. 立即轮换所有密钥
2. 强制注销所有会话
3. 审计访问日志
4. 通知受影响用户
5. 审查最近的更改

**如果怀疑数据泄露**：
1. 隔离受影响的系统
2. 保留日志和证据
3. 联系安全团队
4. 记录时间线
5. 准备披露（如需要）

---

## 🔧 安全维护

### 定期任务

**每周**：
- 审查访问日志
- 检查登录失败尝试
- 监控错误率

**每月**：
- 更新依赖项
- 扫描漏洞
- 审查权限
- 测试备份恢复

**每季度**：
- 轮换密钥
- 安全审计
- 渗透测试（如适用）
- 更新安全文档

### 依赖项安全

**扫描依赖项**：
```bash
# 后端
go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

# 前端
cd webs
yarn audit
```

**更新依赖项**：
```bash
# 后端
go get -u ./...
go mod tidy

# 前端
cd webs
yarn upgrade-interactive
```

---

## 📚 相关文档

- **配置**：参见 `configuration.md`
- **MFA 设置**：参见 `docs/features/mfa.md`
- **部署**：参见 `installation.md`
- **Docker 安全**：参见 `skill-sublinkpro/reference/deploy.md`

---

## ⚠️ 已知限制

### 数据库兼容性

**不兼容**上游项目数据库。

不要尝试：
- 从上游项目导入数据
- 在版本之间共享数据库
- 使用不受支持的数据库迁移

### 生产建议

- 生产环境使用 PostgreSQL 或 MySQL（而非 SQLite）
- 启用 HTTPS/TLS
- 对敏感配置使用密钥管理器
- 实施速率限制
- 启用审计日志
- 定期备份
- 监控和告警

---

## 📞 安全联系方式

对于安全漏洞：
- GitHub Security Advisories（首选）
- GitHub Issues（非敏感安全问题）
- 项目维护者（敏感披露）

**请勿**：
- 在补丁发布前公开披露漏洞
- 在生产系统上测试漏洞
- 公开分享利用代码
