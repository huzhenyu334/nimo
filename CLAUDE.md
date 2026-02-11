# CLAUDE.md - Project Instructions

## 项目信息
- nimo PLM/ERP系统
- 后端: Go (Gin框架 + GORM + PostgreSQL)
- 前端: React + Ant Design + TypeScript (Vite)
- 服务端口: 8080

## 严格规则
1. **只修改指定的文件**，绝不修改任何其他文件
2. **不要重构**已有代码，只做最小改动修复问题
3. **不要添加新功能**，只修指定的bug
4. 修改前先读懂相关代码，理解现有架构
5. 每次修改后确认编译通过

## 部署步骤（修完代码后执行）
```bash
cd /home/claw/.openclaw/workspace && go build -o bin/plm ./cmd/plm/
cd /home/claw/.openclaw/workspace/nimo-plm-web && npm run build
rm -rf /home/claw/.openclaw/workspace/web/plm/* && cp -r /home/claw/.openclaw/workspace/nimo-plm-web/dist/* /home/claw/.openclaw/workspace/web/plm/
kill $(pgrep -f "bin/plm" | head -1) 2>/dev/null; sleep 1 && cd /home/claw/.openclaw/workspace && nohup ./bin/plm > server.log 2>&1 &
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/
```

## 目录结构
- 后端入口: cmd/plm/main.go
- 后端handler: internal/plm/handler/
- 后端service: internal/plm/service/
- 后端entity: internal/plm/entity/
- 前端页面: nimo-plm-web/src/pages/
- 前端API: nimo-plm-web/src/api/
- 前端路由: nimo-plm-web/src/routes/index.tsx
- 后端测试工具: internal/plm/testutil/
- 后端测试: internal/plm/handler/*_test.go
- 前端E2E测试: nimo-plm-web/e2e/
- 前端Playwright配置: nimo-plm-web/playwright.config.ts

## 测试规则（铁律 — 必须严格遵守）
1. 每次后端代码变更，必须编写或更新对应的 Go test
2. 每次前端代码变更，必须编写或更新对应的 Playwright e2e test
3. 新功能必须有测试覆盖，bug修复必须有回归测试
4. **任务完成前必须自己运行全部测试，测试全部通过才算任务完成**
5. 如果测试失败，必须修复代码直到测试通过，不允许带着失败的测试结束任务

### 任务完成检查清单（每次任务结束前必须执行）
```bash
# Step 1: 编译通过
cd /home/claw/.openclaw/workspace && go build -o bin/plm ./cmd/plm/

# Step 2: 后端测试全部通过
go test ./internal/plm/... -v

# Step 3: 前端编译通过
cd /home/claw/.openclaw/workspace/nimo-plm-web && npm run build

# Step 4: 前端E2E测试全部通过
npx playwright test

# Step 5: 部署并验证服务启动
# （按下方部署步骤执行）
```
**以上5步全部通过后，任务才算完成。任何一步失败都必须修复后重试。**

## 测试命令
```bash
# 后端测试
go test ./internal/plm/... -v

# 前端E2E测试
cd /home/claw/.openclaw/workspace/nimo-plm-web && npx playwright test

# 后端单模块测试
go test ./internal/plm/handler/ -v -run TestRole
go test ./internal/plm/handler/ -v -run TestProject
go test ./internal/plm/handler/ -v -run TestFeishu
```
