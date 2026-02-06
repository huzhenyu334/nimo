-- =====================================================
-- PLM v2.1 - 项目模板与任务自动化
-- =====================================================

-- 1. 项目模板表
CREATE TABLE IF NOT EXISTS project_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    template_type VARCHAR(20) NOT NULL DEFAULT 'CUSTOM' CHECK (template_type IN ('SYSTEM', 'CUSTOM')),
    product_type VARCHAR(50),  -- GLASSES/ACCESSORY/PLATFORM
    phases JSONB NOT NULL DEFAULT '["CONCEPT","EVT","DVT","PVT","MP"]',
    estimated_days INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT true,
    parent_template_id UUID REFERENCES project_templates(id),
    version INTEGER NOT NULL DEFAULT 1,
    created_by VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_project_templates_type ON project_templates(template_type);
CREATE INDEX idx_project_templates_active ON project_templates(is_active);

-- 2. 模板任务表
CREATE TABLE IF NOT EXISTS template_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES project_templates(id) ON DELETE CASCADE,
    task_code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    phase VARCHAR(20) NOT NULL,
    parent_task_code VARCHAR(50),
    task_type VARCHAR(20) NOT NULL DEFAULT 'TASK' CHECK (task_type IN ('MILESTONE', 'TASK', 'SUBTASK')),
    default_assignee_role VARCHAR(50),
    estimated_days INTEGER NOT NULL DEFAULT 1,
    is_critical BOOLEAN NOT NULL DEFAULT false,
    deliverables JSONB,  -- 交付物定义
    checklist JSONB,     -- 检查清单
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    approval_type VARCHAR(50),  -- REVIEW_MEETING / FEISHU_APPROVAL
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(template_id, task_code)
);

CREATE INDEX idx_template_tasks_template ON template_tasks(template_id);
CREATE INDEX idx_template_tasks_phase ON template_tasks(phase);
CREATE INDEX idx_template_tasks_parent ON template_tasks(template_id, parent_task_code);

-- 3. 模板任务依赖表
CREATE TABLE IF NOT EXISTS template_task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES project_templates(id) ON DELETE CASCADE,
    task_code VARCHAR(50) NOT NULL,
    depends_on_task_code VARCHAR(50) NOT NULL,
    dependency_type VARCHAR(10) NOT NULL DEFAULT 'FS' CHECK (dependency_type IN ('FS', 'SS', 'FF', 'SF')),
    lag_days INTEGER DEFAULT 0,
    UNIQUE(template_id, task_code, depends_on_task_code)
);

CREATE INDEX idx_template_deps_template ON template_task_dependencies(template_id);

-- 4. 扩展项目表 - 添加模板关联
ALTER TABLE projects ADD COLUMN IF NOT EXISTS template_id UUID REFERENCES project_templates(id);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS auto_start_tasks BOOLEAN DEFAULT true;

-- 5. 扩展任务表 - 添加自动化相关字段
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS task_code VARCHAR(50);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS task_type VARCHAR(20) DEFAULT 'TASK';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS auto_start BOOLEAN DEFAULT false;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS deliverables JSONB;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS checklist JSONB;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN DEFAULT false;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS approval_type VARCHAR(50);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS approval_status VARCHAR(20);  -- PENDING/APPROVED/REJECTED

-- 6. 飞书任务同步表
CREATE TABLE IF NOT EXISTS feishu_task_sync (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    feishu_task_id VARCHAR(100) NOT NULL,
    feishu_task_guid VARCHAR(100),
    sync_status VARCHAR(20) NOT NULL DEFAULT 'SYNCED',
    last_sync_at TIMESTAMP,
    sync_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id),
    UNIQUE(feishu_task_id)
);

CREATE INDEX idx_feishu_task_sync_task ON feishu_task_sync(task_id);
CREATE INDEX idx_feishu_task_sync_feishu ON feishu_task_sync(feishu_task_id);

-- 7. 审批定义表
CREATE TABLE IF NOT EXISTS approval_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    feishu_approval_code VARCHAR(100),
    approval_type VARCHAR(20) NOT NULL CHECK (approval_type IN ('BOM', 'ECN', 'TASK', 'PHASE', 'DOCUMENT')),
    form_definition JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 8. 审批实例表
CREATE TABLE IF NOT EXISTS approval_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_def_id UUID REFERENCES approval_definitions(id),
    feishu_instance_code VARCHAR(100),
    business_type VARCHAR(20) NOT NULL,
    business_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED')),
    applicant_id VARCHAR(64) NOT NULL,
    form_data JSONB,
    approvers JSONB,  -- 审批人列表
    current_approver_id VARCHAR(64),
    comments TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    UNIQUE(feishu_instance_code)
);

