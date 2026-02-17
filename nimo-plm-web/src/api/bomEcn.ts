import apiClient from './client';
import { ApiResponse } from '@/types';
import { BOMECN } from './projectBom';

export const bomEcnApi = {
  // 获取ECN列表
  list: async (params?: { bom_id?: string; status?: string }): Promise<{ items: BOMECN[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.bom_id) searchParams.set('bom_id', params.bom_id);
    if (params?.status) searchParams.set('status', params.status);
    const query = searchParams.toString();
    const response = await apiClient.get<ApiResponse<{ items: BOMECN[]; total: number }>>(
      `/bom-ecn${query ? `?${query}` : ''}`
    );
    return response.data.data;
  },

  // 获取ECN详情
  get: async (id: string): Promise<BOMECN> => {
    const response = await apiClient.get<ApiResponse<BOMECN>>(`/bom-ecn/${id}`);
    return response.data.data;
  },

  // 审批通过
  approve: async (id: string): Promise<BOMECN> => {
    const response = await apiClient.post<ApiResponse<BOMECN>>(`/bom-ecn/${id}/approve`);
    return response.data.data;
  },

  // 审批拒绝
  reject: async (id: string, data: { note: string }): Promise<BOMECN> => {
    const response = await apiClient.post<ApiResponse<BOMECN>>(`/bom-ecn/${id}/reject`, data);
    return response.data.data;
  },
};
