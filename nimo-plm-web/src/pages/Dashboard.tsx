import React from 'react';
import { Card, Row, Col, Statistic, Typography, List, Tag, Progress, Space, Button, Empty, Skeleton, theme } from 'antd';
import {
  ProjectOutlined,
  CheckCircleOutlined,
  RightOutlined,
  ExperimentOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/contexts/AuthContext';
import { useSSE } from '@/hooks/useSSE';
import { projectApi, Project } from '@/api/projects';
import { materialsApi } from '@/api/materials';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';

dayjs.locale('zh-cn');

const { Title, Text } = Typography;

const phaseColors: Record<string, string> = {
  concept: 'purple',
  evt: 'blue',
  dvt: 'cyan',
  pvt: 'orange',
  mp: 'green',
};

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { token } = theme.useToken();

  // SSE: 实时推送自动刷新
  useSSE({
    onTaskUpdate: React.useCallback(() => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    }, [queryClient]),
    onProjectUpdate: React.useCallback(() => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    }, [queryClient]),
  });

  // Fetch real project data
  const { data: projectsData, isLoading: projectsLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectApi.list({ page_size: 999 }),
  });

  // Fetch materials count
  const { data: materialsData } = useQuery({
    queryKey: ['materials-count'],
    queryFn: () => materialsApi.list(),
  });

  const projects = projectsData?.items || [];
  const activeProjects = projects.filter(p => p.status === 'active' || p.status === 'planning');
  const completedProjects = projects.filter(p => p.status === 'completed');
  const materialsCount = materialsData?.total ?? '-';

  // Compute stats from real data
  const stats = {
    totalProjects: projects.length,
    activeProjects: activeProjects.length,
    completedProjects: completedProjects.length,
  };

  return (
    <div style={{ padding: token.paddingLG }}>
      {/* Welcome */}
      <div style={{ marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>
          🏠 工作台
        </Title>
        <Text type="secondary">
          👋 欢迎回来，{user?.name}！今天是 {dayjs().format('YYYY年M月D日 dddd')}
        </Text>
      </div>

      {/* Stats cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable onClick={() => navigate('/projects')}>
            <Statistic
              title="进行中项目"
              value={stats.activeProjects}
              prefix={<ProjectOutlined style={{ color: '#1890ff' }} />}
              suffix={<Text type="secondary" style={{ fontSize: 14 }}>/ {stats.totalProjects}</Text>}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable onClick={() => navigate('/projects?status=completed')}>
            <Statistic
              title="已完成项目"
              value={stats.completedProjects}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable onClick={() => navigate('/materials')}>
            <Statistic
              title="物料选型库"
              value={materialsCount}
              prefix={<ExperimentOutlined style={{ color: '#722ed1' }} />}
            />
          </Card>
        </Col>
      </Row>

      {/* Main content */}
      <Row gutter={[16, 16]}>
        {/* 我参与的项目 */}
        <Col xs={24}>
          <Card
            title={
              <Space>
                <ProjectOutlined />
                <span>我参与的项目</span>
              </Space>
            }
            extra={
              <Button type="link" size="small" onClick={() => navigate('/projects')}>
                查看全部 <RightOutlined />
              </Button>
            }
          >
            {projectsLoading ? (
              <Skeleton active paragraph={{ rows: 4 }} />
            ) : projects.length === 0 ? (
              <Empty description="暂无项目" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              <List
                dataSource={projects.slice(0, 6)}
                renderItem={(project: Project) => (
                  <List.Item
                    style={{ cursor: 'pointer', padding: '10px 0' }}
                    onClick={() => navigate(`/projects/${project.id}`)}
                    actions={[<RightOutlined key="go" style={{ color: '#ccc' }} />]}
                  >
                    <div style={{ width: '100%', paddingRight: 8 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                        <Space>
                          <Text strong>{project.name}</Text>
                          <Text code style={{ fontSize: 11 }}>{project.code}</Text>
                        </Space>
                        <Tag color={phaseColors[project.phase?.toLowerCase()] || 'default'}>
                          {project.phase?.toUpperCase()}
                        </Tag>
                      </div>
                      <Progress
                        percent={project.progress}
                        size="small"
                        strokeColor={
                          project.progress >= 80 ? '#52c41a' :
                          project.progress >= 50 ? '#1890ff' :
                          project.progress >= 20 ? '#faad14' : '#ff4d4f'
                        }
                      />
                    </div>
                  </List.Item>
                )}
              />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
