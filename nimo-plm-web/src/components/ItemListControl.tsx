import React, { useState } from 'react';
import { Table, Input, InputNumber, Button, Typography, Popconfirm, Empty } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Text } = Typography;

export interface ItemListItem {
  name: string;
  unit: string;
  quantity: number;
  unit_price: number;
}

export interface ItemListValue {
  items: ItemListItem[];
  item_count: number;
}

interface ItemListControlProps {
  value?: ItemListValue;
  onChange?: (value: ItemListValue) => void;
  onSaveDraft?: () => void;
  readonly?: boolean;
  title?: string;
}

const ItemListControl: React.FC<ItemListControlProps> = ({
  value,
  onChange,
  onSaveDraft,
  readonly = false,
  title = '清单',
}) => {
  const [editingCell, setEditingCell] = useState<{ rowIdx: number; field: string } | null>(null);

  const items: ItemListItem[] = value?.items || [];

  const emitChange = (newItems: ItemListItem[]) => {
    onChange?.({ items: newItems, item_count: newItems.length });
    onSaveDraft?.();
  };

  const handleCellSave = (idx: number, field: string, val: any) => {
    if (items[idx] && (items[idx] as any)[field] === val) {
      setEditingCell(null);
      return;
    }
    const newItems = items.map((item, i) => i === idx ? { ...item, [field]: val } : item);
    emitChange(newItems);
    setEditingCell(null);
  };

  const handleAddRow = () => {
    emitChange([...items, { name: '', unit: 'pcs', quantity: 1, unit_price: 0 }]);
  };

  const handleDeleteRow = (idx: number) => {
    emitChange(items.filter((_, i) => i !== idx));
  };

  const renderCell = (val: any, idx: number, field: string, type: 'text' | 'number' = 'text') => {
    if (readonly) {
      return <div style={{ minHeight: 22, padding: '0 2px' }}>{val ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}</div>;
    }

    const isEditing = editingCell?.rowIdx === idx && editingCell?.field === field;

    if (isEditing) {
      if (type === 'number') {
        return (
          <InputNumber
            size="small"
            autoFocus
            defaultValue={typeof val === 'string' ? parseFloat(val) || 0 : val}
            style={{ width: '100%' }}
            onBlur={(e) => handleCellSave(idx, field, parseFloat((e.target as HTMLInputElement).value) || 0)}
            onPressEnter={(e) => handleCellSave(idx, field, parseFloat((e.target as HTMLInputElement).value) || 0)}
          />
        );
      }
      return (
        <Input
          size="small"
          autoFocus
          defaultValue={val}
          onBlur={(e) => handleCellSave(idx, field, e.target.value)}
          onPressEnter={(e) => handleCellSave(idx, field, (e.target as HTMLInputElement).value)}
        />
      );
    }

    return (
      <div
        style={{ cursor: 'pointer', minHeight: 22, padding: '0 2px', borderRadius: 2 }}
        className="editable-cell"
        onClick={() => setEditingCell({ rowIdx: idx, field })}
      >
        {val ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
      </div>
    );
  };

  const columns: ColumnsType<ItemListItem> = [
    { title: '序号', width: 55, align: 'center' as const, render: (_, __, idx) => idx + 1 },
    { title: '名称', dataIndex: 'name', width: 200, render: (v, _, idx) => renderCell(v, idx, 'name') },
    { title: '单位', dataIndex: 'unit', width: 80, render: (v, _, idx) => renderCell(v, idx, 'unit') },
    { title: '数量', dataIndex: 'quantity', width: 100, align: 'right' as const, render: (v, _, idx) => renderCell(v, idx, 'quantity', 'number') },
    { title: '单价', dataIndex: 'unit_price', width: 100, align: 'right' as const, render: (v, _, idx) => renderCell(v, idx, 'unit_price', 'number') },
  ];

  if (!readonly) {
    columns.push({
      title: '', width: 40, align: 'center' as const,
      render: (_, __, idx) => (
        <Popconfirm title="确认删除此行？" onConfirm={() => handleDeleteRow(idx)}>
          <Button type="text" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    });
  }

  return (
    <div>
      {!readonly && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
          <Text strong style={{ fontSize: 13 }}>{title}</Text>
          <Button type="dashed" size="small" icon={<PlusOutlined />} onClick={handleAddRow}>添加行</Button>
        </div>
      )}
      <Table
        columns={columns}
        dataSource={items}
        rowKey={(_, idx) => String(idx)}
        size="small"
        pagination={false}
        scroll={{ x: 575 }}
        style={{ fontSize: 12 }}
        locale={{ emptyText: <Empty description={readonly ? '暂无数据' : '点击"添加行"开始录入'} image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
      />
      {!readonly && (
        <style>{`.editable-cell:hover { background: #f0f5ff; }`}</style>
      )}
    </div>
  );
};

export const ToolingListControl: React.FC<Omit<ItemListControlProps, 'title'>> = (props) => (
  <ItemListControl {...props} title="治具清单" />
);

export const ConsumableListControl: React.FC<Omit<ItemListControlProps, 'title'>> = (props) => (
  <ItemListControl {...props} title="组装辅料清单" />
);

export default ItemListControl;
