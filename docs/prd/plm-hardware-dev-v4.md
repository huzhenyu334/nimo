# plm-hardware-dev 智能眼镜产品开发流程 PRD

> 版本：v4.0 | 日期：2026-03-23
> 隶属：比特幻境企业流程 Master PRD
> 流程文件：`flows/plm-hardware-dev.yaml`
> 触发方式：手动（启动新产品开发项目时）
> **核心范式：步骤编排（Step Orchestration）**

---

## 一、定位与范围

### 1.1 目标

端到端编排智能眼镜硬件产品开发全生命周期。流程本身就是执行框架——每个设计任务、评审、测试都是流程中的一个具体步骤，有明确的 executor、输入和输出。

阶段模型：Planning → EVT → DVT → PVT → MP → Closure

### 1.2 v3 → v4 核心变化

| v3.0（任务清单模式） | v4.0（步骤编排模式） |
|---|---|
| 63 个任务定义在 `variables.phases.tasks` 中 | 每个任务是流程中的具体步骤 |
| `foreach task → pm.create_issue` 批量建任务 | 每步有明确 executor（human/agent/llm） |
| 人工在 PLM 系统中线下执行 | 流程内提交交付物、执行评审 |
| 一个 form 确认"阶段完成了" | 每个交付物有独立提交步骤 |
| 流程 = 项目管理工具 | 流程 = 执行框架 |

### 1.3 步骤设计规则

| 规则 | 说明 |
|------|------|
| **Agent 能做 → agent/llm** | 概念图生成、PRD 分析、测试数据分析、文档撰写 |
| **Agent 做不了 → human/form** | 工程设计提交、物理测试结果、制造数据 |
| **需要审批 → human/approval** | 设计评审、BOM Review、Gate 审批 |
| **系统操作 → plm/srm/shell** | 产品/项目创建、BOM 管理、阶段推进、采购 |
| **每步有 I/O** | 输入来自上游 step output，输出供下游 step 引用 |

### 1.4 智能眼镜特有关注点

| 领域 | 关键挑战 |
|------|--------|
| 光学系统 | 波导片/Micro-LED/LCoS 显示、光机集成、视场角 |
| 轻量化 | 整机 <50g 目标，重量分布与佩戴舒适度 |
| 热管理 | 高密度 SoC + 显示模组的散热方案 |
| 音频 | 骨传导/开放式扬声器、多麦克风阵列 |
| 电池 | 小型化锂电安全（IEC 62133）、续航 |
| 铰链机构 | 折叠寿命、线缆穿越铰链、处方镜片兼容 |
| 法规 | FCC/CE/SRRC/CCC + 激光安全 + 眼安全 |

---

## 二、流程输入

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `approver` | string | * | 项目负责人 / 门审审批人 ID |

---

## 三、流程变量

```yaml
variables:
  workspace: "/home/claw/.openclaw/workspace"
  master_prd_doc_id: "208581f9-c273-498b-89e1-f2d19e3e0c25"
  github_org: "nicobao"
  github_credential: "github-token"
```

> **注意：不再在变量中定义任务列表。** 所有工作项直接作为流程步骤定义。

---

## 四、流程设计

### Phase 0: 项目立项（7 步）

**Step 0.1 — 项目信息收集（human/form）**

```yaml
- id: input_form
  executor: human
  command: form
  input:
    assignee: "{{inputs.approver}}"
    title: 智能眼镜硬件产品立项
    prompt: 创建新的硬件产品开发项目。
    form:
      fields:
        - key: product_name
          label: 产品名称
          type: text
          required: true
        - key: product_code
          label: 产品代号
          type: text
          required: true
        - key: product_description
          label: 产品描述
          type: textarea
          required: true
        - key: prd_doc_id
          label: 产品需求 PRD 文档 ID
          type: text
          required: true
        - key: target_price
          label: 目标 BOM 成本 (USD)
          type: text
        - key: target_weight
          label: 目标重量 (g)
          type: text
        - key: target_mp_date
          label: 目标量产日期
          type: text
          description: "YYYY-MM-DD"
```

**Step 0.2 — 创建 PLM 产品 + 项目（plm）**

```yaml
- id: create_product
  executor: plm
  command: create_product
  depends_on: [input_form]
  input:
    name: "{{steps.input_form.output.form.product_name}}"
    code: "{{steps.input_form.output.form.product_code}}"
    description: "{{steps.input_form.output.form.product_description}}"

- id: create_project
  executor: plm
  command: create_project
  depends_on: [create_product]
  input:
    name: "{{steps.input_form.output.form.product_name}} 硬件开发"
    code: "{{steps.input_form.output.form.product_code}}"
    description: "{{steps.input_form.output.form.product_description}}"
    product_id: "{{steps.create_product.output.product_id}}"
```

> 输出：`create_product.output.product_id`、`create_project.output.project_id` — 后续步骤挂在此项目下。PLM 项目自带阶段模型（planning→evt→dvt→pvt→mp→completed），通过 `advance_project` 推进。

**Step 0.3 — 创建研发主任务（plm）**

```yaml
- id: create_master_task
  executor: plm
  command: create_task
  depends_on: [create_project]
  input:
    project_id: "{{steps.create_project.output.project_id}}"
    title: "{{steps.input_form.output.form.product_name}} 全生命周期开发"
    type: task
    priority: P0
    phase: planning
    description: "产品硬件开发主任务：EVT → DVT → PVT → MP"
```

**Step 0.4 — 读取产品 PRD + 创建 SRM 项目 + PLM 文档（并行）**

```yaml
- id: read_prd
  executor: knowledge
  command: fetch
  depends_on: [input_form]
  input:
    doc_id: "{{steps.input_form.output.form.prd_doc_id}}"

- id: create_srm_project
  executor: srm
  command: create_srm_project
  depends_on: [create_project]
  input:
    name: "{{steps.input_form.output.form.product_name}} 供应链项目"
    type: new_product
    phase: evt
    plm_project_id: "{{steps.create_project.output.project_id}}"

- id: create_prd_doc
  executor: plm
  command: create_document
  depends_on: [create_project]
  input:
    title: "产品 PRD - {{steps.input_form.output.form.product_name}}"
    project_id: "{{steps.create_project.output.project_id}}"
    product_id: "{{steps.create_product.output.product_id}}"
    type: spec
    content: "PRD 知识库文档 ID: {{steps.input_form.output.form.prd_doc_id}}"
```

> PLM 文档系统原生管理产品文档，无需通过 PM 的 add_link 做跨模块关联。

**Step 0.5 — 立项审批（human/approval）**

```yaml
- id: project_approval
  executor: human
  command: approval
  depends_on: [create_project, create_srm_project, read_prd]
  input:
    title: "产品立项审批: {{steps.input_form.output.form.product_name}}"
    description: |
      ## 产品信息
      - 名称: {{steps.input_form.output.form.product_name}}
      - 代号: {{steps.input_form.output.form.product_code}}
      - 目标 BOM: ${{steps.input_form.output.form.target_price}}
      - 目标重量: {{steps.input_form.output.form.target_weight}}g
      - 目标量产: {{steps.input_form.output.form.target_mp_date}}

      ## 产品 PRD 摘要
      {{steps.read_prd.output.contents}}

      请审批是否立项。
    approvers:
      - open_id: "{{inputs.approver}}"
```

---

