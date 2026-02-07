import apiClient from './client';
import { ApiResponse } from '@/types';

// 项目BOM相关类型
export interface ProjectBOM {
  id: string;
  project_id: string;
  name: string;
  bom_type: 'EBOM' | 'SBOM' | 'MBOM';
  phase: string;
  version: string;
  status: 'draft' | 'submitted' | 'approved' | 'rejected' | 'released' | 'frozen';
  total_items: number;
  total_cost: number;
  created_by: string;
  creator_name?: string;
  released_at?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectBOMItem {
  id: string;
  bom_id: string;
  sequence: number;
  material_id: string;
  material_name?: string;
  material_code?: string;
  specification?: string;
  quantity: number;
  unit: string;
  supplier?: string;
  unit_price?: number;
  position?: string;
  is_critical: boolean;
  notes?: string;
}

export interface ProjectBOMDetail extends ProjectBOM {
  items: ProjectBOMItem[];
}

export interface CreateProjectBOMRequest {
  name: string;
  bom_type: 'EBOM' | 'SBOM' | 'MBOM';
  phase: string;
  version?: string;
  description?: string;
}

export const projectBomApi = {
  // 获取项目BOM列表
  list: async (projectId: string, params?: { phase?: string; bom_type?: string }): Promise<{ items: ProjectBOM[] }> => {
    const searchParams = new URLSearchParams();
    if (params?.phase) searchParams.set('phase', params.phase);
    if (params?.bom_type) searchParams.set('bom_type', params.bom_type);
    const query = searchParams.toString();
    const response = await apiClient.get<ApiResponse<{ items: ProjectBOM[] }>>(
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

  // 更新BOM
  update: async (projectId: string, bomId: string, data: Partial<ProjectBOM>): Promise<ProjectBOM> => {
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
  approve: async (projectId: string, bomId: string): Promise<ProjectBOM> => {
    const response = await apiClient.post<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms/${bomId}/approve`, {}
    );
    return response.data.data;
  },

  // 审批驳回
  reject: async (projectId: string, bomId: string, reason?: string): Promise<ProjectBOM> => {
    const response = await apiClient.post<ApiResponse<ProjectBOM>>(
      `/projects/${projectId}/boms/${bomId}/reject`, { reason }
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
};
