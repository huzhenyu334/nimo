-- PLM V2.0 重构迁移
-- 核心改动：BOM从产品关联改为项目关联，增加代号、交付物、阶段门控

BEGIN;

-- ============================================================
-- 1. 项目表增加代号字段
-- ============================================================
ALTER TABLE projects ADD COLUMN IF NOT EXISTS codename VARCHAR(32);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS project_type VARCHAR(16) DEFAULT 'platform';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS platform_id VARCHAR(32);
ALTER TABLE projects ADD COLUMN IF NOT EXISTS manager_id VARCHAR(32);
-- rename current_phase if it has old name
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='projects' AND column_name='current_phase') THEN
    NULL; -- already exists
  END IF;
END $$;

-- ============================================================
-- 2. 项目阶段表增加门控字段
-- ============================================================
ALTER TABLE project_phases ADD COLUMN IF NOT EXISTS gate_review_status VARCHAR(16) DEFAULT 'not_started';
ALTER TABLE project_phases ADD COLUMN IF NOT EXISTS gate_review_date TIMESTAMPTZ;
ALTER TABLE project_phases ADD COLUMN IF NOT EXISTS gate_review_notes TEXT;
ALTER TABLE project_phases ADD COLUMN IF NOT EXISTS gate_reviewer_id VARCHAR(32);

-- ============================================================
-- 3. 重构BOM表 — 从product关联改为project+phase关联
-- ============================================================

-- 3a. 新建project_boms表（保留旧bom_headers不动，新流程用新表）
CREATE TABLE IF NOT EXISTS project_boms (
    id VARCHAR(32) PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    project_id VARCHAR(32) NOT NULL REFERENCES projects(id),
    phase_id VARCHAR(32) REFERENCES project_phases(id),
    bom_type VARCHAR(16) NOT NULL DEFAULT 'EBOM',  -- EBOM/SBOM/OBOM/FWBOM
    version VARCHAR(16) NOT NULL DEFAULT 'v1.0',
    name VARCHAR(128) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',  -- draft/pending_review/published/frozen
    description TEXT,
    submitted_by VARCHAR(32) REFERENCES users(id),
    submitted_at TIMESTAMPTZ,
    reviewed_by VARCHAR(32) REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    review_comment TEXT,
    approved_by VARCHAR(32) REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    frozen_at TIMESTAMPTZ,
    frozen_by VARCHAR(32) REFERENCES users(id),
    parent_bom_id VARCHAR(32) REFERENCES project_boms(id),  -- 阶段演进继承
    total_items INT DEFAULT 0,
    estimated_cost NUMERIC(15,4),
    created_by VARCHAR(32) NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_bom_type CHECK (bom_type IN ('EBOM','SBOM','OBOM','FWBOM')),
    CONSTRAINT ck_bom_status CHECK (status IN ('draft','pending_review','published','frozen','rejected'))
);

CREATE INDEX IF NOT EXISTS idx_project_boms_project ON project_boms(project_id);
CREATE INDEX IF NOT EXISTS idx_project_boms_phase ON project_boms(phase_id);
CREATE INDEX IF NOT EXISTS idx_project_boms_status ON project_boms(status);