### Phase 1: EVT — 工程验证测试（16 步）

立项审批通过后，PLM 项目阶段推进到 EVT。各工程团队并行提交设计交付物，汇合后进行联合评审、样机制作、测试、BOM Review、Gate 审批。

**Step 1.0 — 推进项目到 EVT（plm）**

```yaml
- id: advance_to_evt
  executor: plm
  command: advance_project
  depends_on: [project_approval]
  input:
    project_id: "{{steps.create_project.output.project_id}}"
    target_phase: evt
    comment: "立项审批通过，进入 EVT"
```

**Step 1.1 — PRD 需求提取（agent/llm）**

Agent 分析产品 PRD，提取各工程领域的设计约束和目标参数，输出结构化设计要求。

```yaml
- id: evt_prd_analysis
  executor: llm
  command: generate
  depends_on: [advance_to_evt]
  input:
    model: claude-sonnet-4-6
    system: |
      你是硬件产品架构师。从 PRD 中提取各领域的设计约束，输出 JSON 格式。
    prompt: |
      产品: {{steps.input_form.output.form.product_name}}
      目标 BOM: ${{steps.input_form.output.form.target_price}}
      目标重量: {{steps.input_form.output.form.target_weight}}g

      PRD 全文:
      {{steps.read_prd.output.contents}}

      请提取以下领域的设计约束和目标参数：
      1. 工业设计 (ID): 外观风格、尺寸约束、佩戴舒适度要求
      2. 电子 (EE): SoC 方案、通信协议、功耗预算、续航目标
      3. 结构 (ME): 重量分配、材料偏好、铰链要求
      4. 光学: 显示方案、FOV、亮度、分辨率目标
      5. 音频: 扬声器/麦克风方案、降噪需求
      6. CMF: 材料/颜色/表面处理偏好
      7. 法规: 目标市场、必要认证
```

> 输出：`result`（JSON 格式的设计约束，供各团队参考）

**Step 1.2 — 工业设计概念生成（agent/cc）**

Agent 根据 PRD 需求生成概念设计方案描述和要求。实际渲染图由设计团队在专业工具中完成。

```yaml
- id: evt_id_concept_gen
  executor: llm
  command: generate
  depends_on: [evt_prd_analysis]
  input:
    model: claude-sonnet-4-6
    system: |
      你是资深工业设计师。根据产品需求生成 3 套不同风格的智能眼镜概念设计方案。
      每个方案包含：设计理念、外观描述、关键尺寸、材料建议、配色方案、与竞品的差异化。
    prompt: |
      产品: {{steps.input_form.output.form.product_name}}
      目标重量: {{steps.input_form.output.form.target_weight}}g
      设计约束:
      {{steps.evt_prd_analysis.output.result}}

      请输出 3 套概念方案，包含详细的视觉描述，供 3D 建模团队实现。
```

**Step 1.3~1.8 — 各团队设计提交（human/form × 6，并行）**

以下 6 个步骤并行执行，各工程团队独立提交交付物：

```yaml
# [1.3] 工业设计提交
- id: evt_id_submit
  executor: human
  command: form
  depends_on: [evt_id_concept_gen]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT - 工业设计交付物提交"
    prompt: |
      ## Agent 生成的概念方案参考
      {{steps.evt_id_concept_gen.output.result}}

      请基于以上概念方案，上传工业设计交付物。
    form:
      fields:
        - key: concept_count
          label: 概念方案数量
          type: text
          required: true
        - key: selected_concept
          label: 推荐方案编号
          type: text
          required: true
        - key: render_ready
          label: 3D 渲染图完成
          type: select
          required: true
          options:
            - {value: "yes", label: "已完成"}
            - {value: "partial", label: "部分完成"}
        - key: ergonomics_notes
          label: 人机工学评估
          type: textarea
          required: true
          description: "佩戴舒适度、重量分布、镜腿粗细等"
        - key: design_summary
          label: 设计总结与风险
          type: textarea
          required: true

# [1.4] 电子工程提交
- id: evt_ee_submit
  executor: human
  command: form
  depends_on: [evt_prd_analysis]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT - 电子工程交付物提交"
    prompt: |
      请提交 EVT 阶段电子工程设计成果。

      设计约束:
      {{steps.evt_prd_analysis.output.result}}
    form:
      fields:
        - key: schematic_ready
          label: 原理图状态
          type: select
          required: true
          options:
            - {value: "done", label: "完成"}
            - {value: "in_progress", label: "进行中"}
        - key: pcb_ready
          label: PCB Layout 状态
          type: select
          required: true
          options:
            - {value: "done", label: "完成"}
            - {value: "in_progress", label: "进行中"}
        - key: pcb_layers
          label: PCB 层数
          type: text
        - key: power_budget_mw
          label: 功耗预算 (mW)
          type: text
        - key: battery_capacity_mah
          label: 电池容量 (mAh)
          type: text
        - key: estimated_runtime_hrs
          label: 预估续航 (hours)
          type: text
        - key: sensor_list
          label: 传感器选型清单
          type: textarea
          description: "IMU/ToF/环境光/接近 等传感器型号"
        - key: design_summary
          label: 设计总结与风险
          type: textarea
          required: true

# [1.5] 结构工程提交
- id: evt_me_submit
  executor: human
  command: form
  depends_on: [evt_prd_analysis]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT - 结构工程交付物提交"
    prompt: |
      请提交 EVT 阶段结构工程设计成果。
    form:
      fields:
        - key: model_3d_ready
          label: 3D 结构模型状态
          type: select
          required: true
          options:
            - {value: "done", label: "完成"}
            - {value: "in_progress", label: "进行中"}
        - key: hinge_type
          label: 铰链方案
          type: text
          description: "弹簧铰链/阻尼铰链/自定义"
        - key: hinge_lifetime
          label: 铰链寿命目标 (次)
          type: text
        - key: thermal_sim_done
          label: 热仿真完成
          type: select
          required: true
          options:
            - {value: "yes", label: "已完成"}
            - {value: "no", label: "未完成"}
        - key: max_surface_temp
          label: 最高表面温度 (°C)
          type: text
          description: "热仿真结果"
        - key: total_weight_g
          label: 结构件总重量 (g)
          type: text
        - key: prototype_count
          label: 需要手板数量
          type: text
        - key: design_summary
          label: 设计总结与风险
          type: textarea
          required: true

# [1.6] 光学工程提交
- id: evt_optics_submit
  executor: human
  command: form
  depends_on: [evt_prd_analysis]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT - 光学工程交付物提交"
    prompt: 请提交 EVT 阶段光学设计成果。
    form:
      fields:
        - key: optical_module_type
          label: 光机方案
          type: text
          required: true
          description: "LCoS/Micro-LED/DLP + 波导/自由曲面"
        - key: fov_degrees
          label: 视场角 FOV (°)
          type: text
        - key: brightness_nits
          label: 亮度 (nits)
          type: text
        - key: waveguide_type
          label: 波导片方案
          type: text
          description: "衍射/阵列/几何"
        - key: mtf_test_done
          label: MTF 测试完成
          type: select
          options:
            - {value: "yes", label: "已完成"}
            - {value: "no", label: "未完成"}
        - key: design_summary
          label: 设计总结与风险
          type: textarea
          required: true

# [1.7] 音频 + CMF 提交
- id: evt_audio_cmf_submit
  executor: human
  command: form
  depends_on: [evt_prd_analysis]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT - 音频 & CMF 交付物提交"
    prompt: 请提交音频设计和 CMF 选型成果。
    form:
      fields:
        - key: speaker_type
          label: 扬声器方案
          type: text
          description: "骨传导/开放式/其他"
        - key: mic_array
          label: 麦克风阵列方案
          type: text
          description: "数量、布局、降噪方案"
        - key: frame_material
          label: 镜框材料
          type: text
          description: "TR90/钛合金/铝镁/碳纤维"
        - key: surface_treatment
          label: 表面处理
          type: text
          description: "阳极氧化/PVD/喷涂/IMD"
        - key: color_schemes
          label: 配色方案数量
          type: text
        - key: design_summary
          label: 设计总结与风险
          type: textarea
          required: true

# [1.8] 固件提交
- id: evt_firmware_submit
  executor: human
  command: form
  depends_on: [evt_prd_analysis]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT - 固件交付物提交"
    prompt: 请提交 EVT 阶段固件开发状态。
    form:
      fields:
        - key: bsp_status
          label: BSP 移植状态
          type: select
          required: true
          options:
            - {value: "done", label: "完成"}
            - {value: "in_progress", label: "进行中"}
            - {value: "blocked", label: "阻塞"}
        - key: display_driver_status
          label: 显示驱动状态
          type: select
          required: true
          options:
            - {value: "done", label: "完成"}
            - {value: "in_progress", label: "进行中"}
        - key: peripherals_status
          label: 外设驱动状态
          type: textarea
          description: "BT/WiFi/IMU/Touch 等各驱动完成度"
        - key: design_summary
          label: 开发总结与风险
          type: textarea
          required: true
```

