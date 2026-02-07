#!/bin/bash
# 填充V2任务模板到PLM系统
set -e

BASE="http://127.0.0.1:8080/api/v1"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJkNDA2NzA3Yi1mOWJiLTRlOWYtOGNjNS0wY2FiMzhlNCIsIm5hbWUiOiLpmYjms73mlowiLCJlbWFpbCI6InplYmluQGJpdGZhbnRhc3kuaW8iLCJmZWlzaHVfdWlkIjoib3VfNWIxNTlmYzE1N2Q0MDQyZjFlODA4OGIxZmZlYmIyZGEiLCJyb2xlcyI6WyJhZG1pbiJdLCJwZXJtcyI6WyIqIl0sImlzcyI6Im5pbW8tcGxtIiwiZXhwIjoxODAxODc1OTgwLCJpYXQiOjE3NzAzMzk5ODB9.ZhEpSPJ1VlJGrdxOmjmv-PZWNrCnf7e_etdbxohQElk"
H="Authorization: Bearer $TOKEN"

api() {
  curl -s -X "$1" -H "$H" -H "Content-Type: application/json" "$BASE$2" -d "$3"
}

echo "=========================================="
echo "  填充 nimo PLM 研发任务模板 V2.0"
echo "=========================================="

#-----------------------------------------------------------
# 1. 创建平台项目模板
#-----------------------------------------------------------
echo ""
echo ">>> 创建平台项目模板..."
RESULT=$(api POST "/templates" '{
  "code": "TPL-PLATFORM-V1",
  "name": "平台项目研发模板 V1.0",
  "description": "nimo智能眼镜平台项目全流程模板，对标华为/OPPO消费电子研发体系。包含Concept/EVT/DVT/PVT/MP五个阶段，共140+项任务。",
  "product_type": "platform",
  "phases": ["concept","evt","dvt","pvt","mp"],
  "estimated_days": 200
}')
TPL_ID=$(echo "$RESULT" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "  模板ID: $TPL_ID"

if [ -z "$TPL_ID" ]; then
  echo "  创建失败，尝试获取已有模板..."
  RESULT=$(api GET "/templates" "")
  TPL_ID=$(echo "$RESULT" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  echo "  已有模板ID: $TPL_ID"
fi

task() {
  local code="$1" name="$2" phase="$3" parent="$4" role="$5" days="$6" critical="$7" type="$8" sort="$9"
  [ -z "$type" ] && type="TASK"
  [ -z "$sort" ] && sort=0
  [ "$critical" = "true" ] && crit="true" || crit="false"
  
  api POST "/templates/$TPL_ID/tasks" "{
    \"task_code\": \"$code\",
    \"name\": \"$name\",
    \"phase\": \"$phase\",
    \"parent_task_code\": \"$parent\",
    \"default_assignee_role\": \"$role\",
    \"estimated_days\": $days,
    \"is_critical\": $crit,
    \"task_type\": \"$type\",
    \"sort_order\": $sort
  }" > /dev/null
  echo "    ✓ $code $name"
}

#-----------------------------------------------------------
# CONCEPT 阶段
#-----------------------------------------------------------
echo ""
echo ">>> Concept/立项 阶段..."

# 父任务
task "C-10" "市场与产品定义"       "concept" "" "product_manager" 10 true "MILESTONE" 10
task "C-20" "技术可行性评估"       "concept" "" "tech_lead"       8  true "MILESTONE" 20
task "C-30" "物料与成本预估"       "concept" "" "hw_lead"         8  true "MILESTONE" 30
task "C-40" "项目规划"             "concept" "" "project_manager" 7  true "MILESTONE" 40
task "C-99" "G0 立项评审"          "concept" "" "project_manager" 1  true "MILESTONE" 99

# C-10 子任务
task "C-11" "市场需求调研与竞品分析"     "concept" "C-10" "product_manager" 5  true  "TASK" 11
task "C-12" "产品定义与功能规格制定"     "concept" "C-10" "product_manager" 5  true  "TASK" 12

# C-20 子任务
task "C-21" "光学方案可行性评估"         "concept" "C-20" "optical_engineer" 5  true  "TASK" 21
task "C-22" "显示方案可行性评估"         "concept" "C-20" "hw_engineer"      5  true  "TASK" 22
task "C-23" "结构空间预评估"             "concept" "C-20" "me_engineer"      3  true  "TASK" 23

# C-30 子任务
task "C-31" "主要元器件预选型"           "concept" "C-30" "hw_engineer"      5  true  "TASK" 31
task "C-32" "长交期物料识别与预采购评估" "concept" "C-30" "procurement"      3  true  "TASK" 32
task "C-33" "初步成本估算"               "concept" "C-30" "product_manager"  3  true  "TASK" 33
task "C-34" "概念BOM编制(cBOM)"          "concept" "C-30" "hw_lead"          3  true  "TASK" 34

# C-40 子任务
task "C-41" "平台选型/新建决策"          "concept" "C-40" "project_manager"  2  true  "TASK" 41
task "C-42" "项目计划编制"               "concept" "C-40" "project_manager"  3  true  "TASK" 42
task "C-43" "项目团队组建"               "concept" "C-40" "project_manager"  2  true  "TASK" 43
task "C-44" "认证需求识别"               "concept" "C-40" "cert_engineer"    3  false "TASK" 44
task "C-45" "知识产权预检索"             "concept" "C-40" "product_manager"  5  false "TASK" 45

#-----------------------------------------------------------
# EVT 阶段
#-----------------------------------------------------------
echo ""
echo ">>> EVT 阶段..."

# 父任务
task "E-10" "硬件/电子设计"     "evt" "" "hw_lead"       30 true "MILESTONE" 10
task "E-20" "光学设计"           "evt" "" "optical_lead"  30 true "MILESTONE" 20
task "E-30" "结构/机械设计"     "evt" "" "me_lead"       30 true "MILESTONE" 30
task "E-40" "软件/固件"         "evt" "" "sw_lead"       25 true "MILESTONE" 40
task "E-50" "样机制作与测试"    "evt" "" "test_lead"     20 true "MILESTONE" 50
task "E-60" "BOM与采购"         "evt" "" "hw_lead"       15 true "MILESTONE" 60
task "E-90" "EVT评审"           "evt" "" "project_manager" 5 true "MILESTONE" 90

# E-10 硬件设计子任务
task "E-11"  "系统架构设计"            "evt" "E-10" "hw_lead"        5  true  "TASK" 11
task "E-12"  "主板设计"                "evt" "E-10" "hw_engineer"    22 true  "TASK" 12
task "E-12a" "主板原理图设计"          "evt" "E-12" "hw_engineer"    10 true  "SUBTASK" 1
task "E-12b" "主板原理图评审"          "evt" "E-12" "hw_lead"        1  true  "SUBTASK" 2
task "E-12c" "主板PCB Layout"          "evt" "E-12" "layout_engineer" 10 true "SUBTASK" 3
task "E-12d" "主板PCB评审(DFM)"        "evt" "E-12" "hw_lead"        1  true  "SUBTASK" 4
task "E-13"  "FPC设计"                 "evt" "E-10" "hw_engineer"    10 true  "TASK" 13
task "E-13a" "主FPC排线设计"           "evt" "E-13" "hw_engineer"    8  true  "SUBTASK" 1
task "E-13b" "充电FPC设计"             "evt" "E-13" "hw_engineer"    5  true  "SUBTASK" 2
task "E-14"  "器件选型验证"            "evt" "E-10" "hw_engineer"    10 true  "TASK" 14
task "E-14a" "电池规格确认与选型"      "evt" "E-14" "hw_engineer"    5  true  "SUBTASK" 1
task "E-14b" "传感器选型验证"          "evt" "E-14" "hw_engineer"    5  true  "SUBTASK" 2
task "E-14c" "磁吸充电方案设计"        "evt" "E-14" "hw_engineer"    5  false "SUBTASK" 3
task "E-15"  "蓝牙左右耳通信方案验证"  "evt" "E-10" "hw_engineer"    5  true  "TASK" 15

# E-20 光学设计子任务
task "E-21"  "波导设计"                "evt" "E-20" "optical_engineer" 20 true "TASK" 21
task "E-21a" "波导光学设计"            "evt" "E-21" "optical_engineer" 15 true "SUBTASK" 1
task "E-21b" "光学仿真与优化"          "evt" "E-21" "optical_engineer" 10 true "SUBTASK" 2
task "E-22"  "光机设计"                "evt" "E-20" "optical_engineer" 12 true "TASK" 22
task "E-22a" "光机模组方案设计"        "evt" "E-22" "optical_engineer" 10 true "SUBTASK" 1
task "E-22b" "MicroLED芯片规格确认"    "evt" "E-22" "optical_engineer" 5  true "SUBTASK" 2
task "E-23"  "光学样品加工与验证"      "evt" "E-20" "optical_engineer" 15 true "TASK" 23

# E-30 结构设计子任务
task "E-31"  "ID设计(外观造型)"        "evt" "E-30" "id_designer"     10 true  "TASK" 31
task "E-32"  "结构件设计"              "evt" "E-30" "me_engineer"     18 true  "TASK" 32
task "E-32a" "镜腿结构设计"            "evt" "E-32" "me_engineer"     10 true  "SUBTASK" 1
task "E-32b" "桩头结构设计"            "evt" "E-32" "me_engineer"     8  true  "SUBTASK" 2
task "E-32c" "镜架结构设计"            "evt" "E-32" "me_engineer"     8  false "SUBTASK" 3
task "E-32d" "脚套设计"                "evt" "E-32" "me_engineer"     3  false "SUBTASK" 4
task "E-33"  "堆叠设计(空间布局)"      "evt" "E-30" "me_engineer"     10 true  "TASK" 33
task "E-34"  "结构评审(DFA)"           "evt" "E-30" "me_lead"         1  true  "TASK" 34
task "E-35"  "手板/CNC快速样件"        "evt" "E-30" "me_engineer"     10 true  "TASK" 35
task "E-36"  "防水等级方案设计"        "evt" "E-30" "me_engineer"     5  false "TASK" 36

# E-40 软件/固件子任务
task "E-41"  "固件架构设计"            "evt" "E-40" "sw_engineer"     5  true  "TASK" 41
task "E-42"  "驱动开发"                "evt" "E-40" "sw_engineer"     25 true  "TASK" 42
task "E-42a" "BSP/驱动开发"            "evt" "E-42" "sw_engineer"     15 true  "SUBTASK" 1
task "E-42b" "蓝牙协议栈适配"          "evt" "E-42" "sw_engineer"     10 true  "SUBTASK" 2
task "E-42c" "传感器驱动开发"          "evt" "E-42" "sw_engineer"     8  false "SUBTASK" 3
task "E-42d" "充电管理固件"            "evt" "E-42" "sw_engineer"     5  false "SUBTASK" 4
task "E-43"  "基础功能联调"            "evt" "E-40" "sw_engineer"     5  true  "TASK" 43

# E-50 样机与测试子任务
task "E-51"  "样机制作"                "evt" "E-50" "hw_engineer"     12 true  "TASK" 51
task "E-51a" "PCB打板与贴片"           "evt" "E-51" "hw_engineer"     10 true  "SUBTASK" 1
task "E-51b" "EVT样机组装(5-20台)"     "evt" "E-51" "hw_engineer"     5  true  "SUBTASK" 2
task "E-52"  "EVT测试"                 "evt" "E-50" "test_engineer"   12 true  "TASK" 52
task "E-52a" "硬件功能测试"            "evt" "E-52" "test_engineer"   5  true  "SUBTASK" 1
task "E-52b" "光学性能测试"            "evt" "E-52" "optical_engineer" 5 true  "SUBTASK" 2
task "E-52c" "功耗测试"                "evt" "E-52" "hw_engineer"     3  true  "SUBTASK" 3
task "E-52d" "热仿真/散热测试"         "evt" "E-52" "hw_engineer"     3  true  "SUBTASK" 4
task "E-52e" "佩戴舒适度评估"          "evt" "E-52" "product_manager" 3  false "SUBTASK" 5

# E-60 BOM与采购子任务
task "E-61"  "EVT BOM编制(eBOM)"       "evt" "E-60" "hw_engineer"     3  true  "TASK" 61
task "E-62"  "EVT BOM审批"             "evt" "E-60" "hw_lead"         1  true  "TASK" 62
task "E-63"  "EVT物料采购"             "evt" "E-60" "procurement"     15 true  "TASK" 63
task "E-64"  "EVT成本核算"             "evt" "E-60" "procurement"     3  false "TASK" 64

# E-90 评审子任务
task "E-91"  "EVT问题汇总与跟踪"      "evt" "E-90" "project_manager" 2  true  "TASK" 91
task "E-92"  "EVT阶段总结报告"         "evt" "E-90" "project_manager" 2  true  "TASK" 92
task "E-99"  "G1 EVT退出评审"          "evt" "E-90" "project_manager" 1  true  "MILESTONE" 99

#-----------------------------------------------------------
# DVT 阶段
#-----------------------------------------------------------
echo ""
echo ">>> DVT 阶段..."

# 父任务
task "D-10" "设计优化"          "dvt" "" "hw_lead"        20 true "MILESTONE" 10
task "D-20" "开模与试制"        "dvt" "" "me_lead"        30 true "MILESTONE" 20
task "D-30" "可靠性测试"        "dvt" "" "test_lead"      20 true "MILESTONE" 30
task "D-40" "认证预测试"        "dvt" "" "cert_engineer"  15 true "MILESTONE" 40
task "D-50" "BOM与成本"         "dvt" "" "hw_lead"        8  true "MILESTONE" 50
task "D-90" "DVT评审"           "dvt" "" "project_manager" 5 true "MILESTONE" 90

# D-10 设计优化子任务
task "D-11"  "硬件优化"                "dvt" "D-10" "hw_engineer"     18 true  "TASK" 11
task "D-11a" "EVT问题修复-硬件"        "dvt" "D-11" "hw_engineer"     8  true  "SUBTASK" 1
task "D-11b" "主板原理图更新(Rev B)"   "dvt" "D-11" "hw_engineer"     5  true  "SUBTASK" 2
task "D-11c" "主板PCB Layout更新"      "dvt" "D-11" "layout_engineer" 5  true  "SUBTASK" 3
task "D-11d" "FPC优化设计"             "dvt" "D-11" "hw_engineer"     5  true  "SUBTASK" 4
task "D-12"  "光学设计优化"            "dvt" "D-10" "optical_engineer" 10 true "TASK" 12
task "D-13"  "结构优化"                "dvt" "D-10" "me_engineer"     12 true  "TASK" 13
task "D-13a" "结构设计优化"            "dvt" "D-13" "me_engineer"     8  true  "SUBTASK" 1
task "D-13b" "镜架设计定稿"            "dvt" "D-13" "me_engineer"     8  false "SUBTASK" 2
task "D-14"  "固件功能完善"            "dvt" "D-10" "sw_engineer"     15 true  "TASK" 14

# D-20 开模与试制子任务
task "D-21"  "模具开发"                "dvt" "D-20" "me_engineer"     30 true  "TASK" 21
task "D-21a" "模具开发(镜腿/桩头)"     "dvt" "D-21" "me_engineer"     25 true  "SUBTASK" 1
task "D-21b" "T1模具试模"              "dvt" "D-21" "me_engineer"     3  true  "SUBTASK" 2
task "D-21c" "T1验收与修模"            "dvt" "D-21" "me_engineer"     5  true  "SUBTASK" 3
task "D-21d" "T2模具试模"              "dvt" "D-21" "me_engineer"     3  false "SUBTASK" 4
task "D-22"  "DVT样机"                 "dvt" "D-20" "hw_engineer"     13 true  "TASK" 22
task "D-22a" "DVT PCB打板贴片"         "dvt" "D-22" "hw_engineer"     8  true  "SUBTASK" 1
task "D-22b" "DVT样机组装(50-200台)"   "dvt" "D-22" "hw_engineer"     8  true  "SUBTASK" 2

# D-30 可靠性测试子任务
task "D-31"  "环境与机械测试"          "dvt" "D-30" "test_engineer"   10 true  "TASK" 31
task "D-31a" "跌落测试"                "dvt" "D-31" "test_engineer"   3  true  "SUBTASK" 1
task "D-31b" "温湿度循环测试"          "dvt" "D-31" "test_engineer"   5  true  "SUBTASK" 2
task "D-31c" "盐雾测试"                "dvt" "D-31" "test_engineer"   5  false "SUBTASK" 3
task "D-31d" "防水测试(IPX验证)"       "dvt" "D-31" "test_engineer"   3  true  "SUBTASK" 4
task "D-32"  "耐久与寿命测试"          "dvt" "D-30" "test_engineer"   15 true  "TASK" 32
task "D-32a" "按键/触控寿命测试"       "dvt" "D-32" "test_engineer"   5  true  "SUBTASK" 1
task "D-32b" "充电循环寿命测试"        "dvt" "D-32" "test_engineer"   10 true  "SUBTASK" 2
task "D-32c" "铰链(桩头)耐久测试"      "dvt" "D-32" "test_engineer"   5  true  "SUBTASK" 3
task "D-33"  "性能复测"                "dvt" "D-30" "test_engineer"   8  true  "TASK" 33
task "D-33a" "硬件功能复测"            "dvt" "D-33" "test_engineer"   3  true  "SUBTASK" 1
task "D-33b" "光学性能复测"            "dvt" "D-33" "optical_engineer" 3 true  "SUBTASK" 2
task "D-33c" "软件系统测试"            "dvt" "D-33" "test_engineer"   5  true  "SUBTASK" 3
task "D-34"  "EMC与安全测试"           "dvt" "D-30" "hw_engineer"     6  true  "TASK" 34
task "D-34a" "ESD/EMC预测试"           "dvt" "D-34" "hw_engineer"     3  true  "SUBTASK" 1
task "D-34b" "SAR测试(头部辐射安全)"   "dvt" "D-34" "cert_engineer"   5  true  "SUBTASK" 2
task "D-35"  "用户体验测试(内部试用)"  "dvt" "D-30" "product_manager" 5  false "TASK" 35

# D-40 认证预测试
task "D-41"  "FCC认证预测试"           "dvt" "D-40" "cert_engineer"   5  true  "TASK" 41
task "D-42"  "CE认证预测试"            "dvt" "D-40" "cert_engineer"   5  true  "TASK" 42
task "D-43"  "电池安全认证(UN38.3)"    "dvt" "D-40" "cert_engineer"   10 true  "TASK" 43
task "D-44"  "其他市场认证识别"        "dvt" "D-40" "cert_engineer"   3  false "TASK" 44

# D-50 BOM与成本
task "D-51"  "DVT BOM编制(dBOM)"       "dvt" "D-50" "hw_engineer"     3  true  "TASK" 51
task "D-52"  "DVT BOM审批"             "dvt" "D-50" "hw_lead"         1  true  "TASK" 52
task "D-53"  "DVT成本核算与优化"       "dvt" "D-50" "procurement"     5  true  "TASK" 53

# D-90 评审
task "D-91"  "DVT问题汇总与闭环跟踪"  "dvt" "D-90" "project_manager" 2  true  "TASK" 91
task "D-92"  "DVT阶段总结报告"         "dvt" "D-90" "project_manager" 2  true  "TASK" 92
task "D-99"  "G2 DVT退出评审"          "dvt" "D-90" "project_manager" 1  true  "MILESTONE" 99

#-----------------------------------------------------------
# PVT 阶段
#-----------------------------------------------------------
echo ""
echo ">>> PVT 阶段..."

# 父任务
task "P-10" "产线准备"          "pvt" "" "process_lead"    15 true "MILESTONE" 10
task "P-20" "试产"              "pvt" "" "process_lead"    20 true "MILESTONE" 20
task "P-30" "认证正式送测"      "pvt" "" "cert_engineer"   15 true "MILESTONE" 30
task "P-40" "包装与配件"        "pvt" "" "product_manager" 10 true "MILESTONE" 40
task "P-50" "BOM与成本"         "pvt" "" "hw_lead"         5  true "MILESTONE" 50
task "P-90" "PVT评审"           "pvt" "" "project_manager" 5  true "MILESTONE" 90

# P-10 产线准备
task "P-11"  "产线规划"                "pvt" "P-10" "process_engineer" 15 true "TASK" 11
task "P-11a" "产线工艺规划"            "pvt" "P-11" "process_engineer" 5  true "SUBTASK" 1
task "P-11b" "SOP/作业指导书编写"      "pvt" "P-11" "process_engineer" 8  true "SUBTASK" 2
task "P-11c" "治具/工装设计制作"       "pvt" "P-11" "process_engineer" 10 true "SUBTASK" 3
task "P-12"  "品质准备"                "pvt" "P-10" "qa_engineer"      6  true "TASK" 12
task "P-12a" "检验标准制定(IQC/IPQC/OQC)" "pvt" "P-12" "qa_engineer"  5  true "SUBTASK" 1
task "P-12b" "Golden Sample制作"       "pvt" "P-12" "qa_engineer"      3  true "SUBTASK" 2
task "P-13"  "产线工人培训"            "pvt" "P-10" "process_engineer" 3  false "TASK" 13

# P-20 试产
task "P-21"  "物料与生产"              "pvt" "P-20" "process_engineer" 20 true "TASK" 21
task "P-21a" "PVT物料采购"             "pvt" "P-21" "procurement"      15 true "SUBTASK" 1
task "P-21b" "SMT贴片试产"             "pvt" "P-21" "process_engineer" 3  true "SUBTASK" 2
task "P-21c" "整机组装试产(500-2000台)" "pvt" "P-21" "process_engineer" 8 true "SUBTASK" 3
task "P-22"  "良率分析"                "pvt" "P-20" "qa_engineer"      8  true "TASK" 22
task "P-22a" "良率统计与分析"          "pvt" "P-22" "qa_engineer"      3  true "SUBTASK" 1
task "P-22b" "不良分析与改善"          "pvt" "P-22" "qa_engineer"      5  true "SUBTASK" 2
task "P-22c" "生产节拍验证"            "pvt" "P-22" "process_engineer" 2  true "SUBTASK" 3

# P-30 认证
task "P-31"  "FCC正式认证送测"         "pvt" "P-30" "cert_engineer"    15 true "TASK" 31
task "P-32"  "CE正式认证送测"          "pvt" "P-30" "cert_engineer"    15 true "TASK" 32
task "P-33"  "蓝牙BQB认证"             "pvt" "P-30" "cert_engineer"    10 true "TASK" 33
task "P-34"  "其他区域认证"            "pvt" "P-30" "cert_engineer"    15 false "TASK" 34

# P-40 包装
task "P-41"  "包装设计定稿"            "pvt" "P-40" "product_manager"  5  true "TASK" 41
task "P-42"  "包装打样与验证"          "pvt" "P-40" "product_manager"  5  true "TASK" 42
task "P-43"  "配件确认(充电线/说明书)" "pvt" "P-40" "product_manager"  3  false "TASK" 43

# P-50 BOM
task "P-51"  "PVT BOM编制(pBOM)"       "pvt" "P-50" "hw_engineer"     2  true "TASK" 51
task "P-52"  "PVT BOM审批"             "pvt" "P-50" "hw_lead"         1  true "TASK" 52
task "P-53"  "量产成本核算"            "pvt" "P-50" "procurement"     3  true "TASK" 53

# P-90 评审
task "P-91"  "PVT问题汇总与闭环"      "pvt" "P-90" "project_manager" 2  true "TASK" 91
task "P-92"  "PVT阶段总结报告"         "pvt" "P-90" "project_manager" 2  true "TASK" 92
task "P-99"  "G3 PVT退出评审"          "pvt" "P-90" "project_manager" 1  true "MILESTONE" 99

#-----------------------------------------------------------
# MP 阶段
#-----------------------------------------------------------
echo ""
echo ">>> MP 阶段..."

# 父任务
task "M-10" "量产准备"          "mp" "" "hw_lead"         15 true "MILESTONE" 10
task "M-20" "首批量产与验证"    "mp" "" "process_lead"    12 true "MILESTONE" 20
task "M-30" "持续运营"          "mp" "" "project_manager" 10 false "MILESTONE" 30
task "M-99" "G4 量产放行评审"   "mp" "" "project_manager" 1  true "MILESTONE" 99

# M-10 量产准备
task "M-11"  "量产BOM定稿(mBOM)冻结"   "mp" "M-10" "hw_lead"     2  true "TASK" 11
task "M-12"  "量产物料备料"             "mp" "M-10" "procurement" 15 true "TASK" 12

# M-20 首批量产
task "M-21"  "首批量产(Pilot Run)"      "mp" "M-20" "process_engineer" 5  true "TASK" 21
task "M-22"  "首批OQC全检"              "mp" "M-20" "qa_engineer"      3  true "TASK" 22
task "M-23"  "出货前可靠性抽检(ORT)"    "mp" "M-20" "qa_engineer"      5  true "TASK" 23
task "M-24"  "量产放行判定"             "mp" "M-20" "qa_lead"          1  true "TASK" 24

# M-30 持续运营
task "M-31"  "产能爬坡计划制定"         "mp" "M-30" "process_engineer" 3  false "TASK" 31
task "M-32"  "售后维修方案编制"         "mp" "M-30" "qa_engineer"      5  false "TASK" 32
task "M-33"  "量产周报机制建立"         "mp" "M-30" "project_manager"  2  false "TASK" 33
task "M-34"  "持续良率监控"             "mp" "M-30" "qa_engineer"      1  false "TASK" 34
task "M-35"  "客诉/市场反馈跟踪"        "mp" "M-30" "qa_engineer"      1  false "TASK" 35

echo ""
echo "=========================================="
echo "  任务创建完成！开始创建依赖关系..."
echo "=========================================="

#-----------------------------------------------------------
# 依赖关系 (前序任务)
#-----------------------------------------------------------

dep() {
  local task_code="$1" depends_on="$2"
  api POST "/templates/$TPL_ID/tasks/$task_code/dependencies" "{
    \"depends_on_task_code\": \"$depends_on\",
    \"dependency_type\": \"FS\"
  }" > /dev/null 2>&1 || true
  echo "    $task_code → $depends_on"
}

