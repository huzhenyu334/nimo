# MODULES.md - PLM/SRM 功能模块清单

CC（Claude Code）每次开发新功能前必须读此文件，完成后必须更新。
防止不同session重复开发同一功能。

## PLM 模块

| 路由 | 菜单名 | 页面组件 | 状态 | 说明 |
|------|--------|----------|------|------|
| /dashboard | 工作台 | Dashboard | ✅ | |
| /my-tasks | 我的任务 | MyTasks | ✅ | |
| /projects | 项目管理 | Projects | ✅ | |
| /projects/:id | 项目详情 | ProjectDetail | ✅ | |
| /bom-management | BOM管理 | BOMManagement | ✅ | |
| /bom-management/:projectId | BOM详情 | BOMManagementDetail | ✅ | |
| /ecn | ECN变更管理 | ECNList/ECNDetail/ECNForm | ✅ | 完整的ECN管理 |
| /materials | 物料选型库 | Materials | ✅ | 物料管理主入口 |
| /templates | 流程管理 | Templates | ✅ | |
| /approvals | 审批管理 | Approvals | ✅ | |
| /documents | 文档管理 | Documents | ✅ | |
| /roles | 角色管理 | RoleManagement | ✅ | |

## SRM 模块

| 路由 | 菜单名 | 页面组件 | 状态 | 说明 |
|------|--------|----------|------|------|
| /srm | 采购总览 | SRMDashboard | ✅ | |
| /srm/suppliers | 供应商 | Suppliers | ✅ | |
| /srm/purchase-requests | 采购需求 | PurchaseRequests | ✅ | |
| /srm/purchase-orders | 采购订单 | PurchaseOrders | ✅ | |
| /srm/inspections | 来料检验 | Inspections | ✅ | |
| /srm/inventory | 库存管理 | Inventory | ✅ | |
| /srm/settlements | 对账结算 | Settlements | ✅ | |
| /srm/corrective-actions | 8D改进 | CorrectiveActions | ✅ | |
| /srm/evaluations | 供应商评价 | Evaluations | ✅ | |
| /srm/equipment | 通用设备 | SRMEquipment | ✅ | |
