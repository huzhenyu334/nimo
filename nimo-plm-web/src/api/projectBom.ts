import apiClient from './client';
import { ApiResponse } from '@/types';

// 项目BOM相关类型
export interface ProjectBOM {
  id: string;
  project_id: string;
  phase_id?: string;
  name: string;
  bom_type: string;
  version: string;
  status: 'draft' | 'pending_review' | 'published' | 'rejected' | 'frozen';
  description?: string;
  submitted_by?: string;
  submitted_at?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  review_comment?: string;
  approved_by?: string;
  approved_at?: string;
  frozen_at?: string;
  frozen_by?: string;
  total_items: number;
  estimated_cost?: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  // Relations
  phase?: { id: string; phase: string; name?: string };
  creator?: { id: string; name: string; avatar_url?: string };
  submitter?: { id: string; name: string };
  reviewer?: { id: string; name: string };
}

export interface ProjectBOMItem {
  id: string;
  bom_id: string;
  item_number: number;
  parent_item_id?: string;
  level: number;
  material_id?: string;
  category: string;
  name: string;
  specification?: string;
  quantity: number;
  unit: string;
  reference?: string;
  manufacturer?: string;
  manufacturer_pn?: string;
  supplier?: string;
  supplier_pn?: string;
  unit_price?: number;
  extended_cost?: number;
  lead_time_days?: number;
  procurement_type: string;
  moq?: number;
  approved_vendors?: string;
  lifecycle_status: string;
  is_critical: boolean;
  is_alternative: boolean;
  alternative_for?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
  // Relations
  material?: { id: string; name: string; code: string; specification?: string };
  children?: ProjectBOMItem[];
}

export interface ProjectBOMDetail extends ProjectBOM {
  items: ProjectBOMItem[];
}

export interface CreateProjectBOMRequest {
  name: string;
  bom_type: string;
  phase_id?: string;
  version?: string;
  description?: string;
}

export interface BOMItemRequest {
  material_id?: string;
  parent_item_id?: string;
  level?: number;
  category?: string;
  name: string;
  specification?: string;
  quantity: number;
  unit?: string;
  reference?: string;
  manufacturer?: string;
  manufacturer_pn?: string;
  supplier?: string;
  supplier_pn?: string;
  unit_price?: number;
  lead_time_days?: number;
  procurement_type?: string;
  moq?: number;
  lifecycle_status?: string;
  is_critical?: boolean;
  is_alternative?: boolean;
  notes?: string;
  item_number?: number;
}

