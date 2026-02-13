import React from 'react';
import { Card, Row, Col, Statistic, Progress, Tag, Empty, Timeline, Button } from 'antd';
import {
  ProjectOutlined,
  AppstoreOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { srmApi, PurchaseRequest, PRItem, SRMProject, ActivityLog } from '@/api/srm';
import { projectApi } from '@/api/projects';
import dayjs from 'dayjs';

const phaseColors: Record<string, string> = {
  concept: 'purple', evt: 'blue', dvt: 'cyan', pvt: 'orange', mp: 'green',
};

// Derived project card from PR data
interface DerivedProject {
  id: string; // PLM project_id
  name: string;
  code: string;
  phase: string;
  totalItems: number;
  orderedCount: number;
  receivedCount: number;
  passedCount: number;
  srmProjectId?: string; // if linked to SRM project, use for kanban navigation
}

const SRMDashboard: React.FC = () => {
  const navigate = useNavigate();

  // Load all PRs (includes items)
  const { data: prData } = useQuery({
    queryKey: ['srm-prs-dashboard-stats'],
    queryFn: () => srmApi.listPRs({ page_size: 200 }),
  });

  // Load PLM projects for name/phase lookup
  const { data: plmProjectData } = useQuery({
    queryKey: ['plm-projects-dashboard'],
    queryFn: () => projectApi.list({ page_size: 100 }),
  });

  // Load SRM projects for kanban link
  const { data: srmProjectData } = useQuery({
    queryKey: ['srm-projects-dashboard'],
    queryFn: () => srmApi.listProjects({ page_size: 100 }),
  });

  // Build project map from PLM projects
  const plmProjectMap = React.useMemo(() => {
    const map: Record<string, { name: string; code: string; phase: string }> = {};
    (plmProjectData?.items || []).forEach((p: any) => {
      map[p.id] = { name: p.name, code: p.code, phase: p.phase || '' };
    });
    return map;
  }, [plmProjectData]);

  // Build SRM project map (plm_project_id → srm_project_id)
  const srmProjectMap = React.useMemo(() => {
    const map: Record<string, string> = {};
    (srmProjectData?.items || []).forEach((p: SRMProject) => {
      if (p.plm_project_id) {
        map[p.plm_project_id] = p.id;
      }
    });
    return map;
  }, [srmProjectData]);

  // Derive project list from PRs grouped by project_id
  const derivedProjects: DerivedProject[] = React.useMemo(() => {
    const prs = prData?.items || [];
    const groups: Record<string, PRItem[]> = {};

    prs.forEach((pr: PurchaseRequest) => {
      const pid = pr.project_id;
      if (!pid) return;
      if (!groups[pid]) groups[pid] = [];
      (pr.items || []).forEach((item) => groups[pid].push(item));
    });

    return Object.entries(groups).map(([projectId, items]) => {
      const plmInfo = plmProjectMap[projectId];
      let ordered = 0, received = 0, passed = 0;
      items.forEach((item) => {
        if (['ordered', 'shipped', 'received', 'inspecting', 'passed'].includes(item.status)) ordered++;
        if (['received', 'inspecting', 'passed'].includes(item.status)) received++;
        if (item.status === 'passed') passed++;
      });
      return {
        id: projectId,
        name: plmInfo?.name || projectId.slice(0, 8),
        code: plmInfo?.code || '',
        phase: plmInfo?.phase || '',
        totalItems: items.length,
        orderedCount: ordered,
        receivedCount: received,
        passedCount: passed,
        srmProjectId: srmProjectMap[projectId],
      };
    }).sort((a, b) => b.totalItems - a.totalItems);
  }, [prData, plmProjectMap, srmProjectMap]);

  // Compute stats
  const stats = React.useMemo(() => {
    let totalItems = 0;
    let overdueItems = 0;
    let passedItems = 0;

    derivedProjects.forEach((p) => {
      totalItems += p.totalItems;
      passedItems += p.passedCount;
    });

    const prs = prData?.items || [];
    const now = dayjs();
    prs.forEach((pr: PurchaseRequest) => {
      (pr.items || []).forEach((item) => {
        if (item.expected_date && item.status !== 'passed' && item.status !== 'failed') {
          if (dayjs(item.expected_date).isBefore(now, 'day')) {
            overdueItems++;
          }
        }
      });
    });

    const passRate = totalItems > 0 ? Math.round((passedItems / totalItems) * 100) : 0;
    return { projectCount: derivedProjects.length, totalItems, overdueItems, passRate };
  }, [derivedProjects, prData]);

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>采购总览</h2>

      {/* Stat cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="采购项目"
              value={stats.projectCount}
              prefix={<ProjectOutlined style={{ color: '#1890ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="采购物料"
              value={stats.totalItems}
              prefix={<AppstoreOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="超期物料"
              value={stats.overdueItems}
              valueStyle={stats.overdueItems > 0 ? { color: '#ff4d4f' } : undefined}
              prefix={<WarningOutlined style={{ color: stats.overdueItems > 0 ? '#ff4d4f' : '#999' }} />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="检验通过率"
              value={stats.passRate}
              suffix="%"
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
      </Row>

      {/* Project list derived from PRs */}
      <Card
        title="采购项目"
        extra={
          <Button type="link" onClick={() => navigate('/srm/projects')}>
            查看全部
          </Button>
        }
        style={{ marginBottom: 24 }}
      >
        {derivedProjects.length === 0 ? (
          <Empty description="暂无采购项目" />
        ) : (
          <Row gutter={[16, 16]}>
            {derivedProjects.map((project) => {
              const pct = project.totalItems > 0 ? Math.round((project.passedCount / project.totalItems) * 100) : 0;
              return (
                <Col xs={24} sm={12} lg={8} key={project.id}>
                  <Card
                    size="small"
                    hoverable
                    onClick={() => {
                      navigate(`/srm/kanban?project=${project.srmProjectId || project.id}`);
                    }}
                    style={{ cursor: 'pointer' }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                      <strong style={{ fontSize: 14 }}>{project.name}</strong>
                      {project.phase && (
                        <Tag color={phaseColors[project.phase] || 'default'}>
                          {project.phase.toUpperCase()}
                        </Tag>
                      )}
                    </div>
                    {project.code && (
                      <div style={{ fontSize: 12, color: '#999', marginBottom: 8 }}>{project.code}</div>
                    )}
                    <Progress
                      percent={pct}
                      size="small"
                      strokeColor="#52c41a"
                    />
                    <Row gutter={8} style={{ marginTop: 8 }}>
                      <Col span={6}>
                        <Statistic title="总计" value={project.totalItems} valueStyle={{ fontSize: 14 }} />
                      </Col>
                      <Col span={6}>
                        <Statistic title="已下单" value={project.orderedCount} valueStyle={{ fontSize: 14, color: '#1890ff' }} />
                      </Col>
                      <Col span={6}>
                        <Statistic title="已收货" value={project.receivedCount} valueStyle={{ fontSize: 14 }} />
                      </Col>
                      <Col span={6}>
                        <Statistic title="已通过" value={project.passedCount} valueStyle={{ fontSize: 14, color: '#52c41a' }} />
                      </Col>
                    </Row>
                  </Card>
                </Col>
              );
            })}
          </Row>
        )}
      </Card>

      {/* Recent activity */}
      <RecentActivity />
    </div>
  );
};

const RecentActivity: React.FC = () => {
  // Load SRM projects to find one for activity logs
  const { data: srmProjectData } = useQuery({
    queryKey: ['srm-projects-for-activity'],
    queryFn: () => srmApi.listProjects({ page_size: 10 }),
  });

  const firstProjectId = (srmProjectData?.items || [])[0]?.id;

  const { data: activityData } = useQuery({
    queryKey: ['srm-recent-activities', firstProjectId],
    queryFn: () => srmApi.listProjectActivities(firstProjectId!, { page_size: 15 }),
    enabled: !!firstProjectId,
  });

  const activities = activityData?.items || [];

  return (
    <Card title="最近操作记录">
      {activities.length === 0 ? (
        <Empty description="暂无操作记录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      ) : (
        <Timeline
          items={activities.map((log: ActivityLog) => ({
            children: (
              <div>
                <div style={{ fontSize: 13 }}>
                  {log.entity_code && (
                    <Tag style={{ fontSize: 11, marginRight: 6 }}>{log.entity_code}</Tag>
                  )}
                  {log.content}
                </div>
                <div style={{ fontSize: 12, color: '#999' }}>
                  {log.operator_name} · {dayjs(log.created_at).format('MM-DD HH:mm')}
                </div>
              </div>
            ),
          }))}
        />
      )}
    </Card>
  );
};

export default SRMDashboard;
