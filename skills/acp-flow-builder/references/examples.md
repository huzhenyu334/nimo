# Working Flow Examples

All examples below have been published and validated.

## Example 1: 软件项目立项 (Project Init)

Demonstrates: NocoDB query → switch → agent scan → agent generate → NocoDB update → human review → if gate

```yaml
description: "软件项目立项流程 — 新项目从零开始，老项目补全元数据和蓝图"

steps:
  - id: check_project
    name: "检查项目是否存在"
    executor: nocodb
    input:
      command: search
      table: 软件项目
      where: "(项目名称,eq,{{input}})"

  - id: route
    name: "判断新项目/老项目"
    control: switch
    expression: "{{steps.check_project.output.data.total}}"
    branches:
      "0": new_project
      default: existing_project
    depends_on: [check_project]

  - id: new_project
    name: "新项目立项（待实现）"
    executor: agent
    input:
      prompt: |
        新项目立项占位步骤。项目名称：{{input}}。
        [output:status=pending]
    depends_on: [route]

  - id: existing_project
    name: "读取项目配置"
    executor: nocodb
    input:
      command: search
      table: 软件项目
      where: "(项目名称,eq,{{input}})"
    depends_on: [route]

  - id: code_scan
    name: "代码扫描与分析"
    executor: agent
    input:
      prompt: |
        对项目 **{{input}}** 进行代码扫描。
        项目信息：{{steps.existing_project.output.data.records}}
        
        扫描后端/前端结构，统计API/页面/模块数量。
        [output:module_count=X] [output:api_count=X] [output:page_count=X]
    depends_on: [existing_project]

  - id: generate_blueprint
    name: "生成产品蓝图"
    executor: agent
    input:
      prompt: |
        基于扫描结果为 {{input}} 生成蓝图和需求池。
        扫描结果：{{steps.code_scan.output.text}}
        
        1. 创建蓝图文档到知识库
        2. 创建需求池文档到知识库
        3. 注册到NocoDB文档注册表
        
        [output:blueprint_doc_id=知识库文档ID]
        [output:requirement_pool_doc_id=知识库文档ID]
    depends_on: [code_scan]

  - id: update_project
    name: "更新项目元数据"
    executor: nocodb
    input:
      command: update_record
      table: 软件项目
      record_id: "{{steps.existing_project.output.data.records[0].Id}}"
      fields: |
        {
          "蓝图文档ID": "{{steps.generate_blueprint.output.data.blueprint_doc_id}}",
          "需求池文档ID": "{{steps.generate_blueprint.output.data.requirement_pool_doc_id}}"
        }
    depends_on: [generate_blueprint]

  - id: human_review
    name: "立项确认"
    executor: human
    command: form
    input:
      title: "项目立项确认 — {{input}}"
      assignee: "ou_e229cd56698a8e15e629af2447a8e0ed"
      prompt: |
        模块数：{{steps.code_scan.output.data.module_count}}
        蓝图：{{steps.generate_blueprint.output.data.blueprint_doc_id}}
      form:
        output_schema:
          type: object
          properties:
            approved: { type: boolean, title: "确认立项" }
            notes: { type: string, title: "备注" }
    depends_on: [update_project]

  - id: approval_gate
    name: "立项审批"
    control: if
    condition: "{{steps.human_review.output.data.approved}} == true"
    branches:
      "true": complete
      "false": rejected
    depends_on: [human_review]

  - id: complete
    name: "立项完成"
    executor: nocodb
    input:
      command: update_record
      table: 软件项目
      record_id: "{{steps.existing_project.output.data.records[0].Id}}"
      fields: '{"状态": "进行中"}'
    depends_on: [approval_gate]

  - id: rejected
    name: "立项被拒"
    executor: agent
    input:
      prompt: |
        项目 {{input}} 立项被拒。备注：{{steps.human_review.output.data.notes}}
        [output:status=rejected]
    depends_on: [approval_gate]
```

## Example 2: Simple CC Dev Task

Demonstrates: agent → CC code → shell build → human verify

```yaml
description: "开发任务流程 — Agent分析需求 → CC编码 → 编译 → 人工验收"

steps:
  - id: analyze
    name: "需求分析"
    executor: agent
    input:
      prompt: |
        分析开发任务：{{input}}
        确定要修改的文件、技术方案、测试要点。
        [output:files_to_change=file1,file2]
        [output:approach=简述方案]

  - id: develop
    name: "编码开发"
    executor: cc
    input:
      command: run
      prompt: |
        任务：{{input}}
        方案：{{steps.analyze.output.data.approach}}
        目标文件：{{steps.analyze.output.data.files_to_change}}
        
        完成后 git commit。
      workdir: "/home/claw/.openclaw/workspace"
    depends_on: [analyze]

  - id: build
    name: "编译"
    executor: shell
    input:
      command: "cd /home/claw/.openclaw/workspace && go build -o bin/plm ./cmd/plm/"
      timeout: "5m"
    depends_on: [develop]

  - id: verify
    name: "人工验收"
    executor: human
    command: form
    input:
      title: "验收：{{input}}"
      assignee: "ou_e229cd56698a8e15e629af2447a8e0ed"
      prompt: "编译结果：{{steps.build.output.data.exit_code}}"
      form:
        output_schema:
          type: object
          properties:
            passed: { type: boolean, title: "验收通过" }
    depends_on: [build]
```

## Example 3: Knowledge Base Document Flow

Demonstrates: knowledge search → agent write → nocodb register

```yaml
description: "文档创建流程 — 搜索已有 → 生成文档 → 注册索引"

steps:
  - id: search_existing
    name: "搜索已有文档"
    executor: knowledge
    command: search
    input:
      query: "{{input}}"
      domain: "engineering"
      limit: 5

  - id: generate
    name: "生成文档"
    executor: agent
    input:
      prompt: |
        主题：{{input}}
        已有相关文档：{{steps.search_existing.output.data.contents}}
        
        创建新文档到ACP知识库（acp_knowledge_write），避免重复。
        [output:doc_id=新文档ID]
        [output:doc_title=文档标题]
    depends_on: [search_existing]

  - id: register
    name: "注册到索引"
    executor: nocodb
    input:
      command: create_record
      table: 文档注册表
      fields: |
        {
          "文档标题": "{{steps.generate.output.data.doc_title}}",
          "类型": "技术方案",
          "状态": "草稿",
          "知识库文档ID": "{{steps.generate.output.data.doc_id}}",
          "负责人": "Lyra"
        }
    depends_on: [generate]
```