echo ""
echo ">>> Concept 依赖关系..."
dep "C-12" "C-11"
dep "C-21" "C-12"
dep "C-22" "C-12"
dep "C-23" "C-12"
dep "C-31" "C-12"
dep "C-32" "C-31"
dep "C-33" "C-31"
dep "C-33" "C-32"
dep "C-34" "C-31"
dep "C-41" "C-21"
dep "C-41" "C-22"
dep "C-42" "C-21"
dep "C-42" "C-22"
dep "C-42" "C-23"
dep "C-42" "C-33"
dep "C-43" "C-42"
dep "C-44" "C-12"
dep "C-45" "C-12"
dep "C-99" "C-12"
dep "C-99" "C-21"
dep "C-99" "C-22"
dep "C-99" "C-23"
dep "C-99" "C-33"
dep "C-99" "C-34"
dep "C-99" "C-42"
dep "C-99" "C-43"

echo ""
echo ">>> EVT 依赖关系..."
dep "E-11"  "C-99"
dep "E-12a" "E-11"
dep "E-12b" "E-12a"
dep "E-12c" "E-12b"
dep "E-12d" "E-12c"
dep "E-13a" "E-11"
dep "E-13b" "E-11"
dep "E-14a" "E-11"
dep "E-14b" "E-11"
dep "E-14c" "E-13b"
dep "E-15"  "E-12a"
dep "E-21a" "C-99"
dep "E-21b" "E-21a"
dep "E-22a" "C-99"
dep "E-22b" "E-22a"
dep "E-23"  "E-21b"
dep "E-31"  "C-99"
dep "E-32a" "E-31"
dep "E-32a" "E-11"
dep "E-32b" "E-31"
dep "E-32c" "E-31"
dep "E-32d" "E-32a"
dep "E-33"  "E-32a"
dep "E-33"  "E-12c"
dep "E-33"  "E-22a"
dep "E-34"  "E-33"
dep "E-35"  "E-34"
dep "E-36"  "E-32a"
dep "E-41"  "E-11"
dep "E-42a" "E-41"
dep "E-42b" "E-41"
dep "E-42c" "E-42a"
dep "E-42d" "E-42a"
dep "E-43"  "E-42a"
dep "E-43"  "E-42b"
dep "E-51a" "E-12d"
dep "E-51b" "E-51a"
dep "E-51b" "E-35"
dep "E-51b" "E-23"
dep "E-52a" "E-51b"
dep "E-52b" "E-51b"
dep "E-52c" "E-51b"
dep "E-52d" "E-51b"
dep "E-52e" "E-51b"
dep "E-61"  "E-12b"
dep "E-61"  "E-22b"
dep "E-61"  "E-14a"
dep "E-61"  "E-14b"
dep "E-62"  "E-61"
dep "E-63"  "E-62"
dep "E-64"  "E-63"
dep "E-91"  "E-52a"
dep "E-91"  "E-52b"
dep "E-91"  "E-52c"
dep "E-91"  "E-52d"
dep "E-92"  "E-91"
dep "E-99"  "E-92"
dep "E-99"  "E-62"

