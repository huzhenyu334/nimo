-- 物料分类体系升级：5个一级 → 8个一级 + 子类
-- 二级分类code格式: {大类code}-{子类code}，如 EL-CAP

BEGIN;

-- 0. 先去掉code的唯一约束，改为 (code) 不要求全局唯一（或改为 parent_id+code 联合唯一）
ALTER TABLE material_categories DROP CONSTRAINT IF EXISTS material_categories_code_key;
-- 暂不加新约束，后续可加 UNIQUE(parent_id, code)

-- 1. 更新现有一级分类的code为2位大写，parent_id 设为 NULL
UPDATE material_categories SET code = 'EL', sort_order = 1, parent_id = NULL WHERE id = 'mcat_electronic';
UPDATE material_categories SET code = 'ME', sort_order = 2, parent_id = NULL WHERE id = 'mcat_mechanical';
UPDATE material_categories SET code = 'OP', sort_order = 3, parent_id = NULL WHERE id = 'mcat_optical';
UPDATE material_categories SET code = 'PK', sort_order = 4, parent_id = NULL WHERE id = 'mcat_packaging';
UPDATE material_categories SET code = 'OT', name = '其他', sort_order = 99, parent_id = NULL WHERE id = 'mcat_other';

-- 2. 插入新增的一级分类
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_auxiliary', 'AX', '辅料/耗材',   NULL, '', 1, 5, NOW(), NOW()),
('mcat_software',  'SW', '软件/固件',   NULL, '', 1, 6, NOW(), NOW()),
('mcat_finished',  'FG', '成品',        NULL, '', 1, 7, NOW(), NOW()),
('mcat_subassy',   'SA', '半成品/组件', NULL, '', 1, 8, NOW(), NOW());

-- 3. 二级分类 - 电子元器件 (EL)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_el_res', 'EL-RES', '电阻',        'mcat_electronic', 'mcat_electronic', 2, 1, NOW(), NOW()),
('mcat_el_cap', 'EL-CAP', '电容',        'mcat_electronic', 'mcat_electronic', 2, 2, NOW(), NOW()),
('mcat_el_ind', 'EL-IND', '电感',        'mcat_electronic', 'mcat_electronic', 2, 3, NOW(), NOW()),
('mcat_el_ic',  'EL-IC',  '集成电路',     'mcat_electronic', 'mcat_electronic', 2, 4, NOW(), NOW()),
('mcat_el_con', 'EL-CON', '连接器',      'mcat_electronic', 'mcat_electronic', 2, 5, NOW(), NOW()),
('mcat_el_dio', 'EL-DIO', '二极管/ESD',  'mcat_electronic', 'mcat_electronic', 2, 6, NOW(), NOW()),
('mcat_el_trn', 'EL-TRN', '晶体管',     'mcat_electronic', 'mcat_electronic', 2, 7, NOW(), NOW()),
('mcat_el_osc', 'EL-OSC', '晶振/时钟',   'mcat_electronic', 'mcat_electronic', 2, 8, NOW(), NOW()),
('mcat_el_led', 'EL-LED', 'LED/背光',    'mcat_electronic', 'mcat_electronic', 2, 9, NOW(), NOW()),
('mcat_el_sen', 'EL-SEN', '传感器',      'mcat_electronic', 'mcat_electronic', 2, 10, NOW(), NOW()),
('mcat_el_ant', 'EL-ANT', '天线',        'mcat_electronic', 'mcat_electronic', 2, 11, NOW(), NOW()),
('mcat_el_mod', 'EL-MOD', '模组',        'mcat_electronic', 'mcat_electronic', 2, 12, NOW(), NOW()),
('mcat_el_bat', 'EL-BAT', '电池',        'mcat_electronic', 'mcat_electronic', 2, 13, NOW(), NOW()),
('mcat_el_pcb', 'EL-PCB', 'PCB',         'mcat_electronic', 'mcat_electronic', 2, 14, NOW(), NOW()),
('mcat_el_oth', 'EL-OTH', '其他电子',    'mcat_electronic', 'mcat_electronic', 2, 15, NOW(), NOW());

-- 4. 二级分类 - 结构件 (ME)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_me_hsg', 'ME-HSG', '外壳/壳体',  'mcat_mechanical', 'mcat_mechanical', 2, 1, NOW(), NOW()),
('mcat_me_lns', 'ME-LNS', '镜片',       'mcat_mechanical', 'mcat_mechanical', 2, 2, NOW(), NOW()),
('mcat_me_flx', 'ME-FLX', '柔性件',     'mcat_mechanical', 'mcat_mechanical', 2, 3, NOW(), NOW()),
('mcat_me_fst', 'ME-FST', '紧固件',     'mcat_mechanical', 'mcat_mechanical', 2, 4, NOW(), NOW()),
('mcat_me_gsk', 'ME-GSK', '密封件',     'mcat_mechanical', 'mcat_mechanical', 2, 5, NOW(), NOW()),
('mcat_me_spg', 'ME-SPG', '弹性件',     'mcat_mechanical', 'mcat_mechanical', 2, 6, NOW(), NOW()),
('mcat_me_thm', 'ME-THM', '散热件',     'mcat_mechanical', 'mcat_mechanical', 2, 7, NOW(), NOW()),
('mcat_me_dec', 'ME-DEC', '装饰件',     'mcat_mechanical', 'mcat_mechanical', 2, 8, NOW(), NOW()),
('mcat_me_oth', 'ME-OTH', '其他结构',   'mcat_mechanical', 'mcat_mechanical', 2, 9, NOW(), NOW());