CREATE INDEX idx_approval_instances_business ON approval_instances(business_type, business_id);
CREATE INDEX idx_approval_instances_status ON approval_instances(status);
CREATE INDEX idx_approval_instances_applicant ON approval_instances(applicant_id);

-- 9. 评审会议表
CREATE TABLE IF NOT EXISTS review_meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    meeting_type VARCHAR(20) NOT NULL CHECK (meeting_type IN ('DESIGN', 'PHASE', 'BOM', 'ECN')),
    project_id UUID REFERENCES projects(id),
    task_id UUID REFERENCES tasks(id),
    feishu_calendar_event_id VARCHAR(100),
    feishu_meeting_id VARCHAR(100),
    scheduled_at TIMESTAMP NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 60,
    location VARCHAR(200),
    organizer_id VARCHAR(64) NOT NULL,
    attendees JSONB NOT NULL DEFAULT '[]',
    agenda TEXT,
    documents JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED' CHECK (status IN ('SCHEDULED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    conclusion TEXT,
    action_items JSONB,
    minutes_doc_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_review_meetings_project ON review_meetings(project_id);
CREATE INDEX idx_review_meetings_task ON review_meetings(task_id);
CREATE INDEX idx_review_meetings_status ON review_meetings(status);
CREATE INDEX idx_review_meetings_scheduled ON review_meetings(scheduled_at);

-- 10. 自动化规则表
CREATE TABLE IF NOT EXISTS automation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_type VARCHAR(30) NOT NULL CHECK (rule_type IN ('TASK_START', 'TASK_COMPLETE', 'OVERDUE', 'PHASE_COMPLETE', 'APPROVAL_COMPLETE')),
    trigger_condition JSONB NOT NULL,
    action_type VARCHAR(30) NOT NULL CHECK (action_type IN ('UPDATE_STATUS', 'SEND_NOTIFICATION', 'CREATE_TASK', 'CREATE_APPROVAL', 'CREATE_MEETING')),
    action_config JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    project_id UUID REFERENCES projects(id),
    template_id UUID REFERENCES project_templates(id),
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_automation_rules_type ON automation_rules(rule_type);
CREATE INDEX idx_automation_rules_project ON automation_rules(project_id);
CREATE INDEX idx_automation_rules_template ON automation_rules(template_id);
CREATE INDEX idx_automation_rules_active ON automation_rules(is_active);

-- 11. 自动化执行日志表
CREATE TABLE IF NOT EXISTS automation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID REFERENCES automation_rules(id),
    trigger_event JSONB NOT NULL,
    action_result JSONB,
    status VARCHAR(20) NOT NULL CHECK (status IN ('SUCCESS', 'FAILED', 'SKIPPED')),
    error_message TEXT,
    executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_automation_logs_rule ON automation_logs(rule_id);
CREATE INDEX idx_automation_logs_status ON automation_logs(status);
CREATE INDEX idx_automation_logs_executed ON automation_logs(executed_at);

-- 12. 插入默认模板
INSERT INTO project_templates (id, code, name, description, template_type, product_type, phases, estimated_days, is_active, created_by) VALUES
('11111111-1111-1111-1111-111111111111', 'TPL-GLASSES-STD', '智能眼镜标准研发流程', '适用于智能眼镜产品的标准四阶段研发流程模板', 'SYSTEM', 'GLASSES', '["CONCEPT","EVT","DVT","PVT","MP"]', 180, true, 'system'),
('22222222-2222-2222-2222-222222222222', 'TPL-ACCESSORY', '配件快速开发流程', '适用于眼镜配件的简化开发流程', 'SYSTEM', 'ACCESSORY', '["CONCEPT","DVT","PVT","MP"]', 60, true, 'system')
ON CONFLICT (code) DO NOTHING;

-- 13. 插入默认模板任务（智能眼镜标准流程）
INSERT INTO template_tasks (template_id, task_code, name, phase, task_type, default_assignee_role, estimated_days, is_critical, requires_approval, sort_order) VALUES
-- CONCEPT 阶段
('11111111-1111-1111-1111-111111111111', 'CON-001', '产品立项', 'CONCEPT', 'MILESTONE', 'PM', 1, true, true, 1),
('11111111-1111-1111-1111-111111111111', 'CON-002', '市场调研', 'CONCEPT', 'TASK', 'PM', 5, false, false, 2),
('11111111-1111-1111-1111-111111111111', 'CON-003', '竞品分析', 'CONCEPT', 'TASK', 'PM', 3, false, false, 3),
('11111111-1111-1111-1111-111111111111', 'CON-004', '产品定义', 'CONCEPT', 'TASK', 'PM', 5, true, false, 4),
('11111111-1111-1111-1111-111111111111', 'CON-005', '立项评审', 'CONCEPT', 'MILESTONE', 'PM', 1, true, true, 5),

