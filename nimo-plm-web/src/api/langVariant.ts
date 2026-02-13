import apiClient from './client';
import type { LangVariant } from './projectBom';

export interface MultilangPartWithVariants {
  bom_item: {
    id: string;
    name: string;
    item_number: number;
    category?: string;
    is_multilang: boolean;
  };
  lang_variants: LangVariant[];
  bom_id: string;
  bom_name: string;
}

export interface CreateLangVariantInput {
  language_code: string;
  language_name: string;
  design_file_id?: string;
  design_file_name?: string;
  design_file_url?: string;
  notes?: string;
}

export const langVariantApi = {
  getMultilangParts: async (projectId: string): Promise<MultilangPartWithVariants[]> => {
    const res = await apiClient.get(`/projects/${projectId}/multilang-parts`);
    return res.data.data || [];
  },

  listVariants: async (projectId: string, itemId: string): Promise<LangVariant[]> => {
    const res = await apiClient.get(`/projects/${projectId}/bom-items/${itemId}/lang-variants`);
    return res.data.data || [];
  },

  createVariant: async (projectId: string, itemId: string, data: CreateLangVariantInput): Promise<LangVariant> => {
    const res = await apiClient.post(`/projects/${projectId}/bom-items/${itemId}/lang-variants`, data);
    return res.data.data;
  },

  updateVariant: async (projectId: string, variantId: string, data: Partial<CreateLangVariantInput>): Promise<LangVariant> => {
    const res = await apiClient.put(`/projects/${projectId}/lang-variants/${variantId}`, data);
    return res.data.data;
  },

  deleteVariant: async (projectId: string, variantId: string): Promise<void> => {
    await apiClient.delete(`/projects/${projectId}/lang-variants/${variantId}`);
  },
};