**Step 1.9 — EVT 联合设计评审（human/approval）**

所有团队交付物提交后，进行跨职能联合评审。

```yaml
- id: evt_design_review
  executor: human
  command: approval
  depends_on: [evt_id_submit, evt_ee_submit, evt_me_submit, evt_optics_submit, evt_audio_cmf_submit, evt_firmware_submit]
  input:
    title: "EVT 联合设计评审"
    description: |
      ## 工业设计
      方案数: {{steps.evt_id_submit.output.form.concept_count}}
      推荐方案: #{{steps.evt_id_submit.output.form.selected_concept}}
      评估: {{steps.evt_id_submit.output.form.design_summary}}

      ## 电子工程
      原理图: {{steps.evt_ee_submit.output.form.schematic_ready}} | PCB: {{steps.evt_ee_submit.output.form.pcb_ready}}
      功耗: {{steps.evt_ee_submit.output.form.power_budget_mw}}mW | 电池: {{steps.evt_ee_submit.output.form.battery_capacity_mah}}mAh
      续航: {{steps.evt_ee_submit.output.form.estimated_runtime_hrs}}hrs
      评估: {{steps.evt_ee_submit.output.form.design_summary}}

      ## 结构工程
      3D 模型: {{steps.evt_me_submit.output.form.model_3d_ready}}
      铰链: {{steps.evt_me_submit.output.form.hinge_type}} (寿命: {{steps.evt_me_submit.output.form.hinge_lifetime}})
      热仿真: {{steps.evt_me_submit.output.form.thermal_sim_done}} (最高: {{steps.evt_me_submit.output.form.max_surface_temp}}°C)
      重量: {{steps.evt_me_submit.output.form.total_weight_g}}g
      评估: {{steps.evt_me_submit.output.form.design_summary}}

      ## 光学
      方案: {{steps.evt_optics_submit.output.form.optical_module_type}}
      FOV: {{steps.evt_optics_submit.output.form.fov_degrees}}° | 亮度: {{steps.evt_optics_submit.output.form.brightness_nits}}nits
      评估: {{steps.evt_optics_submit.output.form.design_summary}}

      ## 音频 & CMF
      扬声器: {{steps.evt_audio_cmf_submit.output.form.speaker_type}}
      材料: {{steps.evt_audio_cmf_submit.output.form.frame_material}}
      评估: {{steps.evt_audio_cmf_submit.output.form.design_summary}}

      ## 固件
      BSP: {{steps.evt_firmware_submit.output.form.bsp_status}}
      显示驱动: {{steps.evt_firmware_submit.output.form.display_driver_status}}

      ---
      请评审所有设计方案，通过 = 可以进入样机制作，驳回 = 需要修改后重新提交。
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 1.10 — 样机采购（SRM）**

设计评审通过后，创建 EVT 样机物料采购需求。

```yaml
- id: evt_procurement
  executor: srm
  command: create_pr
  depends_on: [evt_design_review]
  input:
    title: "EVT 样机物料采购"
    type: sample
    phase: evt
    project_id: "{{steps.create_srm_project.output.project_id}}"
    priority: high
    description: |
      产品: {{steps.input_form.output.form.product_name}}
      手板数量: {{steps.evt_me_submit.output.form.prototype_count}}
```

**Step 1.11 — 样机制作与测试结果提交（human/form）**

```yaml
- id: evt_test_submit
  executor: human
  command: form
  depends_on: [evt_procurement]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT 样机测试结果提交"
    prompt: |
      样机组装并完成 EVT 功能测试后，请提交测试结果。
    form:
      fields:
        - key: prototype_assembled
          label: 样机组装数量
          type: text
          required: true
        - key: display_test
          label: 显示功能测试
          type: select
          required: true
          options:
            - {value: "pass", label: "通过"}
            - {value: "fail", label: "不通过"}
            - {value: "partial", label: "部分通过"}
        - key: bt_wifi_test
          label: BT/WiFi 连接测试
          type: select
          required: true
          options:
            - {value: "pass", label: "通过"}
            - {value: "fail", label: "不通过"}
        - key: sensor_test
          label: 传感器功能测试
          type: select
          required: true
          options:
            - {value: "pass", label: "通过"}
            - {value: "fail", label: "不通过"}
        - key: audio_test
          label: 音频功能测试
          type: select
          required: true
          options:
            - {value: "pass", label: "通过"}
            - {value: "fail", label: "不通过"}
        - key: charging_test
          label: 充电功能测试
          type: select
          required: true
          options:
            - {value: "pass", label: "通过"}
            - {value: "fail", label: "不通过"}
        - key: esd_test
          label: 电气安全初测
          type: select
          options:
            - {value: "pass", label: "通过"}
            - {value: "fail", label: "不通过"}
            - {value: "na", label: "未测"}
        - key: overall_pass_rate
          label: 总体测试通过率 (%)
          type: text
          required: true
        - key: critical_issues
          label: Critical 问题清单
          type: textarea
        - key: test_summary
          label: 测试总结
          type: textarea
          required: true

```

**Step 1.12 — EVT BOM 编制（human/form）**

```yaml
- id: evt_bom_submit
  executor: human
  command: form
  depends_on: [evt_test_submit]
  input:
    assignee: "{{inputs.approver}}"
    title: "EVT BOM 编制与成本汇总"
    prompt: |
      请编制 EVT 阶段完整 BOM，包含所有物料项的成本。
      目标 BOM 成本: ${{steps.input_form.output.form.target_price}}
    form:
      fields:
        - key: total_bom_cost
          label: BOM 总成本 (USD)
          type: text
          required: true
        - key: bom_item_count
          label: BOM 物料项数
          type: text
          required: true
        - key: top3_cost_items
          label: 成本 TOP3 物料
          type: textarea
          required: true
          description: "物料名称 + 单价 + 占比"
        - key: cost_vs_target
          label: 与目标偏差
          type: text
          description: "百分比，如 +15%"
        - key: cost_reduction_plan
          label: 降本机会点
          type: textarea