-- EVT 阶段
('11111111-1111-1111-1111-111111111111', 'EVT-001', '硬件设计', 'EVT', 'TASK', 'HW_ENG', 14, true, false, 10),
('11111111-1111-1111-1111-111111111111', 'EVT-001-01', '电路原理图设计', 'EVT', 'SUBTASK', 'HW_ENG', 7, true, false, 11),
('11111111-1111-1111-1111-111111111111', 'EVT-001-02', 'PCB布局设计', 'EVT', 'SUBTASK', 'HW_ENG', 5, true, false, 12),
('11111111-1111-1111-1111-111111111111', 'EVT-001-03', '硬件设计评审', 'EVT', 'SUBTASK', 'HW_LEAD', 2, true, true, 13),
('11111111-1111-1111-1111-111111111111', 'EVT-002', '结构设计', 'EVT', 'TASK', 'ME_ENG', 10, true, false, 20),
('11111111-1111-1111-1111-111111111111', 'EVT-002-01', '3D建模', 'EVT', 'SUBTASK', 'ME_ENG', 7, true, false, 21),
('11111111-1111-1111-1111-111111111111', 'EVT-002-02', '结构仿真分析', 'EVT', 'SUBTASK', 'ME_ENG', 3, false, false, 22),
('11111111-1111-1111-1111-111111111111', 'EVT-003', '光学设计', 'EVT', 'TASK', 'OPT_ENG', 8, true, false, 30),
('11111111-1111-1111-1111-111111111111', 'EVT-004', 'EVT样机制作', 'EVT', 'TASK', 'HW_ENG', 10, true, false, 40),
('11111111-1111-1111-1111-111111111111', 'EVT-004-01', 'PCB打样', 'EVT', 'SUBTASK', 'HW_ENG', 5, true, false, 41),
('11111111-1111-1111-1111-111111111111', 'EVT-004-02', '结构件加工', 'EVT', 'SUBTASK', 'ME_ENG', 7, true, false, 42),
('11111111-1111-1111-1111-111111111111', 'EVT-004-03', '样机组装', 'EVT', 'SUBTASK', 'HW_ENG', 3, true, false, 43),
('11111111-1111-1111-1111-111111111111', 'EVT-005', 'EVT测试验证', 'EVT', 'TASK', 'QA_ENG', 7, true, false, 50),
('11111111-1111-1111-1111-111111111111', 'EVT-005-01', '功能测试', 'EVT', 'SUBTASK', 'QA_ENG', 3, true, false, 51),
('11111111-1111-1111-1111-111111111111', 'EVT-005-02', '电气测试', 'EVT', 'SUBTASK', 'QA_ENG', 2, true, false, 52),
('11111111-1111-1111-1111-111111111111', 'EVT-005-03', '光学测试', 'EVT', 'SUBTASK', 'QA_ENG', 2, false, false, 53),
('11111111-1111-1111-1111-111111111111', 'EVT-006', 'EVT评审', 'EVT', 'MILESTONE', 'PM', 2, true, true, 60),

-- DVT 阶段
('11111111-1111-1111-1111-111111111111', 'DVT-001', '设计优化', 'DVT', 'TASK', 'HW_ENG', 14, true, false, 100),
('11111111-1111-1111-1111-111111111111', 'DVT-001-01', '硬件问题修复', 'DVT', 'SUBTASK', 'HW_ENG', 7, true, false, 101),
('11111111-1111-1111-1111-111111111111', 'DVT-001-02', '结构优化', 'DVT', 'SUBTASK', 'ME_ENG', 7, true, false, 102),
('11111111-1111-1111-1111-111111111111', 'DVT-002', 'DVT样机制作', 'DVT', 'TASK', 'HW_ENG', 14, true, false, 110),
('11111111-1111-1111-1111-111111111111', 'DVT-003', '全面测试验证', 'DVT', 'TASK', 'QA_ENG', 21, true, false, 120),
('11111111-1111-1111-1111-111111111111', 'DVT-003-01', '环境测试', 'DVT', 'SUBTASK', 'QA_ENG', 7, true, false, 121),
('11111111-1111-1111-1111-111111111111', 'DVT-003-02', '可靠性测试', 'DVT', 'SUBTASK', 'QA_ENG', 14, true, false, 122),
('11111111-1111-1111-1111-111111111111', 'DVT-003-03', '用户体验测试', 'DVT', 'SUBTASK', 'QA_ENG', 7, false, false, 123),
('11111111-1111-1111-1111-111111111111', 'DVT-004', '认证准备', 'DVT', 'TASK', 'QA_ENG', 14, false, false, 130),
('11111111-1111-1111-1111-111111111111', 'DVT-005', 'BOM定稿', 'DVT', 'TASK', 'HW_ENG', 5, true, true, 140),
('11111111-1111-1111-1111-111111111111', 'DVT-006', 'DVT评审', 'DVT', 'MILESTONE', 'PM', 2, true, true, 150),

