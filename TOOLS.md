# TOOLS.md - Local Notes

## 编程原则（泽斌 2026-02-07）

**术业有专攻：编程任务必须通过 Claude Code CLI 执行，不要自己直接写代码。**

- 我的角色：理解需求 → 拆解任务 → 调度 Claude Code → 审核结果 → 沟通反馈
- Claude Code 的角色：写代码、改bug、编译测试
- 用法：`bash pty:true workdir:~/project background:true command:"claude 'task description'"`
- Claude Code CLI 版本：2.1.31，路径：/home/claw/n/bin/claude

## 服务器

- nimo PLM: 43.134.86.237:8080
- 代码目录: /home/claw/.openclaw/workspace/nimo-plm
- GitHub: https://github.com/huzhenyu334/nimo-plm

## 技术栈

- 后端: Go (Pure Go)
- 数据库: PostgreSQL
- 前端: 内嵌 web/index.html
- 飞书集成: cli_a9efa9afff78dcb5