echo ""
echo ">>> DVT 依赖关系..."
dep "D-11a" "E-99"
dep "D-11b" "D-11a"
dep "D-11c" "D-11b"
dep "D-11d" "D-11a"
dep "D-12"  "E-99"
dep "D-13a" "E-99"
dep "D-13b" "D-13a"
dep "D-14"  "E-99"
dep "D-21a" "D-13a"
dep "D-21b" "D-21a"
dep "D-21c" "D-21b"
dep "D-21d" "D-21c"
dep "D-22a" "D-11c"
dep "D-22b" "D-22a"
dep "D-22b" "D-21b"
dep "D-22b" "D-12"
dep "D-31a" "D-22b"
dep "D-31b" "D-22b"
dep "D-31c" "D-22b"
dep "D-31d" "D-22b"
dep "D-32a" "D-22b"
dep "D-32b" "D-22b"
dep "D-32c" "D-22b"
dep "D-33a" "D-22b"
dep "D-33b" "D-22b"
dep "D-33c" "D-14"
dep "D-34a" "D-22b"
dep "D-34b" "D-22b"
dep "D-35"  "D-22b"
dep "D-41"  "D-22b"
dep "D-42"  "D-22b"
dep "D-43"  "D-22b"
dep "D-44"  "E-99"
dep "D-51"  "D-11b"
dep "D-51"  "D-12"
dep "D-52"  "D-51"
dep "D-53"  "D-52"
dep "D-91"  "D-31a"
dep "D-91"  "D-32a"
dep "D-91"  "D-33a"
dep "D-91"  "D-41"
dep "D-92"  "D-91"
dep "D-99"  "D-92"
dep "D-99"  "D-52"

