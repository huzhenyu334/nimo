import React, { useMemo, useCallback } from 'react';
import {
  Typography,
  Empty,
  Upload,
  Space,
  Tooltip,
  message,
} from 'antd';
import {
  UploadOutlined,
  CloseCircleOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { taskFormApi } from '@/api/taskForms';
import type { CategoryAttrTemplate } from '@/api/projectBom';
import { COMMON_FIELDS } from './bomConstants';
import { useIsMobile } from '@/hooks/useIsMobile';
import EditableTable, { type EditableColumn } from '../EditableTable';
import SupplierSelect from './SupplierSelect';

const { Text } = Typography;

const PAGE_SIZE = 10;

// ========== Types ==========

export interface DynamicBOMTableProps {
  subCategory: string;
  items: Record<string, any>[];
  onChange: (items: Record<string, any>[]) => void;
  fieldOrder?: string[];
  templates: CategoryAttrTemplate[];
  readonly?: boolean;
  onItemSave?: (itemId: string, field: string, value: any) => void;
  showMaterialCode?: boolean;
  onDeleteRow?: (itemId: string) => void;
  bomType?: string;  // EBOM/PBOM/MBOM — 用于控制是否显示制造商/MPN列
  category?: string; // 大类: electronic/structural/... — 电子BOM才显示制造商/MPN
}

// ========== Helpers ==========

const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)}KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
};

const COMMON_FIELD_CONFIG: Record<string, { title: string; width: number; type: 'text' | 'number'; align?: 'left' | 'center' | 'right' }> = {
  name: { title: '名称', width: 120, type: 'text' },
  quantity: { title: '数量', width: 60, type: 'number', align: 'right' },
  unit: { title: '单位', width: 55, type: 'text' },
  supplier: { title: '供应商', width: 100, type: 'text' },
  unit_price: { title: '单价', width: 80, type: 'number', align: 'right' },
  extended_cost: { title: '小计', width: 80, type: 'number', align: 'right' },
  notes: { title: '备注', width: 120, type: 'text' },
};

// ========== Component ==========

