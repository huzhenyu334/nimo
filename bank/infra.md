# 基础设施

## 服务器
- 腾讯云 OpenCloudOS，IP 43.134.86.237
- 2核CPU，7.5GB RAM
- 限制：不能并行编译+E2E测试，CPU会拉满

## 服务
| 服务 | 端口 | 管理方式 |
|------|------|---------|
| PLM | 8080 | 手动 nohup |
| ACP | 3001 | systemd --user |
| Mission Control前端 | - | Docker (已被ACP取代) |
| Mission Control后端 | 8001 | Docker |
| Command Center | 3002 | systemd |

## Mission Control
- 项目：abhi1693/openclaw-mission-control
- 部署目录：/home/claw/.openclaw/workspace/openclaw-mission-control
- Auth token：a90714a6... (LOCAL_AUTH_TOKEN)
- 关键修复：Gateway device identity认证（Ed25519签名）
- Redis端口注意：环境变量REDIS_PORT=6379会覆盖.env的6380

## 技术生态评估
- Antfarm：已卸载，泽斌认为没用（CC原生能做同样的事）
- Moltworker：Cloudflare部署OpenClaw，$10-35/月
- AionUi（⭐16.4K）：多Agent桌面客户端
- Claude Flow（⭐14.1K）：Claude多Agent编排

## 2026-02-24 更新

### Catherine机器配置
- main agent → GLM-5 Coding端点(open.bigmodel.cn/api/coding/paas/v4)
- cat agent → GLM-5，飞书bot "小小白"
- Node跨机器连接：OpenClaw ≥2026.2.17禁止明文ws://，需SSH隧道(28789端口)
- 踩坑：旧paired.json记录冲突需删除、systemd override加gateway token

### memory-lancedb-pro
- 两台机器均已安装（智谱embedding-3, 512维, hybrid模式）

### AI调研结论
- 硬件原理图AI：Flux.ai最成熟（自然语言→原理图→layout）
- AI自主开发：无"一句话出产品"案例，但1人+AI=5人团队模式验证（Vulcan/Every/HumanLayer）
- 核心规律：80%规划+20%执行，给AI验证方式质量提升2-3x