-- 3b. 新建project_bom_items表
CREATE TABLE IF NOT EXISTS project_bom_items (
    id VARCHAR(32) PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    bom_id VARCHAR(32) NOT NULL REFERENCES project_boms(id) ON DELETE CASCADE,
    item_number INT NOT NULL DEFAULT 0,
    material_id VARCHAR(32) REFERENCES materials(id),  -- 可选，新物料先不入库
    category VARCHAR(32),  -- IC/passive/connector/structural/optical/firmware
    name VARCHAR(128) NOT NULL,
    specification TEXT,
    quantity NUMERIC(15,4) NOT NULL DEFAULT 1,
    unit VARCHAR(16) NOT NULL DEFAULT 'pcs',
    reference VARCHAR(256),  -- 位号 R1,R2,C1 (电子件)
    manufacturer VARCHAR(128),
    manufacturer_pn VARCHAR(64),
    supplier VARCHAR(128),
    unit_price NUMERIC(15,4),
    lead_time_days INT,
    is_critical BOOLEAN DEFAULT FALSE,
    is_alternative BOOLEAN DEFAULT FALSE,
    alternative_for VARCHAR(32) REFERENCES project_bom_items(id),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pbom_items_bom ON project_bom_items(bom_id);
CREATE INDEX IF NOT EXISTS idx_pbom_items_material ON project_bom_items(material_id);

-- ============================================================
-- 4. 改造文档表 — 增加项目关联和分类
-- ============================================================
ALTER TABLE documents ADD COLUMN IF NOT EXISTS project_id VARCHAR(32) REFERENCES projects(id);
ALTER TABLE documents ADD COLUMN IF NOT EXISTS phase_id VARCHAR(32) REFERENCES project_phases(id);
ALTER TABLE documents ADD COLUMN IF NOT EXISTS doc_type VARCHAR(32) DEFAULT 'project';
  -- electrical/mechanical/optical/id_design/firmware/test/project
ALTER TABLE documents ADD COLUMN IF NOT EXISTS current_version VARCHAR(16);

-- 文档版本表增加ECN关联
ALTER TABLE document_versions ADD COLUMN IF NOT EXISTS ecn_id VARCHAR(32);
ALTER TABLE document_versions ADD COLUMN IF NOT EXISTS change_description TEXT;

-- ============================================================
-- 5. 新建阶段交付物表
-- ============================================================
CREATE TABLE IF NOT EXISTS phase_deliverables (
    id VARCHAR(32) PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    phase_id VARCHAR(32) NOT NULL REFERENCES project_phases(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    deliverable_type VARCHAR(16) NOT NULL DEFAULT 'document',  -- document/bom/review
    responsible_role VARCHAR(32),
    is_required BOOLEAN DEFAULT TRUE,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending/submitted/approved
    document_id VARCHAR(32) REFERENCES documents(id),
    bom_id VARCHAR(32) REFERENCES project_boms(id),
    submitted_at TIMESTAMPTZ,
    submitted_by VARCHAR(32) REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    approved_by VARCHAR(32) REFERENCES users(id),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deliverables_phase ON phase_deliverables(phase_id);

-- ============================================================
-- 6. 新建项目代号库表
-- ============================================================
CREATE TABLE IF NOT EXISTS project_codenames (
    id VARCHAR(32) PRIMARY KEY DEFAULT replace(uuid_generate_v4()::text, '-', ''),
    codename VARCHAR(32) NOT NULL,
    codename_type VARCHAR(16) NOT NULL,  -- platform/product
    generation INT,  -- 代次(平台专用)
    theme VARCHAR(64),
    description VARCHAR(256),
    is_used BOOLEAN DEFAULT FALSE,
    used_by_project_id VARCHAR(32) REFERENCES projects(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_codenames_unique ON project_codenames(codename, codename_type);

-- ============================================================
-- 7. ECN表增加BOM关联
-- ============================================================
ALTER TABLE ecns ADD COLUMN IF NOT EXISTS project_id VARCHAR(32) REFERENCES projects(id);
ALTER TABLE ecns ADD COLUMN IF NOT EXISTS phase_id VARCHAR(32) REFERENCES project_phases(id);
ALTER TABLE ecns ADD COLUMN IF NOT EXISTS bom_id VARCHAR(32) REFERENCES project_boms(id);

-- ============================================================
-- 8. 插入平台代号预置数据
-- ============================================================
INSERT INTO project_codenames (id, codename, codename_type, generation, theme, description) VALUES
(replace(uuid_generate_v4()::text,'-',''), '微光', 'platform', 1, '光之起源', '微弱之光，万物初始'),
(replace(uuid_generate_v4()::text,'-',''), '晨曦', 'platform', 2, '天色将明', '破晓微光'),
(replace(uuid_generate_v4()::text,'-',''), '朝霞', 'platform', 3, '日出之前', '霞光万道'),
(replace(uuid_generate_v4()::text,'-',''), '旭日', 'platform', 4, '冉冉升起', '旭日东升'),
(replace(uuid_generate_v4()::text,'-',''), '明辉', 'platform', 5, '光芒绽放', '光明辉映'),
(replace(uuid_generate_v4()::text,'-',''), '皓月', 'platform', 6, '月之清辉', '皓月当空'),
(replace(uuid_generate_v4()::text,'-',''), '星河', 'platform', 7, '仰望星空', '银河璀璨'),
(replace(uuid_generate_v4()::text,'-',''), '天枢', 'platform', 8, '星辰导航', '北斗第一星'),
(replace(uuid_generate_v4()::text,'-',''), '瑶光', 'platform', 9, '玉光闪耀', '北斗第七星'),
(replace(uuid_generate_v4()::text,'-',''), '紫微', 'platform', 10, '紫微垣', '帝星，众星之主'),
(replace(uuid_generate_v4()::text,'-',''), '青龙', 'platform', 11, '四象之一', '东方之神'),
(replace(uuid_generate_v4()::text,'-',''), '朱雀', 'platform', 12, '四象之二', '南方之神'),
(replace(uuid_generate_v4()::text,'-',''), '玄武', 'platform', 13, '四象之三', '北方之神'),
(replace(uuid_generate_v4()::text,'-',''), '白虎', 'platform', 14, '四象之四', '西方之神'),
(replace(uuid_generate_v4()::text,'-',''), '麒麟', 'platform', 15, '太平盛世', '瑞兽之首'),
(replace(uuid_generate_v4()::text,'-',''), '凤凰', 'platform', 16, '涅槃重生', '百鸟之王'),
(replace(uuid_generate_v4()::text,'-',''), '鲲鹏', 'platform', 17, '逍遥游', '展翅九万里'),
(replace(uuid_generate_v4()::text,'-',''), '九天', 'platform', 18, '天界之巅', '九重天阙'),
(replace(uuid_generate_v4()::text,'-',''), '太极', 'platform', 19, '万物之本', '阴阳之源'),
(replace(uuid_generate_v4()::text,'-',''), '鸿蒙', 'platform', 20, '宇宙本源', '天地未开之始')
ON CONFLICT DO NOTHING;

-- 插入镜框代号预置数据
INSERT INTO project_codenames (id, codename, codename_type, theme, description) VALUES
(replace(uuid_generate_v4()::text,'-',''), 'Nova', 'product', '新星', '新星爆发'),
(replace(uuid_generate_v4()::text,'-',''), 'BigBang', 'product', '宇宙大爆炸', '起源之力'),
(replace(uuid_generate_v4()::text,'-',''), 'BlackHole', 'product', '黑洞', '神秘深邃'),
(replace(uuid_generate_v4()::text,'-',''), 'Galaxy', 'product', '星系', '璀璨星系'),
(replace(uuid_generate_v4()::text,'-',''), 'Venus', 'product', '金星', '最亮行星'),
(replace(uuid_generate_v4()::text,'-',''), 'Mars', 'product', '火星', '红色星球'),
(replace(uuid_generate_v4()::text,'-',''), 'Orion', 'product', '猎户座', '星座之王'),
(replace(uuid_generate_v4()::text,'-',''), 'Pulsar', 'product', '脉冲星', '旋转发光'),
(replace(uuid_generate_v4()::text,'-',''), 'Quasar', 'product', '类星体', '宇宙灯塔'),
(replace(uuid_generate_v4()::text,'-',''), 'Nebula', 'product', '星云', '梦幻星云'),
(replace(uuid_generate_v4()::text,'-',''), 'Eclipse', 'product', '日食', '光影交错'),
(replace(uuid_generate_v4()::text,'-',''), 'Aurora', 'product', '极光', '北极光'),
(replace(uuid_generate_v4()::text,'-',''), 'Meteor', 'product', '流星', '划过天际'),
(replace(uuid_generate_v4()::text,'-',''), 'Comet', 'product', '彗星', '拖尾之星'),
(replace(uuid_generate_v4()::text,'-',''), 'Zenith', 'product', '天顶', '最高点'),
(replace(uuid_generate_v4()::text,'-',''), 'Horizon', 'product', '地平线', '无限延伸'),
(replace(uuid_generate_v4()::text,'-',''), 'Spectrum', 'product', '光谱', '七彩光谱'),
(replace(uuid_generate_v4()::text,'-',''), 'Stellar', 'product', '恒星', '永恒之光'),
(replace(uuid_generate_v4()::text,'-',''), 'Cosmos', 'product', '宇宙', '浩瀚宇宙'),
(replace(uuid_generate_v4()::text,'-',''), 'Vortex', 'product', '漩涡', '星际漩涡'),
(replace(uuid_generate_v4()::text,'-',''), 'Titan', 'product', '泰坦', '土卫六'),
(replace(uuid_generate_v4()::text,'-',''), 'Atlas', 'product', '阿特拉斯', '擎天之柱'),
(replace(uuid_generate_v4()::text,'-',''), 'Phoenix', 'product', '凤凰', '浴火重生'),
(replace(uuid_generate_v4()::text,'-',''), 'Sirius', 'product', '天狼星', '夜空最亮'),
(replace(uuid_generate_v4()::text,'-',''), 'Polaris', 'product', '北极星', '方向之星'),
(replace(uuid_generate_v4()::text,'-',''), 'Andromeda', 'product', '仙女座', '最近星系'),
(replace(uuid_generate_v4()::text,'-',''), 'Voyager', 'product', '旅行者', '星际探索'),
(replace(uuid_generate_v4()::text,'-',''), 'Cassini', 'product', '卡西尼', '土星探索'),
(replace(uuid_generate_v4()::text,'-',''), 'Lunar', 'product', '月球', '月之光'),
(replace(uuid_generate_v4()::text,'-',''), 'Solaris', 'product', '太阳', '万物之源')
ON CONFLICT DO NOTHING;

-- ============================================================
-- 9. 更新触发器
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'tr_project_boms_updated_at') THEN
        CREATE TRIGGER tr_project_boms_updated_at BEFORE UPDATE ON project_boms FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'tr_pbom_items_updated_at') THEN
        CREATE TRIGGER tr_pbom_items_updated_at BEFORE UPDATE ON project_bom_items FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'tr_deliverables_updated_at') THEN
        CREATE TRIGGER tr_deliverables_updated_at BEFORE UPDATE ON phase_deliverables FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
END $$;

COMMIT;
