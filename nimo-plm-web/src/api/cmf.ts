import apiClient from './client';

export interface CMFDesign {
  id: string;
  project_id: string;
  task_id: string;
  bom_item_id: string;
  scheme_name: string;
  color: string;
  color_code: string;
  gloss_level: string;
  surface_treatment: string;
  texture_pattern: string;
  coating_type: string;
  render_image_file_id?: string;
  render_image_file_name?: string;
  notes: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
  bom_item?: {
    id: string;
    name: string;
    item_number: number;
    material_type?: string;
    thumbnail_url?: string;
  };
  drawings?: CMFDrawing[];
}

export interface CMFDrawing {
  id: string;
  cmf_design_id: string;
  drawing_type: string;
  file_id: string;
  file_name: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface AppearancePart {
  id: string;
  bom_id: string;
  item_number: number;
  name: string;
  material_type?: string;
  thumbnail_url?: string;
  process_type?: string;
}

export const cmfApi = {
  // 获取外观件列表
  getAppearanceParts: async (projectId: string, taskId: string): Promise<AppearancePart[]> => {
    const res = await apiClient.get(`/projects/${projectId}/tasks/${taskId}/cmf/appearance-parts`);
    return res.data.data?.items || [];
  },

  // 列出任务的所有CMF方案
  listDesigns: async (projectId: string, taskId: string): Promise<CMFDesign[]> => {
    const res = await apiClient.get(`/projects/${projectId}/tasks/${taskId}/cmf/designs`);
    return res.data.data?.items || [];
  },

  // 列出项目所有CMF方案
  listDesignsByProject: async (projectId: string): Promise<CMFDesign[]> => {
    const res = await apiClient.get(`/projects/${projectId}/cmf/designs`);
    return res.data.data?.items || [];
  },

  // 创建CMF方案
  createDesign: async (projectId: string, taskId: string, data: Partial<CMFDesign>): Promise<CMFDesign> => {
    const res = await apiClient.post(`/projects/${projectId}/tasks/${taskId}/cmf/designs`, data);
    return res.data.data;
  },

  // 更新CMF方案
  updateDesign: async (projectId: string, taskId: string, designId: string, data: Partial<CMFDesign>): Promise<CMFDesign> => {
    const res = await apiClient.put(`/projects/${projectId}/tasks/${taskId}/cmf/designs/${designId}`, data);
    return res.data.data;
  },

  // 删除CMF方案
  deleteDesign: async (projectId: string, taskId: string, designId: string): Promise<void> => {
    await apiClient.delete(`/projects/${projectId}/tasks/${taskId}/cmf/designs/${designId}`);
  },

  // 添加图纸
  addDrawing: async (designId: string, data: { drawing_type: string; file_id: string; file_name: string; notes?: string }): Promise<CMFDrawing> => {
    const res = await apiClient.post(`/cmf-designs/${designId}/drawings`, data);
    return res.data.data;
  },

  // 删除图纸
  removeDrawing: async (designId: string, drawingId: string): Promise<void> => {
    await apiClient.delete(`/cmf-designs/${designId}/drawings/${drawingId}`);
  },
};
