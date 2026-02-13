import React from 'react';
import { Tag, Typography, Table, Space } from 'antd';
import { ShoppingCartOutlined, LinkOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface ProcurementSource {
  task_code: string;
  field_key: string;
  field_label: string;
  item_count: number;
}

export interface ProcurementControlValue {
  sources?: ProcurementSource[];
  pr_id?: string;
  pr_code?: string;
  pr_status?: string;
  error?: string | null;
}

interface ProcurementControlProps {
  value?: ProcurementControlValue;
}

const statusMap: Record<string, { color: string; text: string }> = {
  created: { color: 'blue', text: '已创建' },
  sourcing: { color: 'processing', text: '寻源中' },
  completed: { color: 'success', text: '已完成' },
  cancelled: { color: 'default', text: '已取消' },
  pending: { color: 'orange', text: '待处理' },
};

const ProcurementControl: React.FC<ProcurementControlProps> = ({ value }) => {
  if (!value) return <Text type="secondary">暂无采购数据</Text>;

  const { sources, pr_code, pr_status, error } = value;
  const statusInfo = statusMap[pr_status || ''] || { color: 'default', text: pr_status || '未知' };

  return (
    <div>
      {error && (
        <div style={{ color: '#ff4d4f', marginBottom: 8 }}>
          <Text type="danger">创建失败: {error}</Text>
        </div>
      )}
      {pr_code && (
        <div style={{ marginBottom: 8 }}>
          <Space>
            <ShoppingCartOutlined />
            <Text strong>采购需求:</Text>
            <Tag icon={<LinkOutlined />} color={statusInfo.color}>{pr_code}</Tag>
            <Text type="secondary">{statusInfo.text}</Text>
          </Space>
        </div>
      )}
      {sources && sources.length > 0 && (
        <Table
          size="small"
          dataSource={sources}
          rowKey={(r) => `${r.task_code}-${r.field_key}`}
          pagination={false}
          columns={[
            { title: '来源任务', dataIndex: 'task_code', width: 120 },
            { title: '字段', dataIndex: 'field_label', width: 150 },
            { title: '物料数', dataIndex: 'item_count', width: 80, align: 'right' as const },
          ]}
        />
      )}
      {!pr_code && !error && !sources?.length && (
        <Text type="secondary">任务启动后将自动创建采购需求</Text>
      )}
    </div>
  );
};

export default ProcurementControl;
