import React, { useState, useCallback, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Button, Spin, Result, message } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import EBOMControl from '@/components/BOM/EBOMControl';
import { projectBomApi } from '@/api/projectBom';
import type { BOMItemRequest } from '@/api/projectBom';
import { EBOM_CATEGORIES, type BOMControlConfig } from '@/components/BOM/bomConstants';

// ========== Default config for embedded EBOM ==========

const DEFAULT_EBOM_CONFIG: BOMControlConfig = {
  bom_type: 'EBOM',
  visible_categories: EBOM_CATEGORIES,
  category_config: {},
};

// ========== Helper: convert item to BOMItemRequest ==========

const toBOMItemRequest = (item: Record<string, any>): BOMItemRequest => ({
  material_id: item.material_id,
  parent_item_id: item.parent_item_id,
  level: item.level,
  category: item.category,
  sub_category: item.sub_category,
  name: item.name || '',
  specification: item.specification,
  quantity: item.quantity || 1,
  unit: item.unit || 'pcs',
  reference: item.reference,
  manufacturer: item.manufacturer,
  manufacturer_pn: item.manufacturer_pn,
  supplier: item.supplier,
  supplier_pn: item.supplier_pn,
  unit_price: item.unit_price,
  lead_time_days: item.lead_time_days,
  is_critical: item.is_critical || false,
  is_alternative: item.is_alternative || false,
  notes: item.notes,
  drawing_no: item.drawing_no,
  item_number: item.item_number,
  extended_attrs: item.extended_attrs,
});

// ========== Component ==========

const EBOMEmbed: React.FC = () => {
  const [searchParams] = useSearchParams();
  const projectId = searchParams.get('project_id') || '';
  const bomId = searchParams.get('bom_id') || '';
  const mode = searchParams.get('mode') || 'edit';
  const token = searchParams.get('token');

  const isReadonly = mode === 'view';

  const [items, setItems] = useState<Record<string, any>[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const originalItemIdsRef = useRef<Set<string>>(new Set());

  // Inject token from URL param into localStorage (runs once, synchronously via ref)
  const tokenInitialized = useRef(false);
  if (token && !tokenInitialized.current) {
    localStorage.setItem('access_token', token);
    tokenInitialized.current = true;
  }

  // Fetch BOM detail
  const { data: bomDetail, isLoading, error } = useQuery({
    queryKey: ['embed-bom', projectId, bomId],
    queryFn: () => projectBomApi.get(projectId, bomId),
    enabled: !!projectId && !!bomId,
  });

  // Initialize items from fetched data
  useEffect(() => {
    if (bomDetail?.items) {
      setItems(bomDetail.items);
      originalItemIdsRef.current = new Set(bomDetail.items.map((item) => item.id));
    }
  }, [bomDetail]);

  // Handle items change from EBOMControl
  const handleChange = useCallback((newItems: Record<string, any>[]) => {
    setItems(newItems);
  }, []);

  // Save BOM data and notify parent window
  const handleSubmit = useCallback(async () => {
    if (!projectId || !bomId) return;

    setSubmitting(true);
    try {
      const originalIds = originalItemIdsRef.current;
      const currentIds = new Set(items.map((item) => item.id));

      // Diff: deleted / new / updated
      const deletedIds = [...originalIds].filter((id) => !currentIds.has(id));
      const newItems = items.filter((item) => String(item.id).startsWith('new-'));
      const updatedItems = items.filter(
        (item) => !String(item.id).startsWith('new-') && originalIds.has(item.id),
      );

      const promises: Promise<any>[] = [];

      // Delete removed items
      for (const id of deletedIds) {
        promises.push(projectBomApi.deleteItem(projectId, bomId, id));
      }

      // Batch add new items
      if (newItems.length > 0) {
        promises.push(
          projectBomApi.batchAddItems(
            projectId,
            bomId,
            newItems.map(toBOMItemRequest),
          ),
        );
      }

      // Update existing items
      for (const item of updatedItems) {
        promises.push(
          projectBomApi.updateItem(projectId, bomId, item.id, toBOMItemRequest(item)),
        );
      }

      await Promise.all(promises);

      message.success('BOM数据保存成功');

      // Notify parent window
      window.parent.postMessage(
        {
          type: 'form_submitted',
          formType: 'ebom',
          data: { project_id: projectId, bom_id: bomId, status: 'success' },
        },
        '*',
      );
    } catch (err) {
      console.error('Failed to save BOM data:', err);
      message.error('保存失败，请重试');

      window.parent.postMessage(
        {
          type: 'form_submitted',
          formType: 'ebom',
          data: { project_id: projectId, bom_id: bomId, status: 'error' },
        },
        '*',
      );
    } finally {
      setSubmitting(false);
    }
  }, [projectId, bomId, items]);

  // ========== Render ==========

  // Validate required params
  if (!projectId || !bomId) {
    return (
      <Result
        status="warning"
        title="参数缺失"
        subTitle="请提供 project_id 和 bom_id 参数"
      />
    );
  }

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" tip="加载BOM数据..." />
      </div>
    );
  }

  if (error) {
    return (
      <Result
        status="error"
        title="加载失败"
        subTitle="无法加载BOM数据，请检查参数或权限"
      />
    );
  }

  return (
    <div style={{ padding: 16, paddingBottom: isReadonly ? 16 : 72 }}>
      <EBOMControl
        config={DEFAULT_EBOM_CONFIG}
        value={items}
        onChange={handleChange}
        readonly={isReadonly}
      />

      {/* Fixed submit button bar (edit mode only) */}
      {!isReadonly && (
        <div
          style={{
            position: 'fixed',
            bottom: 0,
            left: 0,
            right: 0,
            padding: '12px 16px',
            background: '#fff',
            borderTop: '1px solid #f0f0f0',
            boxShadow: '0 -2px 8px rgba(0, 0, 0, 0.06)',
            display: 'flex',
            justifyContent: 'flex-end',
            zIndex: 100,
          }}
        >
          <Button
            type="primary"
            icon={<SaveOutlined />}
            loading={submitting}
            onClick={handleSubmit}
            size="large"
          >
            提交
          </Button>
        </div>
      )}
    </div>
  );
};

export default EBOMEmbed;
