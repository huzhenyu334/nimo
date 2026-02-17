import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Input, Tag, Badge, Progress, Empty, Spin, Typography } from 'antd';
import { SearchOutlined, RightOutlined } from '@ant-design/icons';
import { projectApi, Project } from '@/api/projects';
import { projectBomApi } from '@/api/projectBom';
import { useIsMobile } from '@/hooks/useIsMobile';

const { Text } = Typography;

const phaseLabels: Record<string, string> = {
  concept: '概念', evt: 'EVT', dvt: 'DVT', pvt: 'PVT', mp: 'MP',
  CONCEPT: '概念', EVT: 'EVT', DVT: 'DVT', PVT: 'PVT', MP: 'MP',
};

const phaseColors: Record<string, string> = {
  concept: 'purple', evt: 'blue', dvt: 'cyan', pvt: 'orange', mp: 'green',
  CONCEPT: 'purple', EVT: 'blue', DVT: 'DVT', PVT: 'orange', MP: 'green',
};

const statusLabels: Record<string, string> = {
  planning: '规划中', active: '进行中', completed: '已完成', on_hold: '暂停',
};

const statusColors: Record<string, string> = {
  planning: 'default', active: 'processing', completed: 'success', on_hold: 'warning',
};

const formatCurrency = (v: number): string =>
  `\u00a5${v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

const BOMManagement: React.FC = () => {
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const [search, setSearch] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['projects', 'bom-management'],
    queryFn: () => projectApi.list({ page: 1, page_size: 100 }),
  });

  const { data: costData } = useQuery({
    queryKey: ['bom-cost-summary'],
    queryFn: () => projectBomApi.bomCostSummary(),
    staleTime: 30_000,
  });

  const costMap = useMemo(() => {
    const map: Record<string, { total_cost: number; total_items: number; unpriced_items: number }> = {};
    for (const item of costData || []) {
      map[item.project_id] = item;
    }
    return map;
  }, [costData]);

  // Compute grand total across all projects
  const grandTotal = useMemo(() => {
    let cost = 0;
    let items = 0;
    let unpriced = 0;
    for (const item of costData || []) {
      cost += item.total_cost || 0;
      items += item.total_items || 0;
      unpriced += item.unpriced_items || 0;
    }
    return { cost, items, unpriced };
  }, [costData]);

  const projects: Project[] = (data as any)?.items || data || [];

  const filtered = projects.filter(p =>
    !search || p.name.toLowerCase().includes(search.toLowerCase()) ||
    p.code?.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div style={{ padding: isMobile ? 0 : 24 }}>
      {!isMobile && (
        <div style={{ marginBottom: 16 }}>
          <Text strong style={{ fontSize: 20 }}>BOM管理</Text>
          <Text type="secondary" style={{ marginLeft: 12, fontSize: 14 }}>选择项目进行BOM编辑</Text>
        </div>
      )}

      {/* Search */}
      <div style={{ padding: isMobile ? '8px 16px' : '0 0 16px 0' }}>
        <Input
          placeholder="搜索项目..."
          prefix={<SearchOutlined />}
          value={search}
          onChange={e => setSearch(e.target.value)}
          allowClear
          style={isMobile ? { borderRadius: 20, height: 40, background: '#f5f5f5', border: 'none' } : { width: 320 }}
        />
      </div>

      {/* Grand total cost summary */}
      {grandTotal.items > 0 && (
        <div style={{
          display: 'flex', gap: 16, padding: '8px 16px', background: '#f6ffed',
          borderRadius: 6, border: '1px solid #b7eb8f', marginBottom: 16, flexWrap: 'wrap', alignItems: 'center',
          ...(isMobile ? { margin: '0 12px 12px' } : {}),
        }}>
          <Text style={{ fontSize: 13 }}>
            全部项目BOM总成本: <Text strong style={{ color: '#1677ff', fontSize: 15 }}>{formatCurrency(grandTotal.cost)}</Text>
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>共 {grandTotal.items} 项物料</Text>
          {grandTotal.unpriced > 0 && (
            <Text type="warning" style={{ fontSize: 12 }}>{grandTotal.unpriced}项未定价</Text>
          )}
        </div>
      )}

      {/* Loading */}
      {isLoading && (
        <div style={{ textAlign: 'center', padding: 60 }}><Spin /></div>
      )}

      {/* Empty */}
      {!isLoading && filtered.length === 0 && (
        <Empty description="暂无项目" style={{ padding: 60 }} />
      )}

      {/* Project list */}
      <div style={{ padding: isMobile ? '0 12px' : 0, display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(auto-fill, minmax(340px, 1fr))', gap: isMobile ? 8 : 16 }}>
        {filtered.map(project => {
          const cost = costMap[project.id];
          return (
            <div
              key={project.id}
              onClick={() => navigate(`/bom-management/${project.id}`)}
              style={{
                background: '#fff',
                borderRadius: 12,
                padding: isMobile ? '14px 16px' : '16px 20px',
                cursor: 'pointer',
                border: isMobile ? 'none' : '1px solid #f0f0f0',
                boxShadow: isMobile ? 'none' : '0 1px 2px rgba(0,0,0,0.04)',
                borderBottom: isMobile ? '1px solid #f0f0f0' : undefined,
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                transition: 'background 0.15s',
              }}
            >
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                  <Text strong style={{ fontSize: 15, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {project.name}
                  </Text>
                  {project.code && (
                    <Text type="secondary" style={{ fontSize: 12, flexShrink: 0 }}>{project.code}</Text>
                  )}
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Tag color={phaseColors[project.phase]}>{phaseLabels[project.phase] || project.phase}</Tag>
                  <Badge status={statusColors[project.status] as any} text={statusLabels[project.status] || project.status} />
                </div>
                {cost && cost.total_cost > 0 && (
                  <div style={{ marginTop: 4, fontSize: 12 }}>
                    <Text style={{ color: '#1677ff' }}>{formatCurrency(cost.total_cost)}</Text>
                    <Text type="secondary" style={{ marginLeft: 8 }}>{cost.total_items}项物料</Text>
                    {cost.unpriced_items > 0 && (
                      <Text type="warning" style={{ marginLeft: 6 }}>{cost.unpriced_items}项未定价</Text>
                    )}
                  </div>
                )}
              </div>
              <Progress type="circle" percent={project.progress} size={36} />
              <RightOutlined style={{ color: '#ccc', fontSize: 12 }} />
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default BOMManagement;
