import React, { useState, useEffect } from 'react';
import {
  Table,
  Input,
  InputNumber,
  Select,
  Checkbox,
  Button,
  Typography,
  Popconfirm,
  Empty,
  Upload,
  Space,
  message,
  AutoComplete,
  Tooltip,
} from 'antd';
import { DeleteOutlined, PlusOutlined, UploadOutlined, CloseCircleOutlined, EyeOutlined, LoadingOutlined, EyeTwoTone, SearchOutlined } from '@ant-design/icons';
import { taskFormApi } from '@/api/taskForms';
import type { ColumnsType } from 'antd/es/table';
import STLViewer from './STLViewer';

const { Text } = Typography;

// ========== Option Constants ==========

export const CATEGORY_OPTIONS = [
  '电子元器件', '结构件', '光学器件', '电池', '线缆/FPC', '包装材料', '标签/外观件', '其他',
];

export const PROCUREMENT_OPTIONS = [
  { label: 'Buy（外购）', value: 'buy' },
  { label: 'Make（自制）', value: 'make' },
  { label: 'Phantom（虚拟件）', value: 'phantom' },
];

export const PROCESS_TYPE_OPTIONS = [
  { label: '注塑', value: '注塑' }, { label: 'CNC', value: 'CNC' },
  { label: '冲压', value: '冲压' }, { label: '模切', value: '模切' },
  { label: '3D打印', value: '3D打印' }, { label: '激光切割', value: '激光切割' },
  { label: 'SMT', value: 'SMT' }, { label: '手工', value: '手工' },
];

export const ASSEMBLY_METHOD_OPTIONS = [
  { label: '卡扣', value: '卡扣' }, { label: '螺丝', value: '螺丝' },
  { label: '胶合', value: '胶合' }, { label: '超声波焊接', value: '超声波焊接' },
  { label: '热熔', value: '热熔' }, { label: '激光焊接', value: '激光焊接' },
];

export const TOLERANCE_GRADE_OPTIONS = [
  { label: '普通', value: '普通' }, { label: '精密', value: '精密' }, { label: '超精密', value: '超精密' },
];

export const MATERIAL_TYPE_PRESETS = [
  '钛合金', '铝合金6061', '铝合金7075', '不锈钢304', '不锈钢316L',
  'PC', 'ABS', 'ABS+PC', 'PA66', 'PA66+GF30', 'PMMA', 'POM', 'TPU',
  '硅胶', 'PEEK', '碳纤维', '玻璃', '蓝宝石', '镁合金', '锌合金', '铜合金', 'TR90', 'Ultem',
];

export const TOLERANCE_PRESETS = [
  { label: '普通 ±0.05mm', value: '0.05' },
  { label: '精密 ±0.02mm', value: '0.02' },
  { label: '超精密 ±0.005mm', value: '0.005' },
];

// 格式化公差显示
const formatTolerance = (v: any): string => {
  if (v == null || v === '') return '';
  const num = parseFloat(String(v));
  if (!isNaN(num)) return `±${num}mm`;
  // 兼容旧数据
  const map: Record<string, string> = { '普通': '±0.05mm', '精密': '±0.03mm', '超精密': '±0.005mm' };
  return map[v] || String(v);
};

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)}KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
};

// ========== Types ==========

export type BOMItemRecord = Record<string, any>;

export interface BOMEditableTableProps {
  bomType: 'EBOM' | 'SBOM' | 'MBOM';
  items: BOMItemRecord[];
  onChange: (items: BOMItemRecord[]) => void;
  showAddDelete?: boolean;
  readonly?: boolean;
  showMaterialCode?: boolean;
  onItemSave?: (itemId: string, field: string, value: any) => void;
  onMaterialSearch?: (itemId: string) => void;
  renderDrawingColumn?: (record: BOMItemRecord, type: '2D' | '3D') => React.ReactNode;
  actionColumn?: (record: BOMItemRecord) => React.ReactNode;
  scrollX?: number;
  scrollY?: number;
  noPagination?: boolean;
  rowClassName?: (record: BOMItemRecord) => string;
}

// ========== Component ==========

const PAGE_SIZE = 10;

