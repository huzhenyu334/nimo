import React from 'react';
import {
  Input,
  Select,
  Button,
  Space,
  Typography,
  Spin,
  Empty,
  Popconfirm,
  Image,
  Tag,
  Tooltip,
  ColorPicker,
  Form,
  App,
  Upload,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  BgColorsOutlined,
  SaveOutlined,
  CloseOutlined,
  UploadOutlined,
  PaperClipOutlined,
  ArrowLeftOutlined,
  RightOutlined,
} from '@ant-design/icons';
import { taskFormApi } from '@/api/taskForms';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { cmfVariantApi, type CMFVariant, type AppearancePartWithCMF } from '@/api/cmfVariant';
import { useIsMobile } from '@/hooks/useIsMobile';

const { Text } = Typography;

// ========== Option Constants ==========

const GLOSS_OPTIONS = ['高光', '半哑', '哑光', '丝光', '镜面'];
const FINISH_OPTIONS = ['阳极氧化', '喷涂', '电镀', 'PVD', 'IMD', 'UV转印', '丝印', '激光雷雕', '水转印'];
const TEXTURE_OPTIONS = ['光面', '磨砂', '皮纹', '拉丝', '碳纤维纹', '木纹'];
const COATING_OPTIONS = ['UV漆', 'PU漆', '粉末涂装', '电泳', '无'];
const DRAWING_TYPE_OPTIONS = ['丝印', '移印', 'UV转印', '激光雕刻', '水转印', '热转印', '烫金', '其他'];

interface DrawingFile {
  file_id: string;
  file_name: string;
  url: string;
}

const parseDrawings = (v: any): DrawingFile[] => {
  if (!v) return [];
  if (Array.isArray(v)) return v;
  try { return JSON.parse(v); } catch { return []; }
};

// ========== Props ==========

interface CMFEditControlProps {
  projectId: string;
  taskId: string;
  readonly?: boolean;
}

// ========== 色块组件 ==========
const ColorSwatch: React.FC<{ hex?: string; size?: number }> = ({ hex, size = 16 }) => {
  if (!hex) return null;
  return (
    <span style={{
      display: 'inline-block', width: size, height: size, borderRadius: 3,
      backgroundColor: hex, border: '1px solid #d9d9d9', verticalAlign: 'middle',
      boxShadow: '0 1px 2px rgba(0,0,0,0.08)',
    }} />
  );
};

