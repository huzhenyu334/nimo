import React, { useState, useMemo } from 'react';
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
  ColorPicker,
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

const { Text } = Typography;

// ========== Types ==========

export interface EditableColumn {
  key: string;
  title: string | React.ReactNode;
  width?: number;
  align?: 'left' | 'center' | 'right';
  type?: 'text' | 'number' | 'select' | 'checkbox' | 'color';
  options?: { label: string; value: string }[];
  editable?: boolean;
  ellipsis?: boolean;
  fixed?: 'left' | 'right';
  render?: (value: any, record: Record<string, any>, index: number) => React.ReactNode;
  formatValue?: (value: any, record: Record<string, any>) => React.ReactNode;
}

export interface EditableTableProps {
  columns: EditableColumn[];
  items: Record<string, any>[];
  onChange?: (items: Record<string, any>[]) => void;
  onCellSave?: (record: Record<string, any>, field: string, value: any, index: number) => void;
  onDeleteRow?: (record: Record<string, any>, index: number) => void;
  readonly?: boolean;
  rowKey?: string | ((record: Record<string, any>, index?: number) => string);
  pageSize?: number;
  showDelete?: boolean;
  deleteConfirmText?: string;
  emptyText?: React.ReactNode;
  style?: React.CSSProperties;
}

// ========== Component ==========