```

**Step 1.13 — Agent BOM 分析（llm）**

```yaml
- id: evt_bom_analysis
  executor: llm
  command: generate
  depends_on: [evt_bom_submit]
  input:
    model: claude-sonnet-4-6
    system: |
      你是硬件产品成本分析师。分析 BOM 成本数据，给出优化建议。
    prompt: |
      产品: {{steps.input_form.output.form.product_name}}
      目标 BOM: ${{steps.input_form.output.form.target_price}}
      实际 BOM: ${{steps.evt_bom_submit.output.form.total_bom_cost}}
      偏差: {{steps.evt_bom_submit.output.form.cost_vs_target}}
      成本 TOP3: {{steps.evt_bom_submit.output.form.top3_cost_items}}
      降本机会: {{steps.evt_bom_submit.output.form.cost_reduction_plan}}

      请分析：
      1. 成本结构是否合理
      2. 与目标的差距是否在 EVT 阶段可接受范围（±20%）
      3. 具体降本建议（替代料、方案优化、供应商谈判策略）
```

**Step 1.14 — EVT BOM Review（human/approval）**

```yaml
- id: evt_bom_review
  executor: human
  command: approval
  depends_on: [evt_bom_analysis]
  input:
    title: "EVT BOM Review"
    description: |
      ## BOM 成本
      实际: ${{steps.evt_bom_submit.output.form.total_bom_cost}} | 目标: ${{steps.input_form.output.form.target_price}}
      偏差: {{steps.evt_bom_submit.output.form.cost_vs_target}}
      物料项: {{steps.evt_bom_submit.output.form.bom_item_count}}
      TOP3 成本项: {{steps.evt_bom_submit.output.form.top3_cost_items}}

      ## AI 成本分析
      {{steps.evt_bom_analysis.output.result}}

      请审核 BOM 是否可接受（EVT 允许偏差 ±20%）。
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 1.15 — EVT Gate Review（human/approval）**

```yaml
- id: evt_gate
  executor: human
  command: approval
  depends_on: [evt_bom_review]
  input:
    title: "EVT 阶段门审"
    description: |
      ## 产品: {{steps.input_form.output.form.product_name}}

      ## 关键指标
      | 指标 | 当前值 | 目标 | 状态 |
      |------|--------|------|------|
      | BOM 成本 | ${{steps.evt_bom_submit.output.form.total_bom_cost}} | ${{steps.input_form.output.form.target_price}} (±20%) | {{steps.evt_bom_submit.output.form.cost_vs_target}} |
      | 整机重量 | {{steps.evt_me_submit.output.form.total_weight_g}}g | {{steps.input_form.output.form.target_weight}}g (±15%) | - |
      | 测试通过率 | {{steps.evt_test_submit.output.form.overall_pass_rate}}% | >80% | - |
      | 显示 | {{steps.evt_test_submit.output.form.display_test}} | pass | - |
      | BT/WiFi | {{steps.evt_test_submit.output.form.bt_wifi_test}} | pass | - |

      ## 设计评审结果
      {{steps.evt_design_review.output.result}} - {{steps.evt_design_review.output.comment}}

      ## BOM Review 结果
      {{steps.evt_bom_review.output.result}} - {{steps.evt_bom_review.output.comment}}

      ## 测试总结
      {{steps.evt_test_submit.output.form.test_summary}}

      ## Critical 问题
      {{steps.evt_test_submit.output.form.critical_issues}}

      ---
      **门审标准**: 核心功能可演示、BOM ±20%、重量 ±15%、关键物料已锁定供应商
      通过 = 进入 DVT | 驳回 = 流程暂停
    approvers:
      - open_id: "{{inputs.approver}}"
```

---

### Phase 2: DVT — 设计验证测试（15 步）

EVT Gate 通过后，PLM 项目阶段推进到 DVT。核心目标：设计冻结、开模、可靠性验证、量产供应商定点。

**Step 2.0 — 推进项目到 DVT（plm）**

```yaml
- id: advance_to_dvt
  executor: plm
  command: advance_project
  depends_on: [evt_gate]
  input:
    project_id: "{{steps.create_project.output.project_id}}"
    target_phase: dvt
    comment: "EVT Gate 通过，进入 DVT"
```

**Step 2.1 — 设计冻结签核（human/approval）**

```yaml
- id: dvt_design_freeze
  executor: human
  command: approval
  depends_on: [advance_to_dvt]
  input:
    title: "DVT 设计冻结签核"
    description: |
      EVT 已通过门审。请确认以下设计冻结：
      - 外观设计（ID）冻结
      - 核心电子架构冻结
      - 光学方案冻结
      冻结后任何设计变更须走 ECN 流程。
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 2.2~2.6 — DVT 工程交付（human/form × 5，并行）**

```yaml
# [2.2] CMF 配色打样
- id: dvt_cmf_submit
  executor: human
  command: form
  depends_on: [dvt_design_freeze]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - CMF 配色打样结果"
    prompt: 请提交配色打样和工艺验证结果。
    form:
      fields:
        - key: sku_color_count
          label: SKU 配色数量
          type: text
          required: true
        - key: salt_spray_result
          label: 盐雾测试 (500hr)
          type: select
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: uv_aging_result
          label: UV 老化测试 (100hr)
          type: select
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: summary
          label: CMF 总结
          type: textarea
          required: true

# [2.3] EE 优化 (EMI/ESD/射频)
- id: dvt_ee_submit
  executor: human
  command: form
  depends_on: [dvt_design_freeze]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - 电子优化交付"
    prompt: 请提交 PCB 优化和射频调优结果。
    form:
      fields:
        - key: emi_countermeasures
          label: EMI 对策
          type: textarea
          required: true
        - key: esd_level
          label: ESD 防护等级 (kV)
          type: text
        - key: bt_wifi_antenna_efficiency
          label: 天线效率
          type: text
          description: "百分比"
        - key: sar_pretest
          label: SAR 预测试结果
          type: select
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}, {value: "pending", label: "待测"}]
        - key: summary
          label: 优化总结
          type: textarea
          required: true

# [2.4] 模具设计 + T0/T1
- id: dvt_mold_submit
  executor: human
  command: form
  depends_on: [dvt_design_freeze]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - 模具与试模结果"
    prompt: 请提交模具设计、DFM 评审、T0/T1 试模结果。
    form:
      fields:
        - key: mold_vendor
          label: 模具供应商
          type: text
          required: true
        - key: dfm_done
          label: DFM 评审完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: t0_status
          label: T0 试模状态
          type: select
          required: true
          options: [{value: "done", label: "完成"}, {value: "in_progress", label: "进行中"}, {value: "not_started", label: "未开始"}]
        - key: t1_status
          label: T1 试模状态
          type: select
          options: [{value: "done", label: "完成"}, {value: "in_progress", label: "进行中"}, {value: "not_started", label: "未开始"}]
        - key: critical_dim_cpk
          label: 关键尺寸 CPK
          type: text
          description: "目标 ≥1.0"
        - key: summary
          label: 模具总结与风险
          type: textarea
          required: true

