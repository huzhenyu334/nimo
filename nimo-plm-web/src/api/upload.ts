import apiClient from './client';
import { ApiResponse } from '@/types';

export interface UploadResult {
  id: string;
  url: string;
  filename: string;
  size: number;
  thumbnail_url?: string;
}

export const uploadApi = {
  /**
   * Upload a file to the server.
   * POST /upload with multipart/form-data
   */
  uploadFile: async (file: File): Promise<UploadResult> => {
    const formData = new FormData();
    formData.append('files', file);
    const response = await apiClient.post<ApiResponse<UploadResult[]>>('/upload', formData);
    return response.data.data[0];
  },
};