export const projectBomApi = {
  // 获取项目BOM列表
  list: async (projectId: string, params?: { phase?: string; bom_type?: string; status?: string }): Promise<ProjectBOM[]> => {
    const searchParams = new URLSearchParams();
    if (params?.phase) searchParams.set('phase', params.phase);
    if (params?.bom_type) searchParams.set('bom_type', params.bom_type);
    if (params?.status) searchParams.set('status', params.status);
    const query = searchParams.toString();
    const response = await apiClient.get<ApiResponse<ProjectBOM[]>>(
      `/projects/${projectId}/boms${query ? `?${query}` : ''}`
    );
    return response.data.data;
  },

  // 创建BOM
  create: async (projectId: string, data: CreateProjectBOMRequest): Promise<ProjectBOM> => {
    const response = await apiClient.post<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms`, data
    );
    return response.data.data;
  },

  // 获取BOM详情（含items）
  get: async (projectId: string, bomId: string): Promise<ProjectBOMDetail> => {
    const response = await apiClient.get<ApiResponse<ProjectBOMDetail>>(
      `/projects/${projectId}/boms/${bomId}`
    );
    return response.data.data;
  },

  // 更新BOM基本信息
  update: async (projectId: string, bomId: string, data: { name?: string; description?: string; version?: string }): Promise<ProjectBOM> => {
    const response = await apiClient.put<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms/${bomId}`, data
    );
    return response.data.data;
  },

  // 提交审批
  submit: async (projectId: string, bomId: string): Promise<ProjectBOM> => {
    const response = await apiClient.post<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms/${bomId}/submit`, {}
    );
    return response.data.data;
  },

  // 审批通过
  approve: async (projectId: string, bomId: string, comment?: string): Promise<ProjectBOM> => {
    const response = await apiClient.post<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms/${bomId}/approve`, { comment }
    );
    return response.data.data;
  },

  // 审批驳回
  reject: async (projectId: string, bomId: string, comment: string): Promise<ProjectBOM> => {
    const response = await apiClient.post<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms/${bomId}/reject`, { comment }
    );
    return response.data.data;
  },

  // 冻结
  freeze: async (projectId: string, bomId: string): Promise<ProjectBOM> => {
    const response = await apiClient.post<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms/${bomId}/freeze`, {}
    );
    return response.data.data;
  },

  // 添加物料行项
  addItem: async (projectId: string, bomId: string, data: BOMItemRequest): Promise<ProjectBOMItem> => {
    const response = await apiClient.post<ApiResponse<ProjectBOMItem>>(
      `/projects/${projectId}/boms/${bomId}/items`, data
    );
    return response.data.data;
  },

  // 批量添加物料行项
  batchAddItems: async (projectId: string, bomId: string, items: BOMItemRequest[]): Promise<{ created: number }> => {
    const response = await apiClient.post<ApiResponse<{ created: number }>>(
      `/projects/${projectId}/boms/${bomId}/items/batch`, { items }
    );
    return response.data.data;
  },

  // 更新物料行项
  updateItem: async (projectId: string, bomId: string, itemId: string, data: BOMItemRequest): Promise<ProjectBOMItem> => {
    const response = await apiClient.put<ApiResponse<ProjectBOMItem>>(
      `/projects/${projectId}/boms/${bomId}/items/${itemId}`, data
    );
    return response.data.data;
  },

  // 删除物料行项
  deleteItem: async (projectId: string, bomId: string, itemId: string): Promise<void> => {
    await apiClient.delete(`/projects/${projectId}/boms/${bomId}/items/${itemId}`);
  },

  // 拖拽排序
  reorderItems: async (projectId: string, bomId: string, itemIds: string[]): Promise<void> => {
    await apiClient.post(`/projects/${projectId}/boms/${bomId}/reorder`, { item_ids: itemIds });
  },

  // Excel导出（返回blob下载）
  exportExcel: async (projectId: string, bomId: string): Promise<void> => {
    const response = await apiClient.get(`/projects/${projectId}/boms/${bomId}/export`, { responseType: 'blob' });
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const a = document.createElement('a');
    a.href = url;
    a.download = `BOM_${bomId}.xlsx`;
    a.click();
    window.URL.revokeObjectURL(url);
  },

  // Excel导入（multipart/form-data）
  importExcel: async (projectId: string, bomId: string, file: File): Promise<any> => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await apiClient.post(`/projects/${projectId}/boms/${bomId}/import`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data.data;
  },

  // 下载导入模板
  downloadTemplate: async (): Promise<void> => {
    const response = await apiClient.get('/bom-template', { responseType: 'blob' });
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const a = document.createElement('a');
    a.href = url;
    a.download = 'BOM导入模板.xlsx';
    a.click();
    window.URL.revokeObjectURL(url);
  },

  // EBOM转MBOM
  convertToMBOM: async (projectId: string, bomId: string): Promise<any> => {
    const response = await apiClient.post(`/projects/${projectId}/boms/${bomId}/convert-to-mbom`, {});
    return response.data.data;
  },

  // BOM版本对比
  compareBOMs: async (bomId1: string, bomId2: string): Promise<any> => {
    const response = await apiClient.get(`/bom-compare?bom1=${bomId1}&bom2=${bomId2}`);
    return response.data.data;
  },
};