# [2.5] 光学显示精调
- id: dvt_optics_submit
  executor: human
  command: form
  depends_on: [dvt_design_freeze]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - 光学精调结果"
    prompt: 请提交光学系统精调和集成验证结果。
    form:
      fields:
        - key: mtf_optimized
          label: MTF 优化完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: color_calibration
          label: 色彩校准完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: ghost_elimination
          label: 鬼影消除状态
          type: select
          options: [{value: "resolved", label: "已解决"}, {value: "acceptable", label: "可接受"}, {value: "unresolved", label: "未解决"}]
        - key: summary
          label: 光学总结
          type: textarea
          required: true

# [2.6] 固件 Feature Complete + App
- id: dvt_firmware_submit
  executor: human
  command: form
  depends_on: [dvt_design_freeze]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - 固件 & App 交付"
    prompt: 请提交固件 Feature Complete 和手机 App 状态。
    form:
      fields:
        - key: fw_feature_complete
          label: 固件 P0 功能完成度 (%)
          type: text
          required: true
        - key: ota_framework
          label: OTA 框架完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: app_ios_status
          label: iOS App 状态
          type: select
          options: [{value: "done", label: "完成"}, {value: "in_progress", label: "开发中"}, {value: "not_started", label: "未开始"}]
        - key: app_android_status
          label: Android App 状态
          type: select
          options: [{value: "done", label: "完成"}, {value: "in_progress", label: "开发中"}, {value: "not_started", label: "未开始"}]
        - key: summary
          label: 开发总结
          type: textarea
          required: true
```

**Step 2.7~2.9 — DVT 测试（human/form × 3，并行）**

```yaml
# [2.7] 机械可靠性
- id: dvt_reliability_test
  executor: human
  command: form
  depends_on: [dvt_mold_submit, dvt_cmf_submit]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - 机械可靠性测试"
    prompt: 请提交可靠性测试结果。
    form:
      fields:
        - key: drop_test
          label: 跌落测试 (1.5m/6面)
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: hinge_life_test
          label: 铰链寿命测试 (20000次)
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: temp_humidity_test
          label: 温湿度循环测试
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: ip_rating
          label: 防水等级测试
          type: text
          description: "如 IPX4"
        - key: summary
          label: 测试总结
          type: textarea
          required: true

# [2.8] EMC 预测试
- id: dvt_emc_test
  executor: human
  command: form
  depends_on: [dvt_ee_submit]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - EMC 预测试结果"
    prompt: 请提交 EMC 预测试结果。
    form:
      fields:
        - key: fcc_pretest
          label: FCC Part 15 预测试
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "marginal", label: "边缘通过"}, {value: "fail", label: "不通过"}]
        - key: ce_red_pretest
          label: CE RED 预测试
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "marginal", label: "边缘通过"}, {value: "fail", label: "不通过"}]
        - key: rf_test
          label: BT/WiFi 射频指标
          type: select
          options: [{value: "pass", label: "达标"}, {value: "fail", label: "不达标"}]
        - key: deviation_list
          label: 偏差项清单
          type: textarea
        - key: summary
          label: EMC 总结
          type: textarea
          required: true

# [2.9] 法规预认证
- id: dvt_regulatory_pretest
  executor: human
  command: form
  depends_on: [dvt_ee_submit]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - 法规预认证结果"
    prompt: 请提交法规预认证测试状态。
    form:
      fields:
        - key: iec62133_pretest
          label: IEC 62133 电池安全预测试
          type: select
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}, {value: "pending", label: "待测"}]
        - key: iec62471_pretest
          label: IEC 62471 光生物安全预测试
          type: select
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}, {value: "pending", label: "待测"}]
        - key: summary
          label: 法规状态总结
          type: textarea
          required: true
```

**Step 2.10 — 量产供应商定点（human/form + SRM）**

```yaml
- id: dvt_vendor_selection
  executor: human
  command: form
  depends_on: [dvt_mold_submit]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT - 量产供应商定点"
    prompt: 请确认关键物料供应商定点状态。
    form:
      fields:
        - key: soc_vendor
          label: SoC 供应商
          type: text
          required: true
        - key: display_vendor
          label: 显示模组供应商
          type: text
          required: true
        - key: battery_vendor
          label: 电池供应商
          type: text
          required: true
        - key: waveguide_vendor
          label: 波导片供应商
          type: text
        - key: all_vendors_confirmed
          label: 全部关键供应商已确认
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: summary
          label: 供应商总结
          type: textarea
          required: true
```

**Step 2.11~2.12 — DVT BOM + Review**

```yaml
- id: dvt_bom_submit
  executor: human
  command: form
  depends_on: [dvt_vendor_selection, dvt_reliability_test, dvt_emc_test]
  input:
    assignee: "{{inputs.approver}}"
    title: "DVT BOM 更新"
    prompt: |
      请更新 DVT BOM。目标成本: ${{steps.input_form.output.form.target_price}} (±10%)
    form:
      fields:
        - key: total_bom_cost
          label: DVT BOM 总成本 (USD)
          type: text
          required: true
        - key: cost_vs_target
          label: 与目标偏差
          type: text
          required: true
        - key: cost_reduction_items
          label: DVT 降本项
          type: textarea
        - key: alt_materials
          label: 替代料方案
          type: textarea

- id: dvt_bom_review
  executor: human
  command: approval
  depends_on: [dvt_bom_submit]
  input:
    title: "DVT BOM Review"
    description: |
      DVT BOM: ${{steps.dvt_bom_submit.output.form.total_bom_cost}}
      目标: ${{steps.input_form.output.form.target_price}} (±10%)
      偏差: {{steps.dvt_bom_submit.output.form.cost_vs_target}}
      降本项: {{steps.dvt_bom_submit.output.form.cost_reduction_items}}
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 2.13~2.14 — DVT Gate**

```yaml
- id: dvt_gate
  executor: human
  command: approval
  depends_on: [dvt_bom_review, dvt_reliability_test, dvt_emc_test, dvt_regulatory_pretest, dvt_optics_submit, dvt_firmware_submit]
  input:
    title: "DVT 阶段门审"
    description: |
      ## 产品: {{steps.input_form.output.form.product_name}}

      ## 关键指标
      | 指标 | 结果 | 标准 |
      |------|------|------|
      | 设计冻结 | {{steps.dvt_design_freeze.output.result}} | approved |
      | BOM 成本 | ${{steps.dvt_bom_submit.output.form.total_bom_cost}} | ±10% |
      | 跌落测试 | {{steps.dvt_reliability_test.output.form.drop_test}} | pass |
      | 铰链寿命 | {{steps.dvt_reliability_test.output.form.hinge_life_test}} | pass |
      | 温湿度 | {{steps.dvt_reliability_test.output.form.temp_humidity_test}} | pass |
      | FCC 预测试 | {{steps.dvt_emc_test.output.form.fcc_pretest}} | pass |
      | CE RED 预测试 | {{steps.dvt_emc_test.output.form.ce_red_pretest}} | pass |
      | 模具 T1 CPK | {{steps.dvt_mold_submit.output.form.critical_dim_cpk}} | ≥1.0 |
      | 固件完成度 | {{steps.dvt_firmware_submit.output.form.fw_feature_complete}}% | 100% |
      | 供应商定点 | {{steps.dvt_vendor_selection.output.form.all_vendors_confirmed}} | yes |

      通过 = 进入 PVT | 驳回 = 流程暂停
    approvers:
      - open_id: "{{inputs.approver}}"
```