// ========== Click-to-edit cell (BOM table style) ==========
const EditableCell: React.FC<{
  value: any;
  field: string;
  type?: 'text' | 'select' | 'color';
  options?: { label: string; value: string }[];
  readonly?: boolean;
  onSave: (field: string, value: any) => void;
}> = ({ value, field, type = 'text', options, readonly, onSave }) => {
  const [editing, setEditing] = React.useState(false);

  if (readonly) {
    if (type === 'color') {
      return (
        <div style={{ display: 'flex', alignItems: 'center', gap: 4, minHeight: 22, padding: '0 2px' }}>
          <ColorSwatch hex={value} />
          <span>{value || <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}</span>
        </div>
      );
    }
    return (
      <div style={{ minHeight: 22, padding: '0 2px' }}>
        {value ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
      </div>
    );
  }

  if (editing) {
    if (type === 'color') {
      return (
        <ColorPicker
          showText
          format="hex"
          size="small"
          defaultValue={value || undefined}
          onChangeComplete={(c) => {
            onSave(field, c.toHexString());
            setEditing(false);
          }}
          onOpenChange={(open) => { if (!open) setEditing(false); }}
          open
        />
      );
    }
    if (type === 'select' && options) {
      return (
        <Select
          size="small"
          autoFocus
          defaultValue={value || undefined}
          defaultOpen
          style={{ width: '100%' }}
          options={options}
          allowClear
          onChange={(v) => { onSave(field, v || ''); setEditing(false); }}
          onBlur={() => setEditing(false)}
        />
      );
    }
    return (
      <Input
        size="small"
        autoFocus
        defaultValue={value || ''}
        onBlur={(e) => { onSave(field, e.target.value); setEditing(false); }}
        onPressEnter={(e) => { onSave(field, (e.target as HTMLInputElement).value); setEditing(false); }}
      />
    );
  }

  // Display mode
  const displayValue = type === 'select' && options
    ? options.find(o => o.value === value)?.label || value
    : value;

  return (
    <div
      style={{ cursor: 'pointer', minHeight: 22, padding: '0 2px', borderRadius: 2, display: 'flex', alignItems: 'center', gap: 4 }}
      className="editable-cell"
      onClick={() => setEditing(true)}
    >
      {type === 'color' && <ColorSwatch hex={value} />}
      {displayValue ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
    </div>
  );
};

// ========== CMF Table Row ==========
const CMFTableRow: React.FC<{
  variant: CMFVariant;
  projectId: string;
  readonly: boolean;
}> = ({ variant, projectId, readonly }) => {
  const { message } = App.useApp();
  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: (data: Partial<CMFVariant>) =>
      cmfVariantApi.updateVariant(projectId, variant.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appearance-parts', projectId] });
    },
    onError: (err: any) => message.error(err?.response?.data?.message || '保存失败'),
  });

  const deleteMutation = useMutation({
    mutationFn: () => cmfVariantApi.deleteVariant(projectId, variant.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['appearance-parts', projectId] });
    },
    onError: (err: any) => message.error(err?.response?.data?.message || '删除失败'),
  });

  const handleSave = (field: string, value: any) => {
    if (value === variant[field as keyof CMFVariant]) return;
    updateMutation.mutate({ [field]: value });
  };

  const renderImageUrl = variant.reference_image_file_id
    ? (variant.reference_image_url || `/uploads/${variant.reference_image_file_id}/image`)
    : undefined;
  const drawings = parseDrawings(variant.process_drawings);

  const handleRenderUpload = async (file: File) => {
    try {
      const result = await taskFormApi.uploadFile(file);
      updateMutation.mutate({
        reference_image_file_id: result.id,
        reference_image_url: result.url,
      });
    } catch { message.error('上传失败'); }
    return false;
  };

  const handleDrawingUpload = async (file: File) => {
    try {
      const result = await taskFormApi.uploadFile(file);
      const newDrawings = [...drawings, { file_id: result.id, file_name: result.filename || file.name, url: result.url }];
      updateMutation.mutate({
        process_drawings: JSON.stringify(newDrawings) as any,
      });
    } catch { message.error('上传失败'); }
    return false;
  };

  const glossOpts = GLOSS_OPTIONS.map(o => ({ label: o, value: o }));
  const finishOpts = FINISH_OPTIONS.map(o => ({ label: o, value: o }));
  const textureOpts = TEXTURE_OPTIONS.map(o => ({ label: o, value: o }));
  const coatingOpts = COATING_OPTIONS.map(o => ({ label: o, value: o }));

  return (
    <tr style={{ borderBottom: '1px solid #f0f0f0' }}>
      <td style={tdStyle}>
        <Tag color="processing" style={{ margin: 0, fontSize: 11 }}>V{variant.variant_index}</Tag>
      </td>
      <td style={tdStyle}>
        <EditableCell value={variant.color_hex} field="color_hex" type="color" readonly={readonly} onSave={handleSave} />
      </td>
      <td style={tdStyle}>
        <EditableCell value={variant.pantone_code} field="pantone_code" readonly={readonly} onSave={handleSave} />
      </td>
      <td style={tdStyle}>
        <EditableCell value={variant.gloss_level} field="gloss_level" type="select" options={glossOpts} readonly={readonly} onSave={handleSave} />
      </td>
      <td style={tdStyle}>
        <EditableCell value={variant.finish} field="finish" type="select" options={finishOpts} readonly={readonly} onSave={handleSave} />
      </td>
      <td style={tdStyle}>
        <EditableCell value={variant.texture} field="texture" type="select" options={textureOpts} readonly={readonly} onSave={handleSave} />
      </td>
      <td style={tdStyle}>
        <EditableCell value={variant.coating} field="coating" type="select" options={coatingOpts} readonly={readonly} onSave={handleSave} />
      </td>
      <td style={tdStyle}>
        {renderImageUrl ? (
          <Image src={renderImageUrl} width={40} height={30}
            style={{ objectFit: 'cover', borderRadius: 3 }}
            fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iMzAiPjxyZWN0IGZpbGw9IiNmNWY1ZjUiIHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiLz48L3N2Zz4=" />
        ) : readonly ? (
          <Text type="secondary" style={{ fontSize: 11 }}>-</Text>
        ) : (
          <Upload showUploadList={false} accept="image/*" beforeUpload={(file) => { handleRenderUpload(file); return false; }}>
            <Button size="small" type="text" icon={<UploadOutlined />} style={{ fontSize: 11, padding: '0 4px' }} />
          </Upload>
        )}
      </td>
      <td style={tdStyle}>
        {drawings.length > 0 ? (
          <Space size={2} direction="vertical">
            {variant.process_drawing_type && <Tag style={{ fontSize: 10, margin: 0 }}>{variant.process_drawing_type}</Tag>}
            {drawings.map(f => (
              <Tooltip key={f.file_id} title={f.file_name}>
                <a href={f.url} target="_blank" rel="noopener noreferrer" style={{ fontSize: 11 }}>
                  <PaperClipOutlined style={{ color: '#1677ff' }} />
                </a>
              </Tooltip>
            ))}
          </Space>
        ) : readonly ? (
          <Text type="secondary" style={{ fontSize: 11 }}>-</Text>
        ) : (
          <Upload showUploadList={false} accept=".pdf,.dwg,.dxf,.ai,.cdr,image/*" beforeUpload={(file) => { handleDrawingUpload(file); return false; }}>
            <Button size="small" type="text" icon={<UploadOutlined />} style={{ fontSize: 11, padding: '0 4px' }} />
          </Upload>
        )}
      </td>
      <td style={tdStyle}>
        <EditableCell value={variant.notes} field="notes" readonly={readonly} onSave={handleSave} />
      </td>
      {!readonly && (
        <td style={{ ...tdStyle, textAlign: 'center' }}>
          <Popconfirm title="确认删除此CMF方案？" onConfirm={() => deleteMutation.mutate()}>
            <Button size="small" type="text" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </td>
      )}
    </tr>
  );
};