-- PVT 阶段
('11111111-1111-1111-1111-111111111111', 'PVT-001', '生产准备', 'PVT', 'TASK', 'PM', 14, true, false, 200),
('11111111-1111-1111-1111-111111111111', 'PVT-001-01', '供应商确认', 'PVT', 'SUBTASK', 'PM', 7, true, false, 201),
('11111111-1111-1111-1111-111111111111', 'PVT-001-02', '生产线准备', 'PVT', 'SUBTASK', 'PM', 7, true, false, 202),
('11111111-1111-1111-1111-111111111111', 'PVT-002', 'PVT试产', 'PVT', 'TASK', 'PM', 14, true, false, 210),
('11111111-1111-1111-1111-111111111111', 'PVT-003', '量产验证', 'PVT', 'TASK', 'QA_ENG', 7, true, false, 220),
('11111111-1111-1111-1111-111111111111', 'PVT-004', '认证完成', 'PVT', 'TASK', 'QA_ENG', 14, false, false, 230),
('11111111-1111-1111-1111-111111111111', 'PVT-005', 'PVT评审', 'PVT', 'MILESTONE', 'PM', 2, true, true, 240),

-- MP 阶段
('11111111-1111-1111-1111-111111111111', 'MP-001', '量产启动', 'MP', 'MILESTONE', 'PM', 1, true, true, 300),
('11111111-1111-1111-1111-111111111111', 'MP-002', '首批生产', 'MP', 'TASK', 'PM', 7, true, false, 310),
('11111111-1111-1111-1111-111111111111', 'MP-003', '质量监控', 'MP', 'TASK', 'QA_ENG', 30, false, false, 320)
ON CONFLICT DO NOTHING;

-- 14. 插入模板任务依赖
INSERT INTO template_task_dependencies (template_id, task_code, depends_on_task_code, dependency_type, lag_days) VALUES
-- CONCEPT 依赖
('11111111-1111-1111-1111-111111111111', 'CON-004', 'CON-002', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'CON-004', 'CON-003', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'CON-005', 'CON-004', 'FS', 0),

-- EVT 依赖
('11111111-1111-1111-1111-111111111111', 'EVT-001', 'CON-005', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-001-02', 'EVT-001-01', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-001-03', 'EVT-001-02', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-002', 'CON-005', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-002-02', 'EVT-002-01', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-003', 'CON-005', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-004', 'EVT-001', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-004', 'EVT-002', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-004', 'EVT-003', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-004-02', 'EVT-002-01', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-004-03', 'EVT-004-01', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-004-03', 'EVT-004-02', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-005', 'EVT-004', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'EVT-006', 'EVT-005', 'FS', 0),

-- DVT 依赖
('11111111-1111-1111-1111-111111111111', 'DVT-001', 'EVT-006', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'DVT-002', 'DVT-001', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'DVT-003', 'DVT-002', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'DVT-004', 'DVT-002', 'SS', 0),
('11111111-1111-1111-1111-111111111111', 'DVT-005', 'DVT-003', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'DVT-006', 'DVT-003', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'DVT-006', 'DVT-005', 'FS', 0),

-- PVT 依赖
('11111111-1111-1111-1111-111111111111', 'PVT-001', 'DVT-006', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'PVT-002', 'PVT-001', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'PVT-003', 'PVT-002', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'PVT-004', 'DVT-004', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'PVT-005', 'PVT-003', 'FS', 0),

-- MP 依赖
('11111111-1111-1111-1111-111111111111', 'MP-001', 'PVT-005', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'MP-002', 'MP-001', 'FS', 0),
('11111111-1111-1111-1111-111111111111', 'MP-003', 'MP-002', 'FS', 0)
ON CONFLICT DO NOTHING;

-- 15. 插入默认审批定义
INSERT INTO approval_definitions (id, code, name, approval_type, is_active) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'APR-BOM', 'BOM发布审批', 'BOM', true),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'APR-ECN', 'ECN变更审批', 'ECN', true),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'APR-TASK', '任务交付物审批', 'TASK', true),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'APR-PHASE', '阶段评审审批', 'PHASE', true)
ON CONFLICT (code) DO NOTHING;

-- 完成
SELECT 'Migration 003 completed successfully' as status;