echo ""
echo ">>> PVT 依赖关系..."
dep "P-11a" "D-99"
dep "P-11b" "P-11a"
dep "P-11c" "P-11a"
dep "P-12a" "P-11a"
dep "P-12b" "D-22b"
dep "P-13"  "P-11b"
dep "P-21a" "D-52"
dep "P-21b" "P-21a"
dep "P-21b" "P-11c"
dep "P-21c" "P-21b"
dep "P-22a" "P-21c"
dep "P-22b" "P-22a"
dep "P-22c" "P-21c"
dep "P-31"  "P-21c"
dep "P-32"  "P-21c"
dep "P-33"  "P-21c"
dep "P-34"  "P-21c"
dep "P-41"  "D-99"
dep "P-42"  "P-41"
dep "P-43"  "D-99"
dep "P-51"  "D-52"
dep "P-52"  "P-51"
dep "P-53"  "P-52"
dep "P-91"  "P-22b"
dep "P-91"  "P-31"
dep "P-91"  "P-32"
dep "P-91"  "P-33"
dep "P-92"  "P-91"
dep "P-99"  "P-92"
dep "P-99"  "P-52"
dep "P-99"  "P-31"
dep "P-99"  "P-32"
dep "P-99"  "P-33"

echo ""
echo ">>> MP 依赖关系..."
dep "M-11"  "P-99"
dep "M-12"  "M-11"
dep "M-21"  "M-12"
dep "M-22"  "M-21"
dep "M-23"  "M-21"
dep "M-24"  "M-22"
dep "M-24"  "M-23"
dep "M-31"  "M-24"
dep "M-32"  "P-99"
dep "M-33"  "M-24"
dep "M-34"  "M-24"
dep "M-35"  "M-24"
dep "M-99"  "M-24"

echo ""
echo "=========================================="
echo "  ✅ 全部完成!"
echo "  模板ID: $TPL_ID"
echo "=========================================="