const tdStyle: React.CSSProperties = {
  padding: '6px 8px',
  fontSize: 12,
  verticalAlign: 'middle',
};

const thStyle: React.CSSProperties = {
  padding: '6px 8px',
  fontSize: 11,
  fontWeight: 500,
  color: '#8c8c8c',
  textAlign: 'left',
  borderBottom: '1px solid #e8e8e8',
  background: '#fafafa',
  whiteSpace: 'nowrap',
};

// ========== Part Section (BOM sub-category style) ==========
const PartSection: React.FC<{
  part: AppearancePartWithCMF;
  projectId: string;
  readonly: boolean;
  onAddVariant: (itemId: string) => void;
  addLoading: boolean;
}> = ({ part, projectId, readonly, onAddVariant, addLoading }) => {
  const item = part.bom_item;
  const variants = part.cmf_variants || [];

  return (
    <div style={{ marginBottom: 12 }}>
      {/* Section header — matches BOM sub-category style */}
      <div style={{
        display: 'flex', alignItems: 'center', gap: 8, padding: '4px 12px',
        background: '#fff7e6', borderRadius: '4px 4px 0 0',
        borderBottom: '1px solid #f0f0f0',
      }}>
        {item.thumbnail_url && (
          <img src={item.thumbnail_url} alt="" width={24} height={18}
            style={{ objectFit: 'contain', borderRadius: 3, background: '#f5f5f5' }} />
        )}
        <Text style={{ fontSize: 13, fontWeight: 500 }}>#{item.item_number} {item.name}</Text>
        <Tag style={{ fontSize: 11 }}>{variants.length} 方案</Tag>
        <div style={{ flex: 1 }} />
        {!readonly && (
          <Button
            type="dashed"
            size="small"
            icon={<PlusOutlined />}
            onClick={() => onAddVariant(item.id)}
            loading={addLoading}
          >
            添加方案
          </Button>
        )}
      </div>

      {/* Table */}
      {variants.length > 0 ? (
        <div style={{ overflowX: 'auto', border: '1px solid #f0f0f0', borderTop: 0, borderRadius: '0 0 4px 4px' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', tableLayout: 'auto', minWidth: 900 }}>
            <thead>
              <tr>
                <th style={{ ...thStyle, width: 50 }}>版本</th>
                <th style={{ ...thStyle, width: 100 }}>颜色</th>
                <th style={{ ...thStyle, width: 100 }}>色号(Pantone)</th>
                <th style={{ ...thStyle, width: 80 }}>光泽度</th>
                <th style={{ ...thStyle, width: 90 }}>表面处理</th>
                <th style={{ ...thStyle, width: 80 }}>纹理</th>
                <th style={{ ...thStyle, width: 90 }}>涂层类型</th>
                <th style={{ ...thStyle, width: 50 }}>效果图</th>
                <th style={{ ...thStyle, width: 50 }}>图纸</th>
                <th style={thStyle}>备注</th>
                {!readonly && <th style={{ ...thStyle, width: 40, textAlign: 'center' }}>操作</th>}
              </tr>
            </thead>
            <tbody>
              {variants.map(v => (
                <CMFTableRow key={v.id} variant={v} projectId={projectId} readonly={readonly} />
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div style={{ padding: '16px 0', textAlign: 'center', color: '#999', fontSize: 12, border: '1px solid #f0f0f0', borderTop: 0, borderRadius: '0 0 4px 4px' }}>
          暂无CMF方案
        </div>
      )}
    </div>
  );
};

// ========== 移动端: CMF变体全屏编辑面板 ==========
const CMFMobileEditPanel: React.FC<{
  variant: CMFVariant;
  projectId: string;
  readonly: boolean;
  onClose: () => void;
}> = ({ variant, projectId, readonly, onClose }) => {
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const [visible, setVisible] = React.useState(false);
  const [form] = Form.useForm();

  const [renderUploading, setRenderUploading] = React.useState(false);
  const [drawingUploading, setDrawingUploading] = React.useState(false);
  const [renderImageId, setRenderImageId] = React.useState(variant.reference_image_file_id || '');
  const [renderImageUrl, setRenderImageUrlState] = React.useState(variant.reference_image_url || '');
  const [drawingType, setDrawingType] = React.useState(variant.process_drawing_type || '');
  const [drawings, setDrawings] = React.useState<DrawingFile[]>(parseDrawings(variant.process_drawings));

  React.useEffect(() => {
    requestAnimationFrame(() => setVisible(true));
  }, []);

  const handleClose = () => {
    setVisible(false);
    setTimeout(onClose, 300);
  };

  const updateMutation = useMutation({
    mutationFn: (data: Partial<CMFVariant>) =>
      cmfVariantApi.updateVariant(projectId, variant.id, data),
    onSuccess: () => {
      message.success('已保存');
      queryClient.invalidateQueries({ queryKey: ['appearance-parts', projectId] });
      handleClose();
    },
    onError: (err: any) => message.error(err?.response?.data?.message || '保存失败'),
  });

  const handleSave = () => {
    form.validateFields().then(values => {
      if (values.color_hex && typeof values.color_hex === 'object' && values.color_hex.toHexString) {
        values.color_hex = values.color_hex.toHexString();
      }
      values.reference_image_file_id = renderImageId;
      values.reference_image_url = renderImageUrl;
      values.process_drawing_type = drawingType;
      values.process_drawings = JSON.stringify(drawings);
      updateMutation.mutate(values);
    });
  };

  const handleRenderUpload = async (file: File) => {
    setRenderUploading(true);
    try {
      const result = await taskFormApi.uploadFile(file);
      setRenderImageId(result.id);
      setRenderImageUrlState(result.url);
    } catch { message.error('上传失败'); }
    finally { setRenderUploading(false); }
    return false;
  };

  const handleDrawingUpload = async (file: File) => {
    setDrawingUploading(true);
    try {
      const result = await taskFormApi.uploadFile(file);
      setDrawings(prev => [...prev, { file_id: result.id, file_name: result.filename || file.name, url: result.url }]);
    } catch { message.error('上传失败'); }
    finally { setDrawingUploading(false); }
    return false;
  };

  const existingRenderUrl = renderImageId ? (renderImageUrl || `/uploads/${renderImageId}/image`) : undefined;

  return (
    <div style={{
      position: 'fixed',
      top: 0, left: 0, right: 0, bottom: 0,
      zIndex: 1100,
      background: '#fff',
      transform: visible ? 'translateX(0)' : 'translateX(100%)',
      transition: 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
      display: 'flex',
      flexDirection: 'column',
    }}>
      {/* Header */}
      <div style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '12px 16px',
        borderBottom: '1px solid #f0f0f0',
        background: '#fff',
      }}>
        <ArrowLeftOutlined
          onClick={handleClose}
          style={{ fontSize: 18, cursor: 'pointer', padding: 4 }}
        />
        <div style={{ flex: 1 }}>
          <div style={{ fontSize: 16, fontWeight: 600 }}>
            {readonly ? '查看CMF' : '编辑CMF'}
          </div>
          <div style={{ fontSize: 12, color: '#999' }}>
            V{variant.variant_index} {variant.material_code || ''}
          </div>
        </div>
        {!readonly && (
          <Button
            type="primary"
            onClick={handleSave}
            loading={updateMutation.isPending}
            style={{ borderRadius: 8, fontWeight: 500 }}
          >
            保存
          </Button>
        )}
      </div>

      {/* Form body */}
      <div style={{
        flex: 1, overflow: 'auto', padding: 16,
        WebkitOverflowScrolling: 'touch',
      }}>
        <Form form={form} layout="vertical"
          initialValues={{
            color_hex: variant.color_hex || '',
            pantone_code: variant.pantone_code || '',
            gloss_level: variant.gloss_level || undefined,
            finish: variant.finish || undefined,
            texture: variant.texture || undefined,
            coating: variant.coating || undefined,
            notes: variant.notes || '',
          }}
        >
          {/* 颜色 section */}
          <div style={mfSectionStyle}>
            <div style={mfSectionTitleStyle}>颜色信息</div>
            <Form.Item label="颜色" name="color_hex" style={{ marginBottom: 14 }}>
              <ColorPicker showText format="hex" disabled={readonly} />
            </Form.Item>
            <Form.Item label="色号 (Pantone)" name="pantone_code" style={{ marginBottom: 14 }}>
              <Input placeholder="如: Black 6C" size="large" style={{ borderRadius: 10 }} disabled={readonly} />
            </Form.Item>
          </div>

          {/* 表面处理 section */}
          <div style={mfSectionStyle}>
            <div style={mfSectionTitleStyle}>表面处理</div>
            <Form.Item label="光泽度" name="gloss_level" style={{ marginBottom: 14 }}>
              <Select placeholder="选择光泽度" allowClear size="large"
                disabled={readonly}
                options={GLOSS_OPTIONS.map(o => ({ label: o, value: o }))} />
            </Form.Item>
            <Form.Item label="表面处理" name="finish" style={{ marginBottom: 14 }}>
              <Select placeholder="选择表面处理" allowClear size="large"
                disabled={readonly}
                options={FINISH_OPTIONS.map(o => ({ label: o, value: o }))} />
            </Form.Item>
            <Form.Item label="纹理" name="texture" style={{ marginBottom: 14 }}>
              <Select placeholder="选择纹理" allowClear size="large"
                disabled={readonly}
                options={TEXTURE_OPTIONS.map(o => ({ label: o, value: o }))} />
            </Form.Item>
            <Form.Item label="涂层类型" name="coating" style={{ marginBottom: 14 }}>
              <Select placeholder="选择涂层" allowClear size="large"
                disabled={readonly}
                options={COATING_OPTIONS.map(o => ({ label: o, value: o }))} />
            </Form.Item>
          </div>

          {/* 渲染图 & 图纸 */}
          <div style={mfSectionStyle}>
            <div style={mfSectionTitleStyle}>附件</div>
            <div style={{ marginBottom: 14 }}>
              <Text style={{ fontSize: 13, color: '#666', display: 'block', marginBottom: 6 }}>渲染效果图</Text>
              {existingRenderUrl && (
                <div style={{ position: 'relative', display: 'inline-block', marginBottom: 8 }}>
                  <Image src={existingRenderUrl} width={100} height={75}
                    style={{ objectFit: 'cover', borderRadius: 8 }}
                    fallback="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNDUiPjxyZWN0IGZpbGw9IiNmNWY1ZjUiIHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiLz48L3N2Zz4=" />
                  {!readonly && (
                    <DeleteOutlined
                      style={{ position: 'absolute', top: -6, right: -6, color: '#ff4d4f', cursor: 'pointer', fontSize: 14, background: '#fff', borderRadius: '50%', padding: 3, boxShadow: '0 1px 4px rgba(0,0,0,0.15)' }}
                      onClick={() => { setRenderImageId(''); setRenderImageUrlState(''); }}
                    />
                  )}
                </div>
              )}
              {!readonly && (
                <Upload showUploadList={false} accept="image/*" beforeUpload={(file) => { handleRenderUpload(file); return false; }}>
                  <Button icon={<UploadOutlined />} loading={renderUploading} block style={{ borderRadius: 10 }}>上传效果图</Button>
                </Upload>
              )}
            </div>

            <div style={{ marginBottom: 14 }}>
              <Text style={{ fontSize: 13, color: '#666', display: 'block', marginBottom: 6 }}>工艺图纸</Text>
              {!readonly && (
                <Space size={8} style={{ marginBottom: 8, width: '100%' }}>
                  <Select value={drawingType || undefined} placeholder="图纸类型"
                    allowClear style={{ flex: 1 }}
                    onChange={(v) => setDrawingType(v || '')}
                    options={DRAWING_TYPE_OPTIONS.map(o => ({ label: o, value: o }))} />
                  <Upload showUploadList={false} accept=".pdf,.dwg,.dxf,.ai,.cdr,image/*"
                    beforeUpload={(file) => { handleDrawingUpload(file); return false; }}>
                    <Button icon={<UploadOutlined />} loading={drawingUploading}>上传</Button>
                  </Upload>
                </Space>
              )}
              {variant.process_drawing_type && !drawingType && (
                <Tag style={{ marginBottom: 4 }}>{variant.process_drawing_type}</Tag>
              )}
              {drawings.map(f => (
                <div key={f.file_id} style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13, marginBottom: 4, padding: '4px 0' }}>
                  <PaperClipOutlined style={{ color: '#1677ff' }} />
                  <a href={f.url} target="_blank" rel="noopener noreferrer"
                    style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {f.file_name}
                  </a>
                  {!readonly && (
                    <DeleteOutlined style={{ color: '#ff4d4f', cursor: 'pointer' }}
                      onClick={() => setDrawings(prev => prev.filter(d => d.file_id !== f.file_id))} />
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* 备注 */}
          <div style={mfSectionStyle}>
            <Form.Item label="备注" name="notes" style={{ marginBottom: 0 }}>
              <Input.TextArea rows={3} placeholder="备注信息" style={{ borderRadius: 10 }} disabled={readonly} />
            </Form.Item>
          </div>

          <div style={{ height: 40 }} />
        </Form>
      </div>
    </div>
  );
};

const mfSectionStyle: React.CSSProperties = {
  marginBottom: 20,
  background: '#fafafa',
  borderRadius: 12,
  padding: 16,
};

const mfSectionTitleStyle: React.CSSProperties = {
  fontSize: 14,
  fontWeight: 600,
  color: '#333',
  marginBottom: 12,
};

// ========== 移动端: CMF摘要卡片 ==========
const MobileVariantSummaryCard: React.FC<{
  variant: CMFVariant;
  onClick: () => void;
}> = ({ variant, onClick }) => {
  const renderImageUrl = variant.reference_image_file_id
    ? (variant.reference_image_url || `/uploads/${variant.reference_image_file_id}/image`)
    : undefined;

  return (
    <div
      onClick={onClick}
      style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '12px 14px',
        background: '#fff',
        borderRadius: 10,
        border: '1px solid #f0f0f0',
        marginBottom: 8,
        cursor: 'pointer',
        transition: 'background 0.15s',
      }}
    >
      {renderImageUrl ? (
        <img src={renderImageUrl} alt="" width={44} height={44}
          style={{ objectFit: 'cover', borderRadius: 8, background: '#f5f5f5', flexShrink: 0 }} />
      ) : variant.color_hex ? (
        <div style={{
          width: 44, height: 44, borderRadius: 8, flexShrink: 0,
          backgroundColor: variant.color_hex,
          border: '1px solid #e8e8e8',
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
        }} />
      ) : (
        <div style={{
          width: 44, height: 44, borderRadius: 8, flexShrink: 0,
          background: '#f5f5f5', display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <BgColorsOutlined style={{ color: '#bfbfbf', fontSize: 18 }} />
        </div>
      )}

      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 2 }}>
          <Tag color="processing" style={{ margin: 0, fontSize: 10, lineHeight: '16px', padding: '0 4px' }}>
            V{variant.variant_index}
          </Tag>
          {variant.material_code && (
            <Text style={{ fontSize: 11, color: '#8c8c8c' }}>{variant.material_code}</Text>
          )}
        </div>
        <div style={{ fontSize: 13, color: '#333', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {variant.color_hex || '未设颜色'}
          {variant.pantone_code ? ` / ${variant.pantone_code}` : ''}
        </div>
        <div style={{ fontSize: 11, color: '#999', marginTop: 2 }}>
          {[variant.gloss_level, variant.finish, variant.texture].filter(Boolean).join(' / ') || '未设置表面处理'}
        </div>
      </div>

      <RightOutlined style={{ color: '#bfbfbf', fontSize: 12, flexShrink: 0 }} />
    </div>
  );
};

// ========== 主组件 ==========

const CMFEditControl: React.FC<CMFEditControlProps> = ({ projectId, taskId: _taskId, readonly = false }) => {
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const isMobile = useIsMobile();
  const [editingVariant, setEditingVariant] = React.useState<CMFVariant | null>(null);

  const { data: parts = [], isLoading } = useQuery({
    queryKey: ['appearance-parts', projectId],
    queryFn: () => cmfVariantApi.getAppearanceParts(projectId),
  });

  const createMutation = useMutation({
    mutationFn: ({ itemId }: { itemId: string }) =>
      cmfVariantApi.createVariant(projectId, itemId, {}),
    onSuccess: () => {
      message.success('已添加CMF方案');
      queryClient.invalidateQueries({ queryKey: ['appearance-parts', projectId] });
    },
    onError: (err: any) => message.error(err?.response?.data?.message || '添加失败'),
  });

  if (isLoading) {
    return <div style={{ textAlign: 'center', padding: 40 }}><Spin tip="加载CMF数据..." /></div>;
  }

  if (parts.length === 0) {
    return <Empty description="未找到外观件。请先在EBOM中将零件标记为外观件。" />;
  }

  // ===== Mobile layout =====
  if (isMobile) {
    return (
      <div>
        <div style={{ marginBottom: 10, display: 'flex', alignItems: 'center', gap: 6 }}>
          <BgColorsOutlined style={{ color: '#1677ff' }} />
          <Text type="secondary" style={{ fontSize: 12 }}>CMF方案</Text>
        </div>

        {parts.map((part: AppearancePartWithCMF) => {
          const item = part.bom_item;
          const variants = part.cmf_variants || [];

          return (
            <div key={item.id} style={{ marginBottom: 16 }}>
              <div style={{
                display: 'flex', alignItems: 'center', gap: 8,
                padding: '8px 0', marginBottom: 4,
              }}>
                {item.thumbnail_url && (
                  <img src={item.thumbnail_url} alt="" width={24} height={18}
                    style={{ objectFit: 'contain', borderRadius: 3, background: '#f5f5f5' }} />
                )}
                <Text strong style={{ fontSize: 13, flex: 1 }}>
                  #{item.item_number} {item.name}
                </Text>
                <Tag color="blue" style={{ fontSize: 10, margin: 0 }}>{variants.length}方案</Tag>
              </div>

              {variants.map(v => (
                <MobileVariantSummaryCard
                  key={v.id}
                  variant={v}
                  onClick={() => setEditingVariant(v)}
                />
              ))}

              {!readonly && (
                <Button
                  type="dashed"
                  icon={<PlusOutlined />}
                  onClick={() => createMutation.mutate({ itemId: item.id })}
                  loading={createMutation.isPending}
                  block
                  style={{ borderRadius: 10, marginTop: 4 }}
                >
                  添加方案
                </Button>
              )}
            </div>
          );
        })}

        {editingVariant && (
          <CMFMobileEditPanel
            variant={editingVariant}
            projectId={projectId}
            readonly={readonly}
            onClose={() => setEditingVariant(null)}
          />
        )}
      </div>
    );
  }

  // ===== Desktop layout: table sections (BOM style) =====
  return (
    <div>
      <div style={{ marginBottom: 12, display: 'flex', alignItems: 'center', gap: 8 }}>
        <BgColorsOutlined style={{ color: '#1677ff' }} />
        <Text strong style={{ fontSize: 14 }}>CMF方案</Text>
        <Tag style={{ fontSize: 11 }}>
          {parts.reduce((n: number, p: AppearancePartWithCMF) => n + (p.cmf_variants?.length || 0), 0)} 方案
        </Tag>
      </div>

      {parts.map((part: AppearancePartWithCMF) => (
        <PartSection
          key={part.bom_item.id}
          part={part}
          projectId={projectId}
          readonly={readonly}
          onAddVariant={(itemId) => createMutation.mutate({ itemId })}
          addLoading={createMutation.isPending}
        />
      ))}
    </div>
  );
};

export default CMFEditControl;