// 缩略图组件：异步加载，SVG没生成好时自动轮询重试，hover放大预览
const ThumbnailCell: React.FC<{ url: string }> = ({ url }) => {
  const [loaded, setLoaded] = useState(false);
  const [retries, setRetries] = useState(0);
  const [imgSrc, setImgSrc] = useState(url);
  const [hover, setHover] = useState(false);
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 });
  const maxRetries = 6;

  useEffect(() => {
    if (loaded || retries >= maxRetries) return;
    const img = new window.Image();
    let timer: ReturnType<typeof setTimeout>;
    img.onload = () => { setLoaded(true); setImgSrc(url + '?t=' + Date.now()); };
    img.onerror = () => { timer = setTimeout(() => setRetries(r => r + 1), 2000); };
    img.src = url + '?t=' + Date.now();
    return () => clearTimeout(timer);
  }, [url, retries, loaded]);

  if (!loaded && retries < maxRetries) {
    return (
      <div style={{ width: 60, height: 45, background: '#f5f5f5', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: 4 }}>
        <LoadingOutlined style={{ color: '#1677ff', fontSize: 16 }} />
      </div>
    );
  }

  if (!loaded) {
    return (
      <div style={{ width: 60, height: 45, background: '#f5f5f5', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: 4 }}>
        <EyeOutlined style={{ color: '#d9d9d9', fontSize: 16 }} />
      </div>
    );
  }

  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      onMouseMove={(e) => setMousePos({ x: e.clientX, y: e.clientY })}
      style={{ position: 'relative', display: 'inline-block' }}
    >
      <img
        src={imgSrc}
        width={60}
        height={45}
        style={{ objectFit: 'contain', cursor: 'pointer', background: '#fff', borderRadius: 2 }}
      />
      {hover && (
        <div style={{
          position: 'fixed',
          left: mousePos.x + 16,
          top: mousePos.y - 120,
          zIndex: 9999,
          pointerEvents: 'none',
          background: '#ffffff',
          borderRadius: 8,
          boxShadow: '0 4px 20px rgba(0,0,0,0.15)',
          padding: 8,
        }}>
          <img
            src={imgSrc}
            width={300}
            height={225}
            style={{ objectFit: 'contain', display: 'block' }}
          />
        </div>
      )}
    </div>
  );
};

