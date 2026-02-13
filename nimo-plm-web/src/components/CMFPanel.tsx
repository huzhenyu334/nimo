import React, { useState, useEffect } from 'react';
import { Collapse, Card, Space, Typography, Spin, Empty, Tag, Tooltip, Image } from 'antd';
import { cmfApi, type CMFDesign } from '@/api/cmf';

const { Text } = Typography;

interface CMFPanelProps {
  projectId: string;
}

const CMFPanel: React.FC<CMFPanelProps> = ({ projectId }) => {
  const [designs, setDesigns] = useState<CMFDesign[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    cmfApi.listDesignsByProject(projectId)
      .then(setDesigns)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [projectId]);

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 40 }}><Spin /></div>;
  }

  if (designs.length === 0) {
    return <Empty description="暂无CMF方案" />;
  }

  // Group by bom_item
  const grouped = new Map<string, { part: { id: string; name: string; item_number: number; material_type?: string }; designs: CMFDesign[] }>();
  for (const d of designs) {
    const key = d.bom_item_id;
    if (!grouped.has(key)) {
      grouped.set(key, {
        part: d.bom_item || { id: key, name: '未知零件', item_number: 0 },
        designs: [],
      });
    }
    grouped.get(key)!.designs.push(d);
  }

  const collapseItems = Array.from(grouped.entries()).map(([key, group]) => ({
    key,
    label: (
      <Space size={8}>
        <Text strong>#{group.part.item_number} {group.part.name}</Text>
        {group.part.material_type && <Tag>{group.part.material_type}</Tag>}
        <Tag color="blue">{group.designs.length} 个方案</Tag>
      </Space>
    ),
    children: (
      <div>
        {group.designs.map(design => (
          <Card key={design.id} size="small" style={{ marginBottom: 8 }}
            title={<Text strong style={{ fontSize: 13 }}>{design.scheme_name || '未命名方案'}</Text>}
          >
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '4px 12px', fontSize: 12 }}>
              <div><Text type="secondary">颜色：</Text>{design.color || '-'}</div>
              <div><Text type="secondary">色号：</Text>{design.color_code || '-'}</div>
              <div><Text type="secondary">光泽度：</Text>{design.gloss_level || '-'}</div>
              <div><Text type="secondary">表面处理：</Text>{design.surface_treatment || '-'}</div>
              <div><Text type="secondary">纹理：</Text>{design.texture_pattern || '-'}</div>
              <div><Text type="secondary">涂层：</Text>{design.coating_type || '-'}</div>
            </div>
            {design.render_image_file_id && (
              <div style={{ marginTop: 8 }}>
                <Text type="secondary" style={{ fontSize: 12 }}>效果图：</Text>
                <Image
                  src={`/uploads/${design.render_image_file_id}/${design.render_image_file_name}`}
                  width={80}
                  height={60}
                  style={{ objectFit: 'cover', borderRadius: 4, marginLeft: 8 }}
                />
              </div>
            )}
            {design.drawings && design.drawings.length > 0 && (
              <div style={{ marginTop: 8 }}>
                <Text type="secondary" style={{ fontSize: 12 }}>工艺图纸：</Text>
                {design.drawings.map(dr => (
                  <span key={dr.id} style={{ marginLeft: 8 }}>
                    <Tag color="blue" style={{ fontSize: 11 }}>{dr.drawing_type}</Tag>
                    <Tooltip title={dr.file_name}>
                      <a href={`/uploads/${dr.file_id}/${dr.file_name}`} target="_blank" rel="noreferrer" style={{ fontSize: 12 }}>
                        {dr.file_name}
                      </a>
                    </Tooltip>
                  </span>
                ))}
              </div>
            )}
            {design.notes && (
              <div style={{ marginTop: 4, fontSize: 12 }}>
                <Text type="secondary">备注：</Text>{design.notes}
              </div>
            )}
          </Card>
        ))}
      </div>
    ),
  }));

  return (
    <Collapse
      defaultActiveKey={Array.from(grouped.keys())}
      items={collapseItems}
    />
  );
};

export default CMFPanel;
