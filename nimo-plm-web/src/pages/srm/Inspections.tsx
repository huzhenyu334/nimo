import React, { useState } from 'react';
import {
  Table,
  Card,
  Button,
  Space,
  Tag,
  Input,
  Select,
  message,
  Drawer,
  Descriptions,
  Form,
  InputNumber,
  Spin,
} from 'antd';
import { ReloadOutlined, CheckCircleOutlined, SearchOutlined, RightOutlined } from '@ant-design/icons';
import { useIsMobile } from '@/hooks/useIsMobile';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { srmApi, Inspection, InspectionItem } from '@/api/srm';
import dayjs from 'dayjs';

const { Search } = Input;

const statusLabels: Record<string, string> = {
  pending: '待检验', inspecting: '检验中', in_progress: '检验中', completed: '已完成',
};
const statusColors: Record<string, string> = {
  pending: 'default', inspecting: 'processing', in_progress: 'processing', completed: 'success',
};

const resultLabels: Record<string, string> = {
  passed: '合格', failed: '不合格', conditional: '条件放行', '': '待检',
};
const resultColors: Record<string, string> = {
  passed: 'green', failed: 'red', conditional: 'orange',
};

const Inspections: React.FC = () => {
  const queryClient = useQueryClient();
  const isMobile = useIsMobile();
  const [searchText, setSearchText] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>();
  const [filterResult, setFilterResult] = useState<string>();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [currentInsp, setCurrentInsp] = useState<Inspection | null>(null);
  const [completeForm] = Form.useForm();
  const [itemResults, setItemResults] = useState<Record<string, Partial<InspectionItem>>>({});

  const { data, isLoading } = useQuery({
    queryKey: ['srm-inspections', searchText, filterStatus, filterResult, page, pageSize],
    queryFn: () =>
      srmApi.listInspections({
        status: filterStatus,
        result: filterResult,
        page,
        page_size: pageSize,
      }),
  });

  const { data: inspDetail } = useQuery({
    queryKey: ['srm-inspection', currentInsp?.id],
    queryFn: () => srmApi.getInspection(currentInsp!.id),
    enabled: !!currentInsp?.id && drawerVisible,
  });

  const completeMutation = useMutation({
    mutationFn: (values: { result: string; notes?: string; items?: any[] }) =>
      srmApi.completeInspection(currentInsp!.id, values),
    onSuccess: () => {
      message.success('检验完成');
      completeForm.resetFields();
      setItemResults({});
      queryClient.invalidateQueries({ queryKey: ['srm-inspections'] });
      queryClient.invalidateQueries({ queryKey: ['srm-inspection', currentInsp?.id] });
    },
    onError: () => message.error('操作失败'),
  });

  const columns = [
    {
      title: '检验编码',
      dataIndex: 'inspection_code',
      key: 'inspection_code',
      width: 140,
      render: (text: string) => <span style={{ fontFamily: 'monospace' }}>{text}</span>,
    },
    {
      title: 'PO编码',
      key: 'po_code',
      width: 130,
      render: (_: unknown, record: Inspection) => {
        const poCode = record.po?.po_code;
        return poCode ? <span style={{ fontFamily: 'monospace' }}>{poCode}</span> : '-';
      },
    },
    { title: '物料', dataIndex: 'material_name', key: 'material_name', width: 160, ellipsis: true },
    {
      title: '供应商',
      key: 'supplier_name',
      width: 120,
      render: (_: unknown, record: Inspection) => {
        return record.supplier?.name || record.supplier?.short_name || '-';
      },
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 80,
      render: (q?: number) => q ?? '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (s: string) => <Tag color={statusColors[s]}>{statusLabels[s] || s}</Tag>,
    },
    {
      title: '结果',
      key: 'result',
      width: 90,
      render: (_: unknown, record: Inspection) => {
        const r = record.overall_result || record.result;
        if (!r) return <Tag>待检</Tag>;
        return <Tag color={resultColors[r]}>{resultLabels[r] || r}</Tag>;
      },
    },
    {
      title: '检验员',
      key: 'inspector',
      width: 100,
      render: (_: unknown, record: Inspection) => record.inspector || '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: unknown, record: Inspection) => (
        <Button
          type="link"
          size="small"
          onClick={() => { setCurrentInsp(record); setDrawerVisible(true); }}
        >
          详情
        </Button>
      ),
    },
  ];

  const detail = inspDetail || currentInsp;
  const canComplete = detail?.status !== 'completed';
  const inspections = data?.items || [];
  const detailItems = (inspDetail as Inspection)?.items || [];

  const handleCompleteSubmit = (values: any) => {
    // Build items array from edited state
    const items = detailItems.map(item => ({
      id: item.id,
      inspected_quantity: itemResults[item.id]?.inspected_quantity ?? item.inspected_quantity,
      qualified_quantity: itemResults[item.id]?.qualified_quantity ?? item.qualified_quantity,
      defect_quantity: itemResults[item.id]?.defect_quantity ?? item.defect_quantity,
      defect_description: itemResults[item.id]?.defect_description ?? item.defect_description ?? '',
      result: itemResults[item.id]?.result ?? item.result ?? values.result,
    }));

    completeMutation.mutate({
      result: values.result,
      notes: values.notes,
      items: items.length > 0 ? items : undefined,
    });
  };

  // Inspection items table columns (for the drawer)
  const inspItemColumns = [
    { title: '物料', dataIndex: 'material_name', key: 'material_name', width: 120, ellipsis: true },
    { title: '编码', dataIndex: 'material_code', key: 'material_code', width: 90 },
    {
      title: '检验数',
      dataIndex: 'inspected_quantity',
      key: 'inspected_quantity',
      width: 90,
      render: (v: number, record: InspectionItem) => canComplete ? (
        <InputNumber
          size="small"
          min={0}
          defaultValue={v}
          style={{ width: 70 }}
          onChange={(val) => setItemResults(prev => ({ ...prev, [record.id]: { ...prev[record.id], inspected_quantity: val || 0 } }))}
        />
      ) : v,
    },
    {
      title: '合格数',
      dataIndex: 'qualified_quantity',
      key: 'qualified_quantity',
      width: 90,
      render: (v: number, record: InspectionItem) => canComplete ? (
        <InputNumber
          size="small"
          min={0}
          defaultValue={v}
          style={{ width: 70 }}
          onChange={(val) => setItemResults(prev => ({ ...prev, [record.id]: { ...prev[record.id], qualified_quantity: val || 0 } }))}
        />
      ) : v,
    },
    {
      title: '不良数',
      dataIndex: 'defect_quantity',
      key: 'defect_quantity',
      width: 90,
      render: (v: number, record: InspectionItem) => canComplete ? (
        <InputNumber
          size="small"
          min={0}
          defaultValue={v}
          style={{ width: 70 }}
          onChange={(val) => setItemResults(prev => ({ ...prev, [record.id]: { ...prev[record.id], defect_quantity: val || 0 } }))}
        />
      ) : <span style={{ color: v > 0 ? '#cf1322' : undefined }}>{v}</span>,
    },
    {
      title: '结果',
      key: 'result',
      width: 90,
      render: (_: unknown, record: InspectionItem) => {
        if (canComplete) {
          return (
            <Select
              size="small"
              defaultValue={record.result || undefined}
              placeholder="结果"
              style={{ width: 80 }}
              onChange={(val) => setItemResults(prev => ({ ...prev, [record.id]: { ...prev[record.id], result: val } }))}
              options={[
                { value: 'passed', label: '合格' },
                { value: 'failed', label: '不合格' },
                { value: 'conditional', label: '让步' },
              ]}
            />
          );
        }
        const r = record.result;
        return r ? <Tag color={resultColors[r]}>{resultLabels[r] || r}</Tag> : '-';
      },
    },
  ];

  // Shared drawer component
  const drawerNode = (
    <Drawer
      title={detail?.inspection_code || '检验详情'}
      open={drawerVisible}
      onClose={() => { setDrawerVisible(false); setCurrentInsp(null); completeForm.resetFields(); setItemResults({}); }}
      width={isMobile ? '100%' : 760}
    >
      {detail && (
        <>
          <Descriptions column={isMobile ? 1 : 2} bordered size="small" style={{ marginBottom: 16 }}>
            <Descriptions.Item label="检验编码">{detail.inspection_code}</Descriptions.Item>
            <Descriptions.Item label="PO编码">{detail.po?.po_code || '-'}</Descriptions.Item>
            <Descriptions.Item label="供应商">{detail.supplier?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label="数量">{detail.quantity ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={statusColors[detail.status]}>{statusLabels[detail.status] || detail.status}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="结果">
              {(detail.overall_result || detail.result) ? (
                <Tag color={resultColors[detail.overall_result || detail.result]}>
                  {resultLabels[detail.overall_result || detail.result] || detail.overall_result || detail.result}
                </Tag>
              ) : '待检'}
            </Descriptions.Item>
            <Descriptions.Item label="检验员">{detail.inspector || '-'}</Descriptions.Item>
            <Descriptions.Item label="检验时间">
              {detail.inspection_date ? dayjs(detail.inspection_date).format('YYYY-MM-DD HH:mm') : (detail.inspected_at ? dayjs(detail.inspected_at).format('YYYY-MM-DD HH:mm') : '-')}
            </Descriptions.Item>
            <Descriptions.Item label="备注" span={isMobile ? 1 : 2}>{detail.notes || '-'}</Descriptions.Item>
          </Descriptions>

          {/* Inspection Items Table */}
          {detailItems.length > 0 && (
            <>
              <h4 style={{ marginBottom: 8 }}>质检行项 ({detailItems.length})</h4>
              <Table
                columns={inspItemColumns}
                dataSource={detailItems}
                rowKey="id"
                size="small"
                scroll={{ x: 600 }}
                pagination={false}
                style={{ marginBottom: 16 }}
              />
            </>
          )}

          {canComplete && (
            <>
              <h4>完成检验</h4>
              <Form
                form={completeForm}
                layout="vertical"
                onFinish={handleCompleteSubmit}
              >
                <Form.Item name="result" label="整体检验结果" rules={[{ required: true, message: '请选择结果' }]}>
                  <Select
                    placeholder="选择结果"
                    options={[
                      { value: 'passed', label: '合格' },
                      { value: 'failed', label: '不合格' },
                      { value: 'conditional', label: '条件放行' },
                    ]}
                  />
                </Form.Item>
                <Form.Item name="notes" label="备注">
                  <Input.TextArea rows={2} placeholder="检验备注" />
                </Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  icon={<CheckCircleOutlined />}
                  loading={completeMutation.isPending}
                >
                  完成检验
                </Button>
              </Form>
            </>
          )}
        </>
      )}
    </Drawer>
  );

  // ========== Mobile Layout ==========
  if (isMobile) {
    const statusFilterOptions = [
      { label: '全部', value: undefined as string | undefined },
      ...Object.entries(statusLabels).filter(([k]) => k !== 'in_progress').map(([k, v]) => ({ label: v, value: k as string | undefined })),
    ];
    return (
      <div style={{ background: '#f5f5f5', minHeight: '100vh' }}>
        <div style={{ padding: '12px 16px', background: '#fff', position: 'sticky', top: 0, zIndex: 10 }}>
          <Input
            placeholder="搜索检验编码/物料"
            prefix={<SearchOutlined style={{ color: '#bbb' }} />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onPressEnter={() => setPage(1)}
            allowClear
            style={{ borderRadius: 20 }}
          />
        </div>
        <div className="mobile-filter-pills" style={{ padding: '8px 12px' }}>
          {statusFilterOptions.map(opt => (
            <div
              key={opt.value || 'all'}
              className={`mobile-filter-pill ${filterStatus === opt.value ? 'active' : ''}`}
              onClick={() => { setFilterStatus(opt.value); setPage(1); }}
            >{opt.label}</div>
          ))}
        </div>
        <div style={{ padding: '0 12px' }}>
          {isLoading ? (
            <div style={{ textAlign: 'center', padding: 40 }}><Spin tip="加载中..." /></div>
          ) : inspections.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>暂无检验记录</div>
          ) : inspections.map((insp) => (
            <div
              key={insp.id}
              onClick={() => { setCurrentInsp(insp); setDrawerVisible(true); }}
              style={{ background: '#fff', borderRadius: 10, padding: '12px 14px', marginBottom: 8, boxShadow: '0 1px 3px rgba(0,0,0,0.04)', cursor: 'pointer' }}
            >
              <div style={{ display: 'flex', alignItems: 'center', marginBottom: 4 }}>
                <span style={{ fontFamily: 'monospace', color: '#1677ff', fontSize: 13, flex: 1 }}>{insp.inspection_code}</span>
                <Tag color={statusColors[insp.status]} style={{ margin: 0 }}>{statusLabels[insp.status] || insp.status}</Tag>
              </div>
              <div style={{ fontWeight: 600, fontSize: 15, marginBottom: 6, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {insp.material_name || '-'}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', fontSize: 13, color: '#666', gap: 8, flexWrap: 'wrap' }}>
                {insp.po?.po_code && <span style={{ fontFamily: 'monospace', fontSize: 12 }}>PO: {insp.po.po_code}</span>}
                {insp.supplier?.name && <span>{insp.supplier.name}</span>}
                {insp.quantity != null && <span>x{insp.quantity}</span>}
                {(insp.overall_result || insp.result) ? (
                  <Tag color={resultColors[insp.overall_result || insp.result]} style={{ margin: 0, fontSize: 11 }}>
                    {resultLabels[insp.overall_result || insp.result]}
                  </Tag>
                ) : null}
                <RightOutlined style={{ fontSize: 10, color: '#ccc', marginLeft: 'auto' }} />
              </div>
            </div>
          ))}
        </div>
        {drawerNode}
      </div>
    );
  }

  // ========== Desktop Layout ==========
  return (
    <div>
      <Card
        title="来料检验"
        extra={
          <Space wrap>
            <Select
              placeholder="状态"
              allowClear
              style={{ width: 110 }}
              options={Object.entries(statusLabels).filter(([k]) => k !== 'in_progress').map(([k, v]) => ({ value: k, label: v }))}
              value={filterStatus}
              onChange={(v) => { setFilterStatus(v); setPage(1); }}
            />
            <Select
              placeholder="结果"
              allowClear
              style={{ width: 110 }}
              options={[
                { value: 'passed', label: '合格' },
                { value: 'failed', label: '不合格' },
                { value: 'conditional', label: '条件放行' },
              ]}
              value={filterResult}
              onChange={(v) => { setFilterResult(v); setPage(1); }}
            />
            <Search
              placeholder="搜索"
              allowClear
              style={{ width: 180 }}
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              onSearch={() => setPage(1)}
            />
            <Button
              icon={<ReloadOutlined />}
              onClick={() => queryClient.invalidateQueries({ queryKey: ['srm-inspections'] })}
            >
              刷新
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={inspections}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1050 }}
          pagination={{
            current: page,
            pageSize,
            total: data?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); },
          }}
          onRow={(record) => ({
            onClick: () => { setCurrentInsp(record); setDrawerVisible(true); },
            style: { cursor: 'pointer' },
          })}
        />
      </Card>
      {drawerNode}
    </div>
  );
};

export default Inspections;