---

### Phase 3: PVT — 生产验证测试（13 步）

DVT Gate 通过后，PLM 项目阶段推进到 PVT。核心目标：产线搭建、试产验证、法规认证、BOM 成本锁定。

**Step 3.0 — 推进项目到 PVT（plm）**

```yaml
- id: advance_to_pvt
  executor: plm
  command: advance_project
  depends_on: [dvt_gate]
  input:
    project_id: "{{steps.create_project.output.project_id}}"
    target_phase: pvt
    comment: "DVT Gate 通过，进入 PVT"
```

**Step 3.1~3.4 — PVT 生产准备（human/form × 4，并行）**

```yaml
# [3.1] 模具最终确认
- id: pvt_mold_final
  executor: human
  command: form
  depends_on: [advance_to_pvt]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT - 模具 T2 最终确认"
    prompt: 请提交 T2 试模结果和模具验收状态。
    form:
      fields:
        - key: t2_cpk
          label: T2 关键尺寸 CPK
          type: text
          required: true
          description: "目标 ≥1.33"
        - key: appearance_pass_rate
          label: 外观合格率 (%)
          type: text
          required: true
          description: "目标 ≥98%"
        - key: mold_accepted
          label: 模具验收签核
          type: select
          required: true
          options: [{value: "yes", label: "已签核"}, {value: "no", label: "未签核"}]
        - key: summary
          label: 模具总结
          type: textarea
          required: true

# [3.2] 产线测试站搭建
- id: pvt_test_station
  executor: human
  command: form
  depends_on: [advance_to_pvt]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT - 产线测试站搭建"
    prompt: 请提交产线测试站搭建状态。
    form:
      fields:
        - key: stations_list
          label: 测试工站清单
          type: textarea
          required: true
          description: "SMT → 组装 → 校准 → 功能测试 → 老化 → OQC"
        - key: takt_time_seconds
          label: 节拍时间 (秒)
          type: text
        - key: test_program_ready
          label: 测试程序开发完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: aql_plan
          label: AQL 抽样方案
          type: textarea
        - key: summary
          label: 产线总结
          type: textarea
          required: true

# [3.3] PFMEA
- id: pvt_pfmea
  executor: human
  command: form
  depends_on: [advance_to_pvt]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT - 过程 PFMEA"
    prompt: 请提交 PFMEA 分析结果。
    form:
      fields:
        - key: high_rpn_count
          label: 高 RPN 项数量
          type: text
          required: true
        - key: control_plan_ready
          label: 控制计划完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: top_risks
          label: TOP3 风险项
          type: textarea
          required: true

# [3.4] OTA 系统验证
- id: pvt_ota_test
  executor: human
  command: form
  depends_on: [advance_to_pvt]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT - OTA 系统验证"
    prompt: 请提交 OTA 全链路测试结果。
    form:
      fields:
        - key: diff_upgrade
          label: 差分升级
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: full_upgrade
          label: 全量升级
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: rollback
          label: 回滚测试
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: batch_stress
          label: 批量升级压力测试
          type: select
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: summary
          label: OTA 总结
          type: textarea
          required: true
```

**Step 3.5 — PVT 试产（human/form）**

```yaml
- id: pvt_trial_production
  executor: human
  command: form
  depends_on: [pvt_mold_final, pvt_test_station, pvt_pfmea]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT - 试产结果"
    prompt: 请提交 PVT 试产数据。
    form:
      fields:
        - key: trial_quantity
          label: 试产数量
          type: text
          required: true
        - key: yield_rate
          label: 良率 (%)
          type: text
          required: true
          description: "目标 ≥95%"
        - key: top3_defects
          label: TOP3 不良项
          type: textarea
          required: true
        - key: spc_cpk
          label: 关键工序 CPK
          type: text
          description: "目标 ≥1.33"
        - key: summary
          label: 试产总结
          type: textarea
          required: true
```

**Step 3.6~3.7 — 法规认证提交 + 包装（并行）**

```yaml
# [3.6] 法规认证正式送审
- id: pvt_regulatory_submit
  executor: human
  command: form
  depends_on: [pvt_trial_production]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT - 法规认证送审状态"
    prompt: 请更新法规认证送审进度。
    form:
      fields:
        - key: fcc_status
          label: FCC Part 15
          type: select
          required: true
          options: [{value: "submitted", label: "已送审"}, {value: "passed", label: "已通过"}, {value: "pending", label: "待送审"}]
        - key: ce_status
          label: CE RED
          type: select
          required: true
          options: [{value: "submitted", label: "已送审"}, {value: "passed", label: "已通过"}, {value: "pending", label: "待送审"}]
        - key: iec62133_status
          label: IEC 62133 电池安全
          type: select
          required: true
          options: [{value: "submitted", label: "已送审"}, {value: "passed", label: "已通过"}, {value: "pending", label: "待送审"}]
        - key: ccc_status
          label: CCC (中国市场)
          type: select
          options: [{value: "submitted", label: "已送审"}, {value: "passed", label: "已通过"}, {value: "na", label: "不适用"}]
        - key: summary
          label: 认证总结
          type: textarea
          required: true

# [3.7] 包装设计
- id: pvt_packaging
  executor: human
  command: form
  depends_on: [pvt_trial_production]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT - 包装设计验证"
    prompt: 请提交包装设计和测试结果。
    form:
      fields:
        - key: package_drop_test
          label: 包装跌落测试 (ISTA 2A)
          type: select
          required: true
          options: [{value: "pass", label: "通过"}, {value: "fail", label: "不通过"}]
        - key: package_bom_ready
          label: 包装 BOM 确认
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: print_files_ready
          label: 印刷文件定稿
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
```

**Step 3.8 — 量产物料采购（SRM）**

```yaml
- id: pvt_mass_procurement
  executor: srm
  command: create_pr
  depends_on: [pvt_trial_production]
  input:
    title: "PVT/MP 量产物料采购"
    type: standard
    phase: pvt
    project_id: "{{steps.create_srm_project.output.project_id}}"
    priority: high
```

**Step 3.9~3.10 — PVT BOM 锁定 + Review**