const DynamicBOMTable: React.FC<DynamicBOMTableProps> = ({
  subCategory: _subCategory,
  items,
  onChange,
  fieldOrder,
  templates,
  readonly = false,
  onItemSave,
  showMaterialCode = false,
  onDeleteRow,
  bomType,
  category,
}) => {
  const isMobile = useIsMobile();

  // Flatten extended_attrs onto items for display
  const flatItems = useMemo(() => items.map(item => ({
    ...item,
    ...(item.extended_attrs || {}),
  })), [items]);

  // Determine which extended fields to show and their order
  const orderedTemplates = useMemo(() => {
    const showable = templates.filter(t => t.show_in_table);
    if (fieldOrder && fieldOrder.length > 0) {
      const ordered: CategoryAttrTemplate[] = [];
      const remaining = [...showable];
      for (const key of fieldOrder) {
        if (COMMON_FIELDS.includes(key)) continue;
        const idx = remaining.findIndex(t => t.field_key === key);
        if (idx >= 0) {
          ordered.push(remaining.splice(idx, 1)[0]);
        }
      }
      remaining.sort((a, b) => a.sort_order - b.sort_order);
      return [...ordered, ...remaining];
    }
    return showable.sort((a, b) => a.sort_order - b.sort_order);
  }, [templates, fieldOrder]);

  // Fields that live on the item directly (not in extended_attrs)
  const DIRECT_FIELDS = useMemo(() => [
    ...COMMON_FIELDS, 'item_number', 'supplier_id', 'manufacturer_id', 'mpn', 'manufacturer_name',
  ], []);

  // Handle cell save — route to extended_attrs or common fields
  const handleCellSave = useCallback((record: Record<string, any>, field: string, value: any, _index: number) => {
    const isExtendedField = !DIRECT_FIELDS.includes(field)
      && templates.some(t => t.field_key === field);

    if (onItemSave && record.id) {
      onItemSave(record.id, field, value);
    } else {
      if (isExtendedField) {
        const newItems = items.map(item => {
          if (item.id !== record.id) return item;
          return { ...item, extended_attrs: { ...(item.extended_attrs || {}), [field]: value } };
        });
        onChange(newItems);
      } else {
        const newItems = items.map(item => {
          if (item.id !== record.id) return item;
          const updated = { ...item, [field]: value };
          // Auto-compute extended_cost when quantity or unit_price changes
          if (field === 'quantity' || field === 'unit_price') {
            const qty = field === 'quantity' ? (Number(value) || 0) : (item.quantity || 0);
            const price = field === 'unit_price' ? (Number(value) || 0) : (item.unit_price || 0);
            updated.extended_cost = qty * price;
          }
          return updated;
        });
        onChange(newItems);
      }
    }
  }, [items, onChange, onItemSave, templates, DIRECT_FIELDS]);

  // Handle delete
  const handleDeleteRow = useCallback((record: Record<string, any>, _index: number) => {
    if (onDeleteRow && record.id) {
      onDeleteRow(record.id);
    } else {
      onChange(items.filter(item => item.id !== record.id));
    }
  }, [items, onChange, onDeleteRow]);

  // Build columns
  const editableColumns = useMemo((): EditableColumn[] => {
    const cols: EditableColumn[] = [];

    // Sequence number
    cols.push({ key: 'item_number', title: '序号', width: 55, align: 'center', type: 'number' });

    // Material code
    if (showMaterialCode) {
      cols.push({
        key: 'material_code', title: '物料编码', width: 150,
        render: (v: string) => v ? (
          <span style={{
            background: '#f5f5f5', border: '1px solid #d9d9d9', borderRadius: 4,
            padding: '2px 8px', fontFamily: 'monospace', fontSize: 13, whiteSpace: 'nowrap',
          }}>{v}</span>
        ) : <Text type="secondary" style={{ fontSize: 11 }}>-</Text>,
      });
    }

    // Common fields
    for (const fieldKey of COMMON_FIELDS) {
      const config = COMMON_FIELD_CONFIG[fieldKey];
      if (!config) continue;

      if (fieldKey === 'extended_cost') {
        cols.push({
          key: fieldKey, title: config.title, width: config.width, align: config.align,
          render: (v: any, record: Record<string, any>) => {
            const cost = v ?? ((record.quantity || 0) * (record.unit_price || 0));
            return cost > 0 ? Number(cost).toFixed(2) : <Text type="secondary" style={{ fontSize: 11 }}>-</Text>;
          },
        });
      } else if (fieldKey === 'supplier') {
        // Supplier column: Select with search (关联suppliers表)
        cols.push({
          key: 'supplier_id', title: config.title, width: 130, editable: false,
          render: (_: any, record: Record<string, any>) => (
            <SupplierSelect
              value={record.supplier_id}
              displayName={record.supplier}
              categoryNe="manufacturer"
              readonly={readonly}
              placeholder="选择供应商"
              onChange={(supplierId, supplierName) => {
                const newItems = items.map(item =>
                  item.id === record.id
                    ? { ...item, supplier_id: supplierId, supplier: supplierName }
                    : item,
                );
                onChange(newItems);
              }}
            />
          ),
        });
      } else {
        cols.push({
          key: fieldKey, title: config.title, width: config.width, align: config.align,
          type: config.type, ellipsis: fieldKey === 'notes',
          formatValue: fieldKey === 'unit_price'
            ? (v: any) => (v != null && v !== '') ? Number(v).toFixed(2) : null
            : undefined,
        });
      }
    }

    // EBOM electronic: 制造商 + MPN 列
    if (bomType === 'EBOM' && category === 'electronic') {
      cols.push({
        key: 'manufacturer_id', title: '制造商', width: 130, editable: false,
        render: (_: any, record: Record<string, any>) => (
          <SupplierSelect
            value={record.manufacturer_id}
            displayName={record.manufacturer_name}
            category="manufacturer"
            readonly={readonly}
            placeholder="选择制造商"
            onChange={(mfrId, mfrName) => {
              const newItems = items.map(item =>
                item.id === record.id
                  ? { ...item, manufacturer_id: mfrId, manufacturer_name: mfrName }
                  : item,
              );
              onChange(newItems);
            }}
          />
        ),
      });
      cols.push({
        key: 'mpn', title: 'MPN', width: 120, type: 'text', ellipsis: true,
      });
    }

    // Extended template fields
    for (const tmpl of orderedTemplates) {
      const title = tmpl.unit ? `${tmpl.field_name}(${tmpl.unit})` : tmpl.field_name;

      if (tmpl.field_type === 'file') {
        const fk = tmpl.field_key;
        cols.push({
          key: fk, title, width: 140,
          render: (_: any, record: Record<string, any>) => {
            const fileValue = record[fk];
            const hasFile = fileValue && typeof fileValue === 'object' && fileValue.file_id;

            if (readonly) {
              return hasFile ? (
                <Tooltip title={`${fileValue.file_name}${fileValue.file_size ? ' ' + formatFileSize(fileValue.file_size) : ''}`}>
                  <a href={`/uploads/${fileValue.file_id}/${fileValue.file_name}`} target="_blank" rel="noreferrer"
                    style={{ fontSize: 11, maxWidth: 100, display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {fileValue.file_name}
                  </a>
                  {fileValue.file_size > 0 && <Text type="secondary" style={{ fontSize: 10, marginLeft: 4 }}>{formatFileSize(fileValue.file_size)}</Text>}
                </Tooltip>
              ) : <Text type="secondary" style={{ fontSize: 11 }}>-</Text>;
            }

            return (
              <Space size={4} style={{ width: '100%' }}>
                {hasFile ? (
                  <Tooltip title={`${fileValue.file_name}${fileValue.file_size ? ' ' + formatFileSize(fileValue.file_size) : ''}`}>
                    <a href={`/uploads/${fileValue.file_id}/${fileValue.file_name}`} target="_blank" rel="noreferrer"
                      style={{ fontSize: 11, maxWidth: 70, display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', verticalAlign: 'middle' }}>
                      {fileValue.file_name}
                    </a>
                    {fileValue.file_size > 0 && <Text type="secondary" style={{ fontSize: 10 }}>{formatFileSize(fileValue.file_size)}</Text>}
                  </Tooltip>
                ) : <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
                <Upload
                  showUploadList={false}
                  customRequest={() => {}}
                  beforeUpload={(file) => {
                    taskFormApi.uploadFile(file).then((result) => {
                      const fileData = { file_id: result.id, file_name: result.filename, file_size: file.size };
                      const newItems = items.map(it => {
                        if (it.id !== record.id) return it;
                        return { ...it, extended_attrs: { ...(it.extended_attrs || {}), [fk]: fileData } };
                      });
                      onChange(newItems);
                      message.success('上传成功');
                    }).catch(() => { message.error('上传失败'); });
                    return Upload.LIST_IGNORE;
                  }}
                >
                  <UploadOutlined style={{ color: '#1677ff', cursor: 'pointer', fontSize: 12 }} />
                </Upload>
                {hasFile && (
                  <CloseCircleOutlined
                    style={{ color: '#ff4d4f', cursor: 'pointer', fontSize: 11 }}
                    onClick={() => {
                      const newItems = items.map(it => {
                        if (it.id !== record.id) return it;
                        const newAttrs = { ...(it.extended_attrs || {}) };
                        delete newAttrs[fk];
                        return { ...it, extended_attrs: newAttrs };
                      });
                      onChange(newItems);
                    }}
                  />
                )}
              </Space>
            );
          },
        });
      } else if (tmpl.field_type === 'thumbnail') {
        const fk = tmpl.field_key;
        cols.push({
          key: fk, title, width: 80, align: 'center',
          render: (_: any, record: Record<string, any>) => {
            const url = record[fk];
            if (url && typeof url === 'string') {
              return (
                <img src={url} width={64} height={64}
                  style={{ objectFit: 'contain', background: '#fff', borderRadius: 2 }} />
              );
            }
            return (
              <div style={{ width: 64, height: 64, background: '#f5f5f5', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: 4 }}>
                <EyeOutlined style={{ color: '#d9d9d9', fontSize: 16 }} />
              </div>
            );
          },
        });
      } else {
        const colType = tmpl.field_type === 'number' ? 'number'
          : tmpl.field_type === 'select' ? 'select'
          : tmpl.field_type === 'boolean' ? 'checkbox'
          : 'text';
        const selectOpts = tmpl.field_type === 'select' && tmpl.options?.values
          ? (tmpl.options.values as string[]).map((v: string) => ({ label: v, value: v }))
          : undefined;

        cols.push({
          key: tmpl.field_key, title,
          width: colType === 'checkbox' ? 60 : (colType === 'number' ? 80 : 100),
          align: (colType === 'checkbox' ? 'center' : colType === 'number' ? 'right' : undefined),
          type: colType as any,
          options: selectOpts,
          ellipsis: colType === 'text',
        });
      }
    }

    return cols;
  }, [orderedTemplates, showMaterialCode, items, onChange, readonly, bomType, category]);

  // Mobile card view (BOM-specific)
  if (isMobile) {
    if (flatItems.length === 0) {
      return <Empty description="暂无物料" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ padding: '16px 0' }} />;
    }
    return (
      <div className="bom-mobile-card-list">
        {flatItems.map((item, idx) => {
          const cost = item.extended_cost ?? ((item.quantity || 0) * (item.unit_price || 0));
          return (
            <div key={item.id || idx} className="bom-item-card">
              <div className="bom-item-card-header">
                <span className="bom-item-card-name">{item.name || `#${item.item_number || idx + 1}`}</span>
                {item.material_code && <span className="bom-item-card-code">{item.material_code}</span>}
              </div>
              <div className="bom-item-card-meta">
                <span className="bom-item-card-meta-item">
                  <span className="bom-item-card-meta-label">数量</span>
                  <span className="bom-item-card-meta-value">{item.quantity || 0} {item.unit || ''}</span>
                </span>
                {item.unit_price > 0 && (
                  <span className="bom-item-card-meta-item">
                    <span className="bom-item-card-meta-label">单价</span>
                    <span className="bom-item-card-meta-value">{'\u00a5'}{Number(item.unit_price).toFixed(2)}</span>
                  </span>
                )}
                {cost > 0 && (
                  <span className="bom-item-card-meta-item">
                    <span className="bom-item-card-meta-label">小计</span>
                    <span className="bom-item-card-cost">{'\u00a5'}{Number(cost).toFixed(2)}</span>
                  </span>
                )}
                {item.supplier && (
                  <span className="bom-item-card-meta-item">
                    <span className="bom-item-card-meta-label">供应商</span>
                    <span className="bom-item-card-meta-value">{item.supplier}</span>
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    );
  }

  return (
    <EditableTable
      columns={editableColumns}
      items={flatItems}
      onCellSave={handleCellSave}
      onDeleteRow={handleDeleteRow}
      readonly={readonly}
      rowKey={(r, idx) => r.id || String(idx)}
      pageSize={PAGE_SIZE}
      emptyText={<Empty description="暂无物料" image={Empty.PRESENTED_IMAGE_SIMPLE} />}
    />
  );
};

export default DynamicBOMTable;