const BOMEditableTable: React.FC<BOMEditableTableProps> = ({
  bomType,
  items,
  onChange,
  showAddDelete = true,
  readonly = false,
  showMaterialCode = false,
  onItemSave,
  onMaterialSearch,
  renderDrawingColumn,
  actionColumn,
  scrollX: scrollXProp,
  scrollY,
  noPagination = false,
  rowClassName,
}) => {
  const [editingCell, setEditingCell] = useState<{ rowIdx: number; field: string } | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [preview3D, setPreview3D] = useState<{ fileId: string; fileName: string } | null>(null);

  // Convert page-relative index to global array index
  const toGlobalIdx = (pageIdx: number) =>
    items.length > PAGE_SIZE ? (currentPage - 1) * PAGE_SIZE + pageIdx : pageIdx;

  const handleCellSave = (idx: number, field: string, value: any) => {
    const gi = toGlobalIdx(idx);
    // Skip save if value unchanged
    if (items[gi] && items[gi][field] === value) {
      setEditingCell(null);
      return;
    }
    if (onItemSave && items[gi]?.id) {
      onItemSave(items[gi].id, field, value);
    } else {
      const newItems = items.map((item, i) => i === gi ? { ...item, [field]: value } : item);
      onChange(newItems);
    }
    setEditingCell(null);
  };

  const handleDeleteRow = (idx: number) => {
    const gi = toGlobalIdx(idx);
    onChange(items.filter((_, i) => i !== gi));
  };

  const handleAddRow = () => {
    const nextNum = items.length > 0 ? Math.max(...items.map(i => i.item_number || 0)) + 1 : 1;
    const newItem: BOMItemRecord = {
      item_number: nextNum,
      name: '新零件',
      quantity: 1,
      unit: 'pcs',
      weight_grams: 0,
      target_price: 0,
      tooling_estimate: 0,
      tolerance_grade: '0.05',
    };
    const newItems = [...items, newItem];
    onChange(newItems);
    // Auto-navigate to last page
    if (newItems.length > PAGE_SIZE) {
      setCurrentPage(Math.ceil(newItems.length / PAGE_SIZE));
    }
  };

  // Click-to-edit cell renderer
  const renderCell = (
    value: any, idx: number, field: string,
    type: 'text' | 'number' | 'select' | 'checkbox' = 'text',
    options?: { label: string; value: string }[],
  ) => {
    const gi = toGlobalIdx(idx);
    // Checkbox: always show inline
    if (type === 'checkbox') {
      return (
        <Checkbox
          checked={!!value}
          disabled={readonly}
          onChange={(e) => handleCellSave(idx, field, e.target.checked)}
        />
      );
    }

    // Format display value
    let displayValue = value;
    if (type === 'select' && options && value) {
      displayValue = options.find(o => o.value === value)?.label?.split('（')[0] || value;
    }
    if (type === 'number' && (value != null && value !== '') && field === 'target_price') displayValue = `¥${Number(value).toFixed(2)}`;
    else if (type === 'number' && field === 'tooling_estimate') displayValue = `¥${Number(value || 0).toFixed(2)}`;
    else if (type === 'number' && (value != null && value !== '') && field === 'unit_price') displayValue = Number(value).toFixed(2);

    // Readonly mode: just show value
    if (readonly) {
      return (
        <div style={{ minHeight: 22, padding: '0 2px' }}>
          {displayValue ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
        </div>
      );
    }

    const isEditing = editingCell?.rowIdx === gi && editingCell?.field === field;

    if (isEditing) {
      if (type === 'number') {
        return (
          <InputNumber
            size="small"
            autoFocus
            defaultValue={typeof value === 'string' ? parseFloat(value) || 0 : value}
            style={{ width: '100%' }}
            onBlur={(e) => handleCellSave(idx, field, parseFloat((e.target as HTMLInputElement).value) || 0)}
            onPressEnter={(e) => handleCellSave(idx, field, parseFloat((e.target as HTMLInputElement).value) || 0)}
          />
        );
      }
      if (type === 'select' && options) {
        return (
          <Select
            size="small"
            autoFocus
            defaultValue={value}
            defaultOpen
            style={{ width: '100%' }}
            options={options}
            onChange={(v) => handleCellSave(idx, field, v)}
            onBlur={() => setEditingCell(null)}
          />
        );
      }
      return (
        <Input
          size="small"
          autoFocus
          defaultValue={value}
          onBlur={(e) => handleCellSave(idx, field, e.target.value)}
          onPressEnter={(e) => handleCellSave(idx, field, (e.target as HTMLInputElement).value)}
        />
      );
    }

    // Display mode: click to edit
    return (
      <div
        style={{ cursor: 'pointer', minHeight: 22, padding: '0 2px', borderRadius: 2 }}
        className="editable-cell"
        onClick={() => setEditingCell({ rowIdx: gi, field })}
      >
        {displayValue ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
      </div>
    );
  };

  // AutoComplete renderer for material_type
  const renderMaterialTypeCell = (value: any, idx: number) => {
    if (readonly) {
      return (
        <div style={{ minHeight: 22, padding: '0 2px' }}>
          {value ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
        </div>
      );
    }
    const gi = toGlobalIdx(idx);
    const isEditing = editingCell?.rowIdx === gi && editingCell?.field === 'material_type';
    if (isEditing) {
      return (
        <AutoComplete
          size="small"
          autoFocus
          defaultValue={value}
          defaultOpen
          style={{ width: '100%' }}
          options={MATERIAL_TYPE_PRESETS.map(m => ({ value: m }))}
          filterOption={(input, option) =>
            (option?.value as string)?.toLowerCase().includes(input.toLowerCase())
          }
          onSelect={(v) => handleCellSave(idx, 'material_type', v)}
          onBlur={(e) => handleCellSave(idx, 'material_type', (e.target as HTMLInputElement).value)}
        />
      );
    }
    return (
      <div
        style={{ cursor: 'pointer', minHeight: 22, padding: '0 2px', borderRadius: 2 }}
        className="editable-cell"
        onClick={() => setEditingCell({ rowIdx: gi, field: 'material_type' })}
      >
        {value ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
      </div>
    );
  };

  // AutoComplete renderer for tolerance_grade (±mm)
  const renderToleranceCell = (value: any, idx: number) => {
    const display = formatTolerance(value);
    if (readonly) {
      return (
        <div style={{ minHeight: 22, padding: '0 2px' }}>
          {display || <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
        </div>
      );
    }
    const gi = toGlobalIdx(idx);
    const isEditing = editingCell?.rowIdx === gi && editingCell?.field === 'tolerance_grade';
    if (isEditing) {
      return (
        <AutoComplete
          size="small"
          autoFocus
          defaultValue={value != null ? String(value) : ''}
          defaultOpen
          style={{ width: '100%' }}
          options={TOLERANCE_PRESETS}
          onSelect={(v) => handleCellSave(idx, 'tolerance_grade', v)}
          onBlur={(e) => {
            const raw = (e.target as HTMLInputElement).value;
            handleCellSave(idx, 'tolerance_grade', raw);
          }}
        />
      );
    }
    return (
      <div
        style={{ cursor: 'pointer', minHeight: 22, padding: '0 2px', borderRadius: 2 }}
        className="editable-cell"
        onClick={() => setEditingCell({ rowIdx: gi, field: 'tolerance_grade' })}
      >
        {display || <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
      </div>
    );
  };

  // Check if file is a 3D model (STP/STEP)
  const is3DFile = (name: string) => /\.(stp|step)$/i.test(name);

  // Render file name link: 3D files open preview, others download
  const renderFileLink = (fileId: string | undefined, fileName: string, style: React.CSSProperties) => {
    if (fileId && is3DFile(fileName)) {
      return (
        <a
          style={style}
          onClick={(e) => { e.preventDefault(); e.stopPropagation(); setPreview3D({ fileId, fileName }); }}
        >
          {fileName}
          <EyeTwoTone style={{ fontSize: 10, marginLeft: 3 }} />
        </a>
      );
    }
    return (
      <a href={fileId ? `/uploads/${fileId}/${fileName}` : '#'} target="_blank" rel="noreferrer" style={style}>
        {fileName}
      </a>
    );
  };

  // Drawing file upload cell with filename + size display
  const renderDrawingUploadCell = (record: BOMItemRecord, idx: number, fileIdField: string, fileNameField: string, fileSizeField: string) => {
    const gi = toGlobalIdx(idx);
    const fileName = record?.[fileNameField];
    const fileSize = record?.[fileSizeField];
    const fileId = record?.[fileIdField];

    if (readonly) {
      return fileName ? (
        <Tooltip title={`${fileName}${fileSize ? ' ' + formatFileSize(fileSize) : ''}`}>
          {renderFileLink(fileId, fileName, { fontSize: 11, maxWidth: 100, display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' })}
          {fileSize > 0 && <Text type="secondary" style={{ fontSize: 10, marginLeft: 4 }}>{formatFileSize(fileSize)}</Text>}
        </Tooltip>
      ) : <Text type="secondary" style={{ fontSize: 11 }}>-</Text>;
    }

    return (
      <Space size={4} style={{ width: '100%' }}>
        {fileName ? (
          <Tooltip title={`${fileName}${fileSize ? ' ' + formatFileSize(fileSize) : ''}`}>
            {renderFileLink(fileId, fileName, { fontSize: 11, maxWidth: 70, display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', verticalAlign: 'middle' })}
            {fileSize > 0 && <Text type="secondary" style={{ fontSize: 10 }}>{formatFileSize(fileSize)}</Text>}
          </Tooltip>
        ) : <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
        <Upload
          showUploadList={false}
          customRequest={() => {}}
          beforeUpload={(file) => {
            console.log('[BOM Upload] started:', file.name, 'gi:', gi, 'field:', fileNameField);
            taskFormApi.uploadFile(file).then((result) => {
              console.log('[BOM Upload] result:', JSON.stringify(result));
              const updateFields: Record<string, any> = {
                [fileIdField]: result.id,
                [fileNameField]: result.filename,
                [fileSizeField]: file.size,
              };
              // 3D模型上传时，如果返回了缩略图URL，也保存到BOM项
              if (result.thumbnail_url) {
                updateFields.thumbnail_url = result.thumbnail_url;
              }
              const newItems = items.map((it, i) => i === gi ? {
                ...it,
                ...updateFields,
              } : it);
              console.log('[BOM Upload] updated item:', JSON.stringify(newItems[gi]));
              onChange(newItems);
              message.success('上传成功');
            }).catch((err) => {
              console.error('[BOM Upload] failed:', err);
              message.error('上传失败');
            });
            return Upload.LIST_IGNORE;
          }}
        >
          <UploadOutlined style={{ color: '#1677ff', cursor: 'pointer', fontSize: 12 }} />
        </Upload>
        {fileName && (
          <CloseCircleOutlined
            style={{ color: '#ff4d4f', cursor: 'pointer', fontSize: 11 }}
            onClick={() => {
              const newItems = items.map((it, i) => i === gi ? {
                ...it, [fileIdField]: undefined, [fileNameField]: undefined, [fileSizeField]: undefined,
              } : it);
              onChange(newItems);
            }}
          />
        )}
      </Space>
    );
  };

  // Build columns based on bomType
  const commonCols: ColumnsType<BOMItemRecord> = [
    { title: '序号', dataIndex: 'item_number', width: 55, align: 'center',
      render: (v, _, idx) => renderCell(v, idx, 'item_number', 'number') },
  ];

  const ebomCols: ColumnsType<BOMItemRecord> = [
    { title: '位号', dataIndex: 'reference', width: 80,
      render: (v, _, idx) => renderCell(v, idx, 'reference') },
    { title: '名称', dataIndex: 'name', width: 120,
      render: (v, _, idx) => renderCell(v, idx, 'name') },
    { title: '规格', dataIndex: 'specification', width: 140, ellipsis: true,
      render: (v, _, idx) => renderCell(v, idx, 'specification') },
    { title: '数量', dataIndex: 'quantity', width: 60, align: 'right',
      render: (v, _, idx) => renderCell(v, idx, 'quantity', 'number') },
    { title: '单位', dataIndex: 'unit', width: 55,
      render: (v, _, idx) => renderCell(v, idx, 'unit') },
    { title: '类别', dataIndex: 'category', width: 100,
      render: (v, _, idx) => renderCell(v, idx, 'category', 'select',
        CATEGORY_OPTIONS.map(c => ({ label: c, value: c }))) },
    { title: '单价', dataIndex: 'unit_price', width: 80, align: 'right',
      render: (v, _, idx) => renderCell(v, idx, 'unit_price', 'number') },
    { title: '制造商', dataIndex: 'manufacturer', width: 100, ellipsis: true,
      render: (v, _, idx) => renderCell(v, idx, 'manufacturer') },
    { title: '制造商料号', dataIndex: 'manufacturer_pn', width: 100, ellipsis: true,
      render: (v, _, idx) => renderCell(v, idx, 'manufacturer_pn') },
    { title: '供应商', dataIndex: 'supplier', width: 100, ellipsis: true,
      render: (v, _, idx) => renderCell(v, idx, 'supplier') },
    { title: '交期(天)', dataIndex: 'lead_time_days', width: 75, align: 'right',
      render: (v, _, idx) => renderCell(v, idx, 'lead_time_days', 'number') },
    { title: '采购类型', dataIndex: 'procurement_type', width: 100,
      render: (v, _, idx) => renderCell(v, idx, 'procurement_type', 'select', PROCUREMENT_OPTIONS) },
    { title: '关键件', dataIndex: 'is_critical', width: 60, align: 'center',
      render: (v, _, idx) => renderCell(v, idx, 'is_critical', 'checkbox') },
  ];

  const sbomCols: ColumnsType<BOMItemRecord> = [
    { title: '预览', dataIndex: 'thumbnail_url', width: 70, align: 'center',
      render: (url: string) => url ? (
        <ThumbnailCell url={url} />
      ) : (
        <div style={{ width: 60, height: 45, background: '#f5f5f5', display: 'flex', alignItems: 'center', justifyContent: 'center', borderRadius: 4 }}>
          <EyeOutlined style={{ color: '#d9d9d9', fontSize: 16 }} />
        </div>
      ),
    },
    { title: '名称', dataIndex: 'name', width: 120,
      render: (v, _, idx) => renderCell(v, idx, 'name') },
    { title: '数量', dataIndex: 'quantity', width: 60, align: 'right',
      render: (v, _, idx) => renderCell(v, idx, 'quantity', 'number') },
    { title: '单位', dataIndex: 'unit', width: 55,
      render: (v, _, idx) => renderCell(v, idx, 'unit') },
    { title: '材质', dataIndex: 'material_type', width: 110,
      render: (v, _, idx) => renderMaterialTypeCell(v, idx) },
    { title: '工艺类型', dataIndex: 'process_type', width: 90,
      render: (v, _, idx) => renderCell(v, idx, 'process_type', 'select', PROCESS_TYPE_OPTIONS) },
    { title: '2D图纸', dataIndex: 'drawing_2d_file_name', width: renderDrawingColumn ? 150 : 140,
      render: renderDrawingColumn
        ? (_, record) => renderDrawingColumn(record, '2D')
        : (_, record, idx) => renderDrawingUploadCell(record, idx, 'drawing_2d_file_id', 'drawing_2d_file_name', 'drawing_2d_file_size'),
    },
    { title: '3D模型', dataIndex: 'drawing_3d_file_name', width: renderDrawingColumn ? 150 : 140,
      render: renderDrawingColumn
        ? (_, record) => renderDrawingColumn(record, '3D')
        : (_, record, idx) => renderDrawingUploadCell(record, idx, 'drawing_3d_file_id', 'drawing_3d_file_name', 'drawing_3d_file_size'),
    },
    { title: '重量(g)', dataIndex: 'weight_grams', width: 75, align: 'right',
      render: (v, _, idx) => renderCell(v, idx, 'weight_grams', 'number') },
    { title: '目标价', dataIndex: 'target_price', width: 85, align: 'right',
      render: (v, _, idx) => renderCell(v, idx, 'target_price', 'number') },
    { title: '模具费', dataIndex: 'tooling_estimate', width: 85, align: 'right',
      render: (v, _, idx) => renderCell(v, idx, 'tooling_estimate', 'number') },
    { title: '外观件', dataIndex: 'is_appearance_part', width: 60, align: 'center',
      render: (v, _, idx) => renderCell(v, idx, 'is_appearance_part', 'checkbox') },
    { title: '装配方式', dataIndex: 'assembly_method', width: 90,
      render: (v, _, idx) => renderCell(v, idx, 'assembly_method', 'select', ASSEMBLY_METHOD_OPTIONS) },
    { title: '公差', dataIndex: 'tolerance_grade', width: 95,
      render: (v, _, idx) => renderToleranceCell(v, idx) },
    { title: '备注', dataIndex: 'notes', width: 120, ellipsis: true,
      render: (v, _, idx) => renderCell(v, idx, 'notes') },
  ];

  const materialCodeCol: ColumnsType<BOMItemRecord> = showMaterialCode ? [
    { title: '物料编码', width: 120,
      render: (_, record) => {
        const code = record.material?.code || record.specification;
        return (
          <Space size={4}>
            <Text code style={{ fontSize: 11 }}>{code || '-'}</Text>
            {!readonly && onMaterialSearch && (
              <SearchOutlined
                style={{ color: '#1677ff', cursor: 'pointer', fontSize: 12 }}
                onClick={() => onMaterialSearch(record.id)}
              />
            )}
          </Space>
        );
      },
    },
  ] : [];

  const typeCols = bomType === 'SBOM' ? sbomCols : ebomCols;
  const columns: ColumnsType<BOMItemRecord> = [...commonCols, ...materialCodeCol, ...typeCols];

  if (actionColumn) {
    columns.push({
      title: '操作', width: 80, align: 'center', fixed: 'right',
      render: (_, record) => actionColumn(record),
    });
  } else if (showAddDelete) {
    columns.push({
      title: '', width: 40, align: 'center', fixed: 'right',
      render: (_, _record, idx) => (
        <Popconfirm title="确认删除此行？" onConfirm={() => handleDeleteRow(idx)}>
          <Button type="text" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    });
  }

  const defaultScrollX = bomType === 'SBOM' ? 1570 : 1100;
  const finalScrollX = scrollXProp ?? defaultScrollX;

  return (
    <div>
      {showAddDelete && (
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 8 }}>
          <Button type="dashed" size="small" icon={<PlusOutlined />} onClick={handleAddRow}>添加行</Button>
        </div>
      )}
      <Table
        columns={columns}
        dataSource={items}
        rowKey={(r, idx) => r.id || String(idx)}
        size="small"
        pagination={noPagination ? false : (items.length > PAGE_SIZE ? {
          pageSize: PAGE_SIZE, size: 'small', showTotal: (t: number) => `共 ${t} 条`,
          current: currentPage, onChange: (p) => setCurrentPage(p),
        } : false)}
        scroll={{ x: finalScrollX, ...(scrollY ? { y: scrollY } : {}) }}
        style={{ fontSize: 12 }}
        rowClassName={rowClassName}
        locale={{ emptyText: <Empty description={'暂无物料，点击"添加行"或"导入模板"开始'} image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
      />
      <style>{`
        .editable-cell:hover {
          background: #f0f5ff;
        }
      `}</style>
      {preview3D && (
        <STLViewer
          open
          fileUrl={`/api/v1/files/${preview3D.fileId}/3d`}
          fileName={preview3D.fileName}
          onClose={() => setPreview3D(null)}
        />
      )}
    </div>
  );
};

export default BOMEditableTable;