```yaml
- id: pvt_bom_submit
  executor: human
  command: form
  depends_on: [pvt_trial_production]
  input:
    assignee: "{{inputs.approver}}"
    title: "PVT BOM 成本锁定"
    prompt: |
      请提交 PVT 最终 BOM。目标成本: ${{steps.input_form.output.form.target_price}} (±5%)
    form:
      fields:
        - key: total_bom_cost
          label: PVT BOM 总成本 (USD)
          type: text
          required: true
        - key: cost_vs_target
          label: 与目标偏差
          type: text
          required: true
        - key: all_cost_items_locked
          label: 所有物料价格已锁定
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]

- id: pvt_bom_review
  executor: human
  command: approval
  depends_on: [pvt_bom_submit]
  input:
    title: "PVT BOM Review - 成本锁定"
    description: |
      PVT BOM: ${{steps.pvt_bom_submit.output.form.total_bom_cost}}
      目标: ${{steps.input_form.output.form.target_price}} (±5%)
      偏差: {{steps.pvt_bom_submit.output.form.cost_vs_target}}
      价格锁定: {{steps.pvt_bom_submit.output.form.all_cost_items_locked}}
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 3.11~3.12 — PVT Gate**

```yaml
- id: pvt_gate
  executor: human
  command: approval
  depends_on: [pvt_bom_review, pvt_regulatory_submit, pvt_packaging, pvt_ota_test]
  input:
    title: "PVT 阶段门审"
    description: |
      ## 产品: {{steps.input_form.output.form.product_name}}

      | 指标 | 结果 | 标准 |
      |------|------|------|
      | 试产良率 | {{steps.pvt_trial_production.output.form.yield_rate}}% | ≥95% |
      | BOM 成本 | ${{steps.pvt_bom_submit.output.form.total_bom_cost}} | ±5% |
      | 模具 CPK | {{steps.pvt_mold_final.output.form.t2_cpk}} | ≥1.33 |
      | FCC | {{steps.pvt_regulatory_submit.output.form.fcc_status}} | submitted+ |
      | CE | {{steps.pvt_regulatory_submit.output.form.ce_status}} | submitted+ |
      | IEC 62133 | {{steps.pvt_regulatory_submit.output.form.iec62133_status}} | submitted+ |
      | 包装跌落 | {{steps.pvt_packaging.output.form.package_drop_test}} | pass |
      | OTA 验证 | 差分:{{steps.pvt_ota_test.output.form.diff_upgrade}} 回滚:{{steps.pvt_ota_test.output.form.rollback}} | pass |

      ## 试产问题
      {{steps.pvt_trial_production.output.form.top3_defects}}

      通过 = 进入 MP | 驳回 = 流程暂停
    approvers:
      - open_id: "{{inputs.approver}}"
```

---

### Phase 4: MP — 量产（9 步）

PVT Gate 通过后，PLM 项目阶段推进到 MP。核心目标：产线爬坡、Golden Sample、法规证书确认、产品发布。

**Step 4.0 — 推进项目到 MP（plm）**

```yaml
- id: advance_to_mp
  executor: plm
  command: advance_project
  depends_on: [pvt_gate]
  input:
    project_id: "{{steps.create_project.output.project_id}}"
    target_phase: mp
    comment: "PVT Gate 通过，进入 MP"
```

**Step 4.1 — 产线爬坡（human/form）**

```yaml
- id: mp_ramp
  executor: human
  command: form
  depends_on: [advance_to_mp]
  input:
    assignee: "{{inputs.approver}}"
    title: "MP - 产线爬坡结果"
    prompt: 请提交产线爬坡数据。
    form:
      fields:
        - key: target_daily_capacity
          label: 目标日产能
          type: text
          required: true
        - key: actual_daily_capacity
          label: 实际日产能
          type: text
          required: true
        - key: ramp_days
          label: 爬坡天数
          type: text
        - key: consecutive_yield_5day
          label: 连续 5 天良率 (%)
          type: text
          required: true
          description: "目标 ≥95%"
        - key: sop_training_done
          label: SOP 培训完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: summary
          label: 爬坡总结
          type: textarea
          required: true
```

**Step 4.2 — Golden Sample 签样（human/approval）**

```yaml
- id: mp_golden_sample
  executor: human
  command: approval
  depends_on: [mp_ramp]
  input:
    title: "Golden Sample 签样确认"
    description: |
      产品: {{steps.input_form.output.form.product_name}}
      日产能: {{steps.mp_ramp.output.form.actual_daily_capacity}}
      连续 5 天良率: {{steps.mp_ramp.output.form.consecutive_yield_5day}}%

      Golden Sample 需通过外观、功能、可靠性全面验证。
      确认签样 = 作为量产标准样。
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 4.3 — 认证证书确认（human/approval）**

```yaml
- id: mp_cert_confirm
  executor: human
  command: approval
  depends_on: [advance_to_mp]
  input:
    title: "法规认证证书确认"
    description: |
      请确认所有目标市场法规认证证书已获取或预计在发货前获取：
      - FCC: {{steps.pvt_regulatory_submit.output.form.fcc_status}}
      - CE: {{steps.pvt_regulatory_submit.output.form.ce_status}}
      - IEC 62133: {{steps.pvt_regulatory_submit.output.form.iec62133_status}}
      - CCC: {{steps.pvt_regulatory_submit.output.form.ccc_status}}
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 4.4~4.5 — 发布准备（human/form × 2，并行）**

```yaml
# [4.4] 市场发布准备
- id: mp_launch_prep
  executor: human
  command: form
  depends_on: [mp_golden_sample]
  input:
    assignee: "{{inputs.approver}}"
    title: "MP - 产品发布准备"
    prompt: 请确认产品发布准备状态。
    form:
      fields:
        - key: marketing_materials
          label: 宣传物料完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: user_manual
          label: 使用手册完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: compliance_labels
          label: 合规标签/铭牌确认
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]

# [4.5] 售后准备
- id: mp_aftersales_prep
  executor: human
  command: form
  depends_on: [mp_golden_sample]
  input:
    assignee: "{{inputs.approver}}"
    title: "MP - 售后准备"
    prompt: 请确认售后服务准备状态。
    form:
      fields:
        - key: repair_manual
          label: 维修手册完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: spare_parts_list
          label: 备件清单完成
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
        - key: rma_process
          label: RMA 流程建立
          type: select
          required: true
          options: [{value: "yes", label: "是"}, {value: "no", label: "否"}]
```

**Step 4.6 — MP BOM 最终发布（human/approval）**

```yaml
- id: mp_bom_release
  executor: human
  command: approval
  depends_on: [mp_golden_sample]
  input:
    title: "MP BOM 最终发布确认"
    description: |
      PVT 锁定的 BOM 成本: ${{steps.pvt_bom_submit.output.form.total_bom_cost}}
      请确认 BOM 可以最终发布。
    approvers:
      - open_id: "{{inputs.approver}}"
```

**Step 4.7 — MP Gate Review（human/approval）**

```yaml
- id: mp_gate
  executor: human
  command: approval
  depends_on: [mp_bom_release, mp_cert_confirm, mp_launch_prep, mp_aftersales_prep]
  input:
    title: "MP 阶段门审 — 最终放行"
    description: |
      ## 产品: {{steps.input_form.output.form.product_name}}

      | 指标 | 结果 | 标准 |
      |------|------|------|
      | 日产能 | {{steps.mp_ramp.output.form.actual_daily_capacity}} | 达标 |
      | 连续良率 | {{steps.mp_ramp.output.form.consecutive_yield_5day}}% | ≥95% |
      | Golden Sample | {{steps.mp_golden_sample.output.result}} | approved |
      | 认证证书 | {{steps.mp_cert_confirm.output.result}} | approved |
      | BOM 发布 | {{steps.mp_bom_release.output.result}} | approved |
      | 宣传物料 | {{steps.mp_launch_prep.output.form.marketing_materials}} | yes |
      | 使用手册 | {{steps.mp_launch_prep.output.form.user_manual}} | yes |
      | 维修手册 | {{steps.mp_aftersales_prep.output.form.repair_manual}} | yes |
      | RMA 流程 | {{steps.mp_aftersales_prep.output.form.rma_process}} | yes |

      通过 = 正式量产出货 | 驳回 = 流程暂停
    approvers:
      - open_id: "{{inputs.approver}}"