const EditableTable: React.FC<EditableTableProps> = ({
  columns,
  items,
  onChange,
  onCellSave,
  onDeleteRow,
  readonly = false,
  rowKey = 'id',
  pageSize = 10,
  showDelete,
  deleteConfirmText = '确认删除此行？',
  emptyText,
  style,
}) => {
  const [editingCell, setEditingCell] = useState<{ rowIdx: number; field: string } | null>(null);
  const [currentPage, setCurrentPage] = useState(1);

  const toGlobalIdx = (pageIdx: number) =>
    items.length > pageSize ? (currentPage - 1) * pageSize + pageIdx : pageIdx;

  // ========== Cell save ==========

  const handleCellSave = (idx: number, field: string, value: any) => {
    const gi = toGlobalIdx(idx);
    if (items[gi] && items[gi][field] === value) {
      setEditingCell(null);
      return;
    }
    if (onCellSave) {
      onCellSave(items[gi], field, value, gi);
    } else if (onChange) {
      const newItems = items.map((item, i) => i === gi ? { ...item, [field]: value } : item);
      onChange(newItems);
    }
    setEditingCell(null);
  };

  const handleDeleteRow = (idx: number) => {
    const gi = toGlobalIdx(idx);
    if (onDeleteRow) {
      onDeleteRow(items[gi], gi);
    } else if (onChange) {
      onChange(items.filter((_, i) => i !== gi));
    }
  };

  // ========== Click-to-edit cell renderer ==========

  const renderCell = (
    value: any, idx: number, field: string,
    type: 'text' | 'number' | 'select' | 'checkbox' | 'color' = 'text',
    options?: { label: string; value: string }[],
    formatValue?: (v: any, record: Record<string, any>) => React.ReactNode,
  ) => {
    const gi = toGlobalIdx(idx);

    // Checkbox: always inline
    if (type === 'checkbox') {
      return (
        <Checkbox
          checked={!!value}
          disabled={readonly}
          onChange={(e) => handleCellSave(idx, field, e.target.checked)}
        />
      );
    }

    // Format display value — guard against GORM relation objects
    let displayValue: any = value;
    if (displayValue != null && typeof displayValue === 'object' && type !== 'color') {
      displayValue = null;
    }
    if (formatValue) {
      displayValue = formatValue(value, items[gi]);
    } else if (type === 'select' && options && displayValue) {
      displayValue = options.find(o => o.value === displayValue)?.label || displayValue;
    }

    // Readonly mode
    if (readonly) {
      if (type === 'color') {
        return (
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, minHeight: 22, padding: '0 2px' }}>
            {value && <span style={{
              display: 'inline-block', width: 16, height: 16, borderRadius: 3,
              backgroundColor: value, border: '1px solid #d9d9d9', verticalAlign: 'middle',
            }} />}
            <span>{value || <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}</span>
          </div>
        );
      }
      return (
        <div style={{ minHeight: 22, padding: '0 2px' }}>
          {displayValue ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
        </div>
      );
    }

    const isEditing = editingCell?.rowIdx === gi && editingCell?.field === field;

    if (isEditing) {
      if (type === 'color') {
        return (
          <ColorPicker
            showText
            format="hex"
            size="small"
            defaultValue={value || undefined}
            onChangeComplete={(c) => handleCellSave(idx, field, c.toHexString())}
            onOpenChange={(open) => { if (!open) setEditingCell(null); }}
            open
          />
        );
      }
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
            defaultValue={value || undefined}
            defaultOpen
            style={{ width: '100%' }}
            options={options}
            allowClear
            onChange={(v) => handleCellSave(idx, field, v ?? '')}
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
        style={{ cursor: 'pointer', minHeight: 22, padding: '0 2px', borderRadius: 2, display: 'flex', alignItems: 'center', gap: 4 }}
        className="editable-cell"
        onClick={() => setEditingCell({ rowIdx: gi, field })}
      >
        {type === 'color' && value && <span style={{
          display: 'inline-block', width: 16, height: 16, borderRadius: 3,
          backgroundColor: value, border: '1px solid #d9d9d9', verticalAlign: 'middle',
        }} />}
        {displayValue ?? <Text type="secondary" style={{ fontSize: 11 }}>-</Text>}
      </div>
    );
  };

  // ========== Build Ant Design columns ==========

  const antColumns: ColumnsType<Record<string, any>> = useMemo(() => {
    const cols: ColumnsType<Record<string, any>> = [];

    for (const col of columns) {
      if (col.render) {
        cols.push({
          title: col.title,
          dataIndex: col.key,
          width: col.width,
          align: col.align as any,
          ellipsis: col.ellipsis,
          fixed: col.fixed as any,
          render: col.render,
        });
      } else if (col.editable === false) {
        cols.push({
          title: col.title,
          dataIndex: col.key,
          width: col.width,
          align: col.align as any,
          ellipsis: col.ellipsis,
          fixed: col.fixed as any,
          render: col.formatValue
            ? (v: any, record: any) => col.formatValue!(v, record)
            : undefined,
        });
      } else {
        cols.push({
          title: col.title,
          dataIndex: col.key,
          width: col.width,
          align: col.align as any,
          ellipsis: col.ellipsis,
          fixed: col.fixed as any,
          render: (v: any, _: any, idx: number) =>
            renderCell(v, idx, col.key, col.type || 'text', col.options, col.formatValue),
        });
      }
    }

    // Delete column
    const shouldShowDelete = showDelete !== undefined ? showDelete : !readonly;
    if (shouldShowDelete) {
      cols.push({
        title: '', width: 40, align: 'center', fixed: 'right',
        render: (_, _record, idx) => (
          <Popconfirm title={deleteConfirmText} onConfirm={() => handleDeleteRow(idx)}>
            <Button type="text" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        ),
      });
    }

    return cols;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columns, readonly, editingCell, items, currentPage, showDelete]);

  const scrollX = antColumns.reduce((sum, c) => sum + ((c.width as number) || 100), 0);

  const getRowKey = typeof rowKey === 'function'
    ? rowKey
    : (record: Record<string, any>, idx?: number) => record[rowKey as string] || String(idx);

  return (
    <div style={style}>
      <Table
        columns={antColumns}
        dataSource={items}
        rowKey={getRowKey}
        size="small"
        pagination={items.length > pageSize ? {
          pageSize,
          size: 'small',
          showTotal: (t: number) => `共 ${t} 条`,
          current: currentPage,
          onChange: (p) => setCurrentPage(p),
        } : false}
        scroll={{ x: scrollX }}
        style={{ fontSize: 12 }}
        locale={{ emptyText: emptyText || <Empty description="暂无数据" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
      />
      <style>{`.editable-cell:hover { background: #f0f5ff; }`}</style>
    </div>
  );
};

export default EditableTable;