-- 5. 二级分类 - 光学器件 (OP)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_op_dis', 'OP-DIS', '显示器件',   'mcat_optical', 'mcat_optical', 2, 1, NOW(), NOW()),
('mcat_op_prj', 'OP-PRJ', '投影器件',   'mcat_optical', 'mcat_optical', 2, 2, NOW(), NOW()),
('mcat_op_wgd', 'OP-WGD', '光波导',     'mcat_optical', 'mcat_optical', 2, 3, NOW(), NOW()),
('mcat_op_oth', 'OP-OTH', '其他光学',   'mcat_optical', 'mcat_optical', 2, 4, NOW(), NOW());

-- 6. 二级分类 - 包材 (PK)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_pk_box', 'PK-BOX', '包装盒',     'mcat_packaging', 'mcat_packaging', 2, 1, NOW(), NOW()),
('mcat_pk_bag', 'PK-BAG', '袋类',       'mcat_packaging', 'mcat_packaging', 2, 2, NOW(), NOW()),
('mcat_pk_ins', 'PK-INS', '插页/说明书', 'mcat_packaging', 'mcat_packaging', 2, 3, NOW(), NOW()),
('mcat_pk_try', 'PK-TRY', '托盘/衬垫',  'mcat_packaging', 'mcat_packaging', 2, 4, NOW(), NOW()),
('mcat_pk_lbl', 'PK-LBL', '标签',       'mcat_packaging', 'mcat_packaging', 2, 5, NOW(), NOW());

-- 7. 二级分类 - 辅料 (AX)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_ax_adh', 'AX-ADH', '胶粘剂',     'mcat_auxiliary', 'mcat_auxiliary', 2, 1, NOW(), NOW()),
('mcat_ax_sld', 'AX-SLD', '焊接材料',   'mcat_auxiliary', 'mcat_auxiliary', 2, 2, NOW(), NOW()),
('mcat_ax_cln', 'AX-CLN', '清洗材料',   'mcat_auxiliary', 'mcat_auxiliary', 2, 3, NOW(), NOW()),
('mcat_ax_ins', 'AX-INS', '绝缘材料',   'mcat_auxiliary', 'mcat_auxiliary', 2, 4, NOW(), NOW()),
('mcat_ax_tls', 'AX-TLS', '工装治具',   'mcat_auxiliary', 'mcat_auxiliary', 2, 5, NOW(), NOW()),
('mcat_ax_oth', 'AX-OTH', '其他辅料',   'mcat_auxiliary', 'mcat_auxiliary', 2, 6, NOW(), NOW());

-- 8. 二级分类 - 成品 (FG)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_fg_gls', 'FG-GLS', '智能眼镜',   'mcat_finished', 'mcat_finished', 2, 1, NOW(), NOW()),
('mcat_fg_acc', 'FG-ACC', '配件',       'mcat_finished', 'mcat_finished', 2, 2, NOW(), NOW()),
('mcat_fg_set', 'FG-SET', '套装',       'mcat_finished', 'mcat_finished', 2, 3, NOW(), NOW());

-- 9. 二级分类 - 半成品 (SA)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_sa_pcb', 'SA-PCB', 'PCBA总成',    'mcat_subassy', 'mcat_subassy', 2, 1, NOW(), NOW()),
('mcat_sa_opt', 'SA-OPT', '光学模组总成', 'mcat_subassy', 'mcat_subassy', 2, 2, NOW(), NOW()),
('mcat_sa_arm', 'SA-ARM', '镜腿总成',    'mcat_subassy', 'mcat_subassy', 2, 3, NOW(), NOW()),
('mcat_sa_dsp', 'SA-DSP', '显示总成',    'mcat_subassy', 'mcat_subassy', 2, 4, NOW(), NOW()),
('mcat_sa_oth', 'SA-OTH', '其他组件',    'mcat_subassy', 'mcat_subassy', 2, 5, NOW(), NOW());

-- 10. 二级分类 - 软件 (SW)
INSERT INTO material_categories (id, code, name, parent_id, path, level, sort_order, created_at, updated_at) VALUES
('mcat_sw_fw',  'SW-FW',  '固件',       'mcat_software', 'mcat_software', 2, 1, NOW(), NOW()),
('mcat_sw_app', 'SW-APP', '应用软件',   'mcat_software', 'mcat_software', 2, 2, NOW(), NOW()),
('mcat_sw_lic', 'SW-LIC', '授权/许可',  'mcat_software', 'mcat_software', 2, 3, NOW(), NOW());

COMMIT;
