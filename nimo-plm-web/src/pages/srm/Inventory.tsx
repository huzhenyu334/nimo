import React, { useState } from 'react';
import {
  Table, Card, Button, Space, Tag, Input, Select, Modal, Form,
  InputNumber, message, Drawer, Typography, Spin,
} from 'antd';
import {
  PlusOutlined, ReloadOutlined, SearchOutlined, RightOutlined,
  ArrowUpOutlined, ArrowDownOutlined, SwapOutlined,
} from '@ant-design/icons';
import { useIsMobile } from '@/hooks/useIsMobile';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { srmApi, InventoryRecord } from '@/api/srm';
import dayjs from 'dayjs';

const { Search } = Input;
const { Text } = Typography;

const txTypeLabels: Record<string, string> = { in: '入库', out: '出库', adjust: '调整' };
const txTypeColors: Record<string, string> = { in: 'green', out: 'red', adjust: 'blue' };

const Inventory: React.FC = () => {
  const queryClient = useQueryClient();
  const isMobile = useIsMobile();
  const [searchText, setSearchText] = useState('');
  const [lowStockOnly, setLowStockOnly] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<InventoryRecord | null>(null);
  const [modalType, setModalType] = useState<'in' | 'out' | 'adjust' | null>(null);
  const [form] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['srm-inventory', searchText, lowStockOnly, page, pageSize],
    queryFn: () => srmApi.listInventory({
      search: searchText || undefined,
      low_stock: lowStockOnly ? 'true' : undefined,
      page,
      page_size: pageSize,
    }),
  });

  const { data: txData } = useQuery({
    queryKey: ['srm-inventory-tx', currentRecord?.id],
    queryFn: () => srmApi.getInventoryTransactions(currentRecord!.id, { page: 1, page_size: 50 }),
    enabled: !!currentRecord?.id && drawerVisible,
  });

  const stockInMutation = useMutation({
    mutationFn: (values: any) => srmApi.stockIn(values),
    onSuccess: () => {
      message.success('入库成功');
      setModalType(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['srm-inventory'] });
    },
    onError: () => message.error('入库失败'),
  });

  const stockOutMutation = useMutation({
    mutationFn: (values: any) => srmApi.stockOut(values),
    onSuccess: () => {
      message.success('出库成功');
      setModalType(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['srm-inventory'] });
      queryClient.invalidateQueries({ queryKey: ['srm-inventory-tx', currentRecord?.id] });
    },
    onError: (err: any) => message.error(err?.response?.data?.message || '出库失败'),
  });

  const adjustMutation = useMutation({
    mutationFn: (values: any) => srmApi.stockAdjust(values),
    onSuccess: () => {
      message.success('调整成功');
      setModalType(null);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['srm-inventory'] });
      queryClient.invalidateQueries({ queryKey: ['srm-inventory-tx', currentRecord?.id] });
    },
    onError: () => message.error('调整失败'),
  });

  const records = data?.items || [];
  const transactions = txData?.items || [];

  const isLowStock = (r: InventoryRecord) => r.safety_stock > 0 && r.quantity < r.safety_stock;

  const columns = [
    {
      title: '物料名称', dataIndex: 'material_name', key: 'material_name', width: 160, ellipsis: true,
      render: (v: string, record: InventoryRecord) => (
        <Text strong style={{ color: isLowStock(record) ? '#cf1322' : undefined }}>{v}</Text>
      ),
    },
    { title: '物料编码', dataIndex: 'material_code', key: 'material_code', width: 100, ellipsis: true },
    { title: 'MPN', dataIndex: 'mpn', key: 'mpn', width: 110, ellipsis: true },
    {
      title: '库存数量', dataIndex: 'quantity', key: 'quantity', width: 100, align: 'right' as const,
      render: (v: number, record: InventoryRecord) => (
        <Text strong style={{ color: isLowStock(record) ? '#cf1322' : '#389e0d', fontSize: 14 }}>
          {v} {record.unit}
        </Text>
      ),
    },
    {
      title: '安全库存', dataIndex: 'safety_stock', key: 'safety_stock', width: 90, align: 'right' as const,
      render: (v: number) => v > 0 ? v : '-',
    },
    {
      title: '状态', key: 'stock_status', width: 80,
      render: (_: unknown, record: InventoryRecord) => isLowStock(record) ? (
        <Tag color="red">低库存</Tag>
      ) : <Tag color="green">正常</Tag>,
    },
    { title: '库位', dataIndex: 'warehouse', key: 'warehouse', width: 80 },
    {
      title: '供应商', key: 'supplier', width: 100, ellipsis: true,
      render: (_: unknown, record: InventoryRecord) => record.supplier?.name || '-',
    },
    {
      title: '最近入库', dataIndex: 'last_in_date', key: 'last_in_date', width: 100,
      render: (d: string) => d ? dayjs(d).format('MM-DD HH:mm') : '-',
    },
    {
      title: '操作', key: 'action', width: 120,
      render: (_: unknown, record: InventoryRecord) => (
        <Space size="small">
          <Button type="link" size="small" onClick={(e) => { e.stopPropagation(); setCurrentRecord(record); setDrawerVisible(true); }}>
            流水
          </Button>
          <Button type="link" size="small" onClick={(e) => { e.stopPropagation(); setCurrentRecord(record); setModalType('out'); }}>
            出库
          </Button>
        </Space>
      ),
    },
  ];

  const txColumns = [
    {
      title: '类型', dataIndex: 'type', key: 'type', width: 60,
      render: (t: string) => <Tag color={txTypeColors[t]}>{txTypeLabels[t] || t}</Tag>,
    },
    {
      title: '数量', dataIndex: 'quantity', key: 'quantity', width: 80,
      render: (q: number) => <Text style={{ color: q >= 0 ? '#389e0d' : '#cf1322', fontWeight: 600 }}>{q >= 0 ? `+${q}` : q}</Text>,
    },
    { title: '来源', dataIndex: 'reference_type', key: 'reference_type', width: 70 },
    { title: '操作人', dataIndex: 'operator', key: 'operator', width: 80 },
    { title: '备注', dataIndex: 'notes', key: 'notes', width: 140, ellipsis: true },
    {
      title: '时间', dataIndex: 'created_at', key: 'created_at', width: 120,
      render: (d: string) => dayjs(d).format('MM-DD HH:mm'),
    },
  ];

  const handleModalSubmit = () => {
    form.validateFields().then(values => {
      if (modalType === 'in') {
        stockInMutation.mutate(values);
      } else if (modalType === 'out') {
        stockOutMutation.mutate({ inventory_id: currentRecord!.id, ...values });
      } else if (modalType === 'adjust') {
        adjustMutation.mutate({ inventory_id: currentRecord!.id, ...values });
      }
    });
  };

  // ========== Mobile Layout ==========
  if (isMobile) {
    return (
      <div style={{ background: '#f5f5f5', minHeight: '100vh' }}>
        <div style={{ padding: '12px 16px', background: '#fff', position: 'sticky', top: 0, zIndex: 10 }}>
          <Input
            placeholder="搜索物料名称/编码"
            prefix={<SearchOutlined style={{ color: '#bbb' }} />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onPressEnter={() => setPage(1)}
            allowClear
            style={{ borderRadius: 20 }}
          />
        </div>
        <div className="mobile-filter-pills" style={{ padding: '8px 12px' }}>
          <div className={`mobile-filter-pill ${!lowStockOnly ? 'active' : ''}`} onClick={() => setLowStockOnly(false)}>全部</div>
          <div className={`mobile-filter-pill ${lowStockOnly ? 'active' : ''}`} onClick={() => setLowStockOnly(true)}>低库存</div>
        </div>
        <div style={{ padding: '0 12px' }}>
          {isLoading ? (
            <div style={{ textAlign: 'center', padding: 40 }}><Spin tip="加载中..." /></div>
          ) : records.map(record => (
            <div
              key={record.id}
              onClick={() => { setCurrentRecord(record); setDrawerVisible(true); }}
              style={{
                background: '#fff', borderRadius: 10, padding: '12px 14px', marginBottom: 8,
                boxShadow: '0 1px 3px rgba(0,0,0,0.04)', cursor: 'pointer',
                borderLeft: isLowStock(record) ? '3px solid #cf1322' : undefined,
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', marginBottom: 4 }}>
                <Text strong style={{ flex: 1, fontSize: 14 }}>{record.material_name}</Text>
                {isLowStock(record) && <Tag color="red" style={{ margin: 0 }}>低库存</Tag>}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', fontSize: 13, color: '#666', gap: 8 }}>
                {record.material_code && <Text code style={{ fontSize: 11 }}>{record.material_code}</Text>}
                <Text strong style={{ color: isLowStock(record) ? '#cf1322' : '#389e0d', fontSize: 15, marginLeft: 'auto' }}>
                  {record.quantity} {record.unit}
                </Text>
                <RightOutlined style={{ fontSize: 10, color: '#ccc' }} />
              </div>
            </div>
          ))}
          {records.length === 0 && !isLoading && (
            <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>暂无库存记录</div>
          )}
        </div>
        <div
          onClick={() => { setCurrentRecord(null); setModalType('in'); }}
          style={{ position: 'fixed', bottom: 80, right: 20, width: 52, height: 52, borderRadius: 26, background: '#1677ff', color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 4px 12px rgba(22,119,255,0.4)', zIndex: 100, fontSize: 22, cursor: 'pointer' }}
        >
          <PlusOutlined />
        </div>

        {/* Drawer */}
        <Drawer
          title={currentRecord?.material_name || '库存流水'}
          open={drawerVisible}
          onClose={() => { setDrawerVisible(false); setCurrentRecord(null); }}
          width="100%"
        >
          {currentRecord && (
            <>
              <div style={{ marginBottom: 16, padding: 12, background: '#fafafa', borderRadius: 8 }}>
                <Text>当前库存: <Text strong style={{ fontSize: 18, color: isLowStock(currentRecord) ? '#cf1322' : '#389e0d' }}>{currentRecord.quantity} {currentRecord.unit}</Text></Text>
                {currentRecord.safety_stock > 0 && <Text type="secondary" style={{ marginLeft: 12 }}>安全库存: {currentRecord.safety_stock}</Text>}
              </div>
              <Space style={{ marginBottom: 12 }}>
                <Button size="small" icon={<ArrowDownOutlined />} onClick={() => setModalType('out')}>出库</Button>
                <Button size="small" icon={<SwapOutlined />} onClick={() => setModalType('adjust')}>调整</Button>
              </Space>
              <Table columns={txColumns} dataSource={transactions} rowKey="id" size="small" pagination={false} scroll={{ x: 500 }} />
            </>
          )}
        </Drawer>

        {/* Modal */}
        <Modal
          title={modalType === 'in' ? '入库' : modalType === 'out' ? '出库' : '库存调整'}
          open={!!modalType}
          onOk={handleModalSubmit}
          onCancel={() => { setModalType(null); form.resetFields(); }}
          confirmLoading={stockInMutation.isPending || stockOutMutation.isPending || adjustMutation.isPending}
        >
          <Form form={form} layout="vertical">
            {modalType === 'in' && (
              <>
                <Form.Item name="material_name" label="物料名称" rules={[{ required: true }]}>
                  <Input placeholder="物料名称" />
                </Form.Item>
                <Form.Item name="material_code" label="物料编码">
                  <Input placeholder="物料编码" />
                </Form.Item>
                <Form.Item name="quantity" label="数量" rules={[{ required: true }]}>
                  <InputNumber min={0.01} style={{ width: '100%' }} />
                </Form.Item>
                <Form.Item name="warehouse" label="库位">
                  <Input placeholder="库位" />
                </Form.Item>
              </>
            )}
            {modalType === 'out' && (
              <Form.Item name="quantity" label={`出库数量 (当前库存: ${currentRecord?.quantity ?? 0})`} rules={[{ required: true }]}>
                <InputNumber min={0.01} max={currentRecord?.quantity} style={{ width: '100%' }} />
              </Form.Item>
            )}
            {modalType === 'adjust' && (
              <Form.Item name="quantity" label={`调整后数量 (当前: ${currentRecord?.quantity ?? 0})`} rules={[{ required: true }]}>
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            )}
            <Form.Item name="notes" label="备注">
              <Input.TextArea rows={2} />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  }

  // ========== Desktop Layout ==========
  return (
    <div>
      <Card
        title="库存管理"
        extra={
          <Space wrap>
            <Select
              value={lowStockOnly ? 'low' : 'all'}
              onChange={(v) => { setLowStockOnly(v === 'low'); setPage(1); }}
              style={{ width: 110 }}
              options={[{ value: 'all', label: '全部' }, { value: 'low', label: '低库存' }]}
            />
            <Search
              placeholder="搜索物料"
              allowClear
              style={{ width: 200 }}
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              onSearch={() => setPage(1)}
            />
            <Button icon={<ReloadOutlined />} onClick={() => queryClient.invalidateQueries({ queryKey: ['srm-inventory'] })}>
              刷新
            </Button>
            <Button type="primary" icon={<ArrowUpOutlined />} onClick={() => { setCurrentRecord(null); setModalType('in'); }}>
              手动入库
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={records}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1100 }}
          pagination={{
            current: page,
            pageSize,
            total: data?.pagination?.total || 0,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); },
          }}
          onRow={(record) => ({
            onClick: () => { setCurrentRecord(record); setDrawerVisible(true); },
            style: { cursor: 'pointer' },
          })}
        />
      </Card>

      {/* Transaction Drawer */}
      <Drawer
        title={currentRecord?.material_name || '库存流水'}
        open={drawerVisible}
        onClose={() => { setDrawerVisible(false); setCurrentRecord(null); }}
        width={700}
      >
        {currentRecord && (
          <>
            <div style={{ display: 'flex', gap: 16, marginBottom: 16, padding: 12, background: '#fafafa', borderRadius: 8, alignItems: 'center' }}>
              <Text>当前库存: <Text strong style={{ fontSize: 20, color: isLowStock(currentRecord) ? '#cf1322' : '#389e0d' }}>{currentRecord.quantity} {currentRecord.unit}</Text></Text>
              {currentRecord.safety_stock > 0 && <Text type="secondary">安全库存: {currentRecord.safety_stock}</Text>}
              {isLowStock(currentRecord) && <Tag color="red">低于安全库存</Tag>}
              <Space style={{ marginLeft: 'auto' }}>
                <Button size="small" icon={<ArrowDownOutlined />} onClick={() => setModalType('out')}>出库</Button>
                <Button size="small" icon={<SwapOutlined />} onClick={() => setModalType('adjust')}>调整</Button>
              </Space>
            </div>
            <h4>库存流水</h4>
            <Table columns={txColumns} dataSource={transactions} rowKey="id" size="small" pagination={false} />
          </>
        )}
      </Drawer>

      {/* Stock In/Out/Adjust Modal */}
      <Modal
        title={modalType === 'in' ? '手动入库' : modalType === 'out' ? '出库' : '库存调整'}
        open={!!modalType}
        onOk={handleModalSubmit}
        onCancel={() => { setModalType(null); form.resetFields(); }}
        confirmLoading={stockInMutation.isPending || stockOutMutation.isPending || adjustMutation.isPending}
      >
        <Form form={form} layout="vertical">
          {modalType === 'in' && (
            <>
              <Form.Item name="material_name" label="物料名称" rules={[{ required: true, message: '请输入物料名称' }]}>
                <Input placeholder="物料名称" />
              </Form.Item>
              <Form.Item name="material_code" label="物料编码">
                <Input placeholder="物料编码" />
              </Form.Item>
              <Space style={{ width: '100%' }}>
                <Form.Item name="quantity" label="数量" rules={[{ required: true, message: '请输入数量' }]}>
                  <InputNumber min={0.01} style={{ width: 150 }} />
                </Form.Item>
                <Form.Item name="unit" label="单位" initialValue="pcs">
                  <Input style={{ width: 80 }} />
                </Form.Item>
              </Space>
              <Form.Item name="warehouse" label="库位">
                <Input placeholder="仓库/库位" />
              </Form.Item>
            </>
          )}
          {modalType === 'out' && (
            <Form.Item name="quantity" label={`出库数量 (当前库存: ${currentRecord?.quantity ?? 0} ${currentRecord?.unit || ''})`} rules={[{ required: true }]}>
              <InputNumber min={0.01} max={currentRecord?.quantity} style={{ width: '100%' }} />
            </Form.Item>
          )}
          {modalType === 'adjust' && (
            <Form.Item name="quantity" label={`调整后数量 (当前: ${currentRecord?.quantity ?? 0} ${currentRecord?.unit || ''})`} rules={[{ required: true }]}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          )}
          <Form.Item name="notes" label="备注">
            <Input.TextArea rows={2} placeholder="操作说明" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Inventory;