```

---

### Phase 5: 项目收尾（6 步）

```yaml
# 推进项目到 completed
- id: advance_to_completed
  executor: plm
  command: advance_project
  depends_on: [mp_gate]
  input:
    project_id: "{{steps.create_project.output.project_id}}"
    target_phase: completed
    comment: "MP Gate 通过，项目完成"

# 关闭主任务
- id: close_master_task
  executor: plm
  command: transition_task
  depends_on: [mp_gate]
  on_failure: skip
  input:
    task_id: "{{steps.create_master_task.output.task_id}}"
    status: completed
    comment: "EVT/DVT/PVT/MP 全部门审通过"

# 发布产品
- id: release_product
  executor: plm
  command: release_product
  depends_on: [mp_gate]
  on_failure: skip
  input:
    product_id: "{{steps.create_product.output.product_id}}"
    version: "1.0"
    comment: "量产发布"

# 完成 SRM 项目
- id: complete_srm
  executor: srm
  command: complete_srm_project
  depends_on: [mp_gate]
  on_failure: skip
  input:
    project_id: "{{steps.create_srm_project.output.project_id}}"
    comment: "产品量产达成"

# Agent 生成项目总结报告
- id: project_summary
  executor: llm
  command: generate
  depends_on: [mp_gate]
  input:
    model: claude-sonnet-4-6
    system: 你是项目管理专家。生成产品开发项目完结总结报告。
    prompt: |
      产品: {{steps.input_form.output.form.product_name}}
      目标 BOM: ${{steps.input_form.output.form.target_price}} → 最终: ${{steps.pvt_bom_submit.output.form.total_bom_cost}}
      目标重量: {{steps.input_form.output.form.target_weight}}g → 最终: {{steps.evt_me_submit.output.form.total_weight_g}}g
      试产良率: {{steps.pvt_trial_production.output.form.yield_rate}}%
      量产良率: {{steps.mp_ramp.output.form.consecutive_yield_5day}}%

      请生成项目完结报告，包含：目标达成总结、关键技术决策回顾、经验教训、改善建议。

# 完成通知
- id: done_notify
  executor: human
  command: notification
  depends_on: [project_summary, close_master_task, release_product, complete_srm, advance_to_completed]
  input:
    assignee: "{{inputs.approver}}"
    title: "产品开发完成: {{steps.input_form.output.form.product_name}}"
    prompt: |
      所有阶段 (EVT → DVT → PVT → MP) 门审通过。
      最终 BOM: ${{steps.pvt_bom_submit.output.form.total_bom_cost}}
      量产良率: {{steps.mp_ramp.output.form.consecutive_yield_5day}}%

      ## 项目总结
      {{steps.project_summary.output.result}}
```

---

## 五、各阶段门审标准

### EVT Gate

| 指标 | 标准 |
|------|------|
| 核心功能 | 所有 P0 功能可演示 |
| BOM 成本 | 目标 ±20% |
| 重量 | 目标 ±15% |
| 测试通过率 | >80% |
| 关键物料 | 长交期物料已锁定供应商 |
| 设计评审 | 联合评审通过 |

### DVT Gate

| 指标 | 标准 |
|------|------|
| 设计冻结 | ID/EE/ME 已冻结 |
| 可靠性 | 跌落/温湿度/铰链寿命 pass |
| EMC | FCC/CE 预测试 pass |
| BOM 成本 | 目标 ±10% |
| 模具 | T1 关键尺寸 CPK ≥1.0 |
| 固件 | P0 功能 100% 完成 |
| 供应商 | 关键供应商已定点 |

### PVT Gate

| 指标 | 标准 |
|------|------|
| 试产良率 | ≥95% |
| 法规认证 | 已送审或已通过 |
| 包装 | 跌落测试 pass |
| BOM 成本 | 目标 ±5% |
| OTA | 全链路测试 pass |
| PFMEA | 控制计划已建立 |
| 产线 CPK | ≥1.33 |

### MP Gate

| 指标 | 标准 |
|------|------|
| 日产能 | 达到目标 |
| 连续良率 | 连续 5 天 ≥95% |
| 认证证书 | 全部已获取 |
| Golden Sample | 已签样 |
| 发布物料 | 手册/标签/宣传完成 |
| 售后 | 维修手册/备件/RMA 就绪 |

---

## 六、错误处理

| 场景 | 处理方式 |
|------|----------|
| 立项审批被拒 | 流程终止 |
| 设计评审被拒 | 流程暂停（abort），修改后重新启动 |
| BOM Review 被拒 | 不进入 Gate，修改 BOM 后重新提交 |
| Gate 被拒 | 流程暂停（abort），解决问题后重新启动 |
| SRM 操作失败 | skip，不阻断主流程 |
| PLM 操作失败 | skip，不阻断主流程 |

---

## 七、Credentials

| Slug | 用途 |
|------|------|
| `github-token` | 固件/软件仓库操作（如有需要）|

---

## 八、系统集成

| 系统 | 角色 | 涉及步骤 |
|------|------|----------|
| **PLM** | 产品/项目/任务/文档管理、阶段推进、产品发布 | create_product, create_project, create_task, create_document, advance_project, transition_task, release_product |
| **SRM** | 供应链管理、采购 | create_srm_project, create_pr, complete |
| **Knowledge** | PRD 文档读取 | fetch |
| **LLM** | PRD 分析、概念生成、BOM 分析、项目总结 | generate |
| **Human** | 设计提交、测试结果、评审、Gate 审批 | form, approval, notification |

---

## 九、流程全景图

```
Phase 0 (立项):
  input_form → create_product → create_project → create_master_task
                              → read_prd → create_srm_project → create_prd_doc
  → project_approval

Phase 1 (EVT):
  advance_to_evt → evt_prd_analysis → evt_id_concept_gen
  → [并行] evt_id_submit / evt_ee_submit / evt_me_submit / evt_optics_submit / evt_audio_cmf_submit / evt_firmware_submit
  → evt_design_review → evt_procurement → evt_test_submit → evt_bom_submit
  → evt_bom_analysis → evt_bom_review → evt_gate

Phase 2 (DVT):
  advance_to_dvt → dvt_design_freeze
  → [并行] dvt_cmf / dvt_ee / dvt_mold / dvt_optics / dvt_firmware
  → [并行] dvt_reliability_test / dvt_emc_test / dvt_regulatory_pretest
  → dvt_vendor_selection → dvt_bom_submit → dvt_bom_review → dvt_gate

Phase 3 (PVT):
  advance_to_pvt →
  [并行] pvt_mold_final / pvt_test_station / pvt_pfmea / pvt_ota_test
  → pvt_trial_production
  → [并行] pvt_regulatory_submit / pvt_packaging / pvt_mass_procurement / pvt_bom_submit
  → pvt_bom_review → pvt_gate

Phase 4 (MP):
  advance_to_mp → mp_ramp → mp_golden_sample
  mp_cert_confirm (并行)
  → [并行] mp_launch_prep / mp_aftersales_prep
  → mp_bom_release → mp_gate

Phase 5 (收尾):
  advance_to_completed + close_master_task + release_product + complete_srm + project_summary → done_notify
```

总步骤数：**约 63 步**（Phase 0: 7, EVT: 16, DVT: 15, PVT: 13, MP: 9, Closure: 6）
