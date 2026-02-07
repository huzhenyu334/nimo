import React, { useState, useMemo, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Card,
  Tabs,
  Tag,
  Typography,
  Space,
  Button,
  Spin,
  Descriptions,
  Progress,
  Table,
  Modal,
  Form,
  Input,
  Select,
  Badge,
  message,
  Tooltip,
  Empty,
  Alert,
} from 'antd';
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  PlayCircleOutlined,
  RightOutlined,
  DownOutlined,
  PlusOutlined,
  ExportOutlined,
  EyeOutlined,
  UploadOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { projectApi, Project, Task } from '@/api/projects';
import { projectBomApi, ProjectBOM, CreateProjectBOMRequest } from '@/api/projectBom';
import { deliverablesApi } from '@/api/deliverables';
import { ecnApi, ECN } from '@/api/ecn';
import { documentsApi, Document } from '@/api/documents';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;

// ============ Constants ============

const PHASES = ['concept', 'evt', 'dvt', 'pvt', 'mp'];

const phaseColors: Record<string, string> = {
  concept: 'purple',
  evt: 'blue',
  dvt: 'cyan',
  pvt: 'orange',
  mp: 'green',
  CONCEPT: 'purple',
  EVT: 'blue',
  DVT: 'cyan',
  PVT: 'orange',
  MP: 'green',
};

const phaseLabels: Record<string, string> = {
  concept: '概念阶段',
  evt: 'EVT 工程验证',
  dvt: 'DVT 设计验证',
  pvt: 'PVT 生产验证',
  mp: 'MP 量产',
};

const statusColors: Record<string, string> = {
  planning: 'default',
  active: 'processing',
  on_hold: 'warning',
  completed: 'success',
  cancelled: 'error',
};

const taskStatusConfig: Record<string, { color: string; text: string; icon: React.ReactNode; barColor: string }> = {
  pending: { color: 'default', text: '待开始', icon: <ClockCircleOutlined />, barColor: '#bfbfbf' },
  ready: { color: 'blue', text: '就绪', icon: <PlayCircleOutlined />, barColor: '#69b1ff' },
  in_progress: { color: 'processing', text: '进行中', icon: <ClockCircleOutlined />, barColor: '#1677ff' },
  completed: { color: 'success', text: '已完成', icon: <CheckCircleOutlined />, barColor: '#52c41a' },
  blocked: { color: 'error', text: '阻塞', icon: <ExclamationCircleOutlined />, barColor: '#ff4d4f' },
  needs_review: { color: 'warning', text: '待审批', icon: <ExclamationCircleOutlined />, barColor: '#faad14' },
};

const GANTT_ROW_HEIGHT = 36;
const GANTT_HEADER_HEIGHT = 50;
const DAY_WIDTH = 28;
const LEFT_PANEL_WIDTH = 480;

// ============ Phase Progress Bar ============

const PhaseProgressBar: React.FC<{ currentPhase: string }> = ({ currentPhase }) => {
  const currentIndex = PHASES.indexOf(currentPhase?.toLowerCase());

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
      {PHASES.map((phase, index) => {
        let icon = '⬜';
        let fontWeight: number = 400;
        if (index < currentIndex) {
          icon = '✅';
        } else if (index === currentIndex) {
          icon = '🔵';
          fontWeight = 600;
        }

        return (
          <React.Fragment key={phase}>
            {index > 0 && (
              <span style={{ color: index <= currentIndex ? '#1890ff' : '#d9d9d9', fontSize: 12 }}>──▶</span>
            )}
            <span style={{
              fontWeight,
              fontSize: 13,
              color: index <= currentIndex ? '#333' : '#999',
            }}>
              {icon} {phase.toUpperCase()}
            </span>
          </React.Fragment>
        );
      })}
    </div>
  );
};

// ============ Gantt Helper Types ============

interface TreeTask extends Task {
  children: TreeTask[];
  depth: number;
  expanded?: boolean;
}

// ============ Gantt Chart Component ============

const GanttChart: React.FC<{
  tasks: Task[];
  projectId: string;
  onCompleteTask: (taskId: string) => void;
  completingTask: boolean;
}> = ({ tasks, projectId: _projectId, onCompleteTask: _onCompleteTask, completingTask: _completingTask }) => {
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [collapsedTasks, setCollapsedTasks] = useState<Set<string>>(new Set());
  const [groupBy, setGroupBy] = useState<'phase' | 'none'>('phase');
  const timelineRef = useRef<HTMLDivElement>(null);
  const leftPanelRef = useRef<HTMLDivElement>(null);

  const handleTimelineScroll = (e: React.UIEvent<HTMLDivElement>) => {
    if (leftPanelRef.current) {
      leftPanelRef.current.scrollTop = e.currentTarget.scrollTop;
    }
  };
  const handleLeftScroll = (e: React.UIEvent<HTMLDivElement>) => {
    if (timelineRef.current) {
      timelineRef.current.scrollTop = e.currentTarget.scrollTop;
    }
  };

  const buildTree = (tasks: Task[]): TreeTask[] => {
    const map = new Map<string, TreeTask>();
    const roots: TreeTask[] = [];
    tasks.forEach(t => map.set(t.id, { ...t, children: [], depth: 0 }));
    map.forEach(node => {
      if (node.parent_task_id && map.has(node.parent_task_id)) {
        const parent = map.get(node.parent_task_id)!;
        node.depth = parent.depth + 1;
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    });
    return roots;
  };

  const flattenTree = (nodes: TreeTask[]): TreeTask[] => {
    const result: TreeTask[] = [];
    const walk = (items: TreeTask[], depth: number) => {
      items.forEach(item => {
        item.depth = depth;
        result.push(item);
        if (item.children.length > 0 && !collapsedTasks.has(item.id)) {
          walk(item.children, depth + 1);
        }
      });
    };
    walk(nodes, 0);
    return result;
  };

  const groupedData = useMemo(() => {
    if (groupBy === 'none') {
      const tree = buildTree(tasks);
      return [{ phase: '', label: '全部任务', tasks: flattenTree(tree) }];
    }
    const phaseOrder = ['concept', 'evt', 'dvt', 'pvt', 'mp', ''];
    const groups = new Map<string, Task[]>();
    tasks.forEach(t => {
      const phase = (typeof t.phase === 'object' && t.phase !== null ? (t.phase as any).phase : (t.phase || '')).toLowerCase();
      if (!groups.has(phase)) groups.set(phase, []);
      groups.get(phase)!.push(t);
    });
    return phaseOrder
      .filter(p => groups.has(p))
      .map(phase => {
        const tree = buildTree(groups.get(phase)!);
        return {
          phase,
          label: phaseLabels[phase] || (phase ? phase.toUpperCase() : '未分类'),
          tasks: flattenTree(tree),
        };
      });
  }, [tasks, groupBy, collapsedTasks]);

  const { startDate, endDate, totalDays } = useMemo(() => {
    let min = dayjs().subtract(7, 'day');
    let max = dayjs().add(30, 'day');
    tasks.forEach(t => {
      if (t.start_date) { const d = dayjs(t.start_date); if (d.isBefore(min)) min = d; }
      if (t.due_date) { const d = dayjs(t.due_date); if (d.isAfter(max)) max = d; }
    });
    min = min.subtract(7, 'day').startOf('week');
    max = max.add(14, 'day').endOf('week');
    return { startDate: min, endDate: max, totalDays: max.diff(min, 'day') + 1 };
  }, [tasks]);

  const monthHeaders = useMemo(() => {
    const months: { label: string; days: number; offset: number }[] = [];
    let cursor = startDate;
    while (cursor.isBefore(endDate)) {
      const monthEnd = cursor.endOf('month');
      const end = monthEnd.isAfter(endDate) ? endDate : monthEnd;
      const days = end.diff(cursor, 'day') + 1;
      months.push({ label: cursor.format('YYYY年M月'), days, offset: cursor.diff(startDate, 'day') });
      cursor = monthEnd.add(1, 'day');
    }
    return months;
  }, [startDate, endDate]);

  const dayHeaders = useMemo(() => {
    const days: { label: string; date: dayjs.Dayjs; isWeekend: boolean; isToday: boolean }[] = [];
    for (let i = 0; i < totalDays; i++) {
      const d = startDate.add(i, 'day');
      days.push({ label: d.format('D'), date: d, isWeekend: d.day() === 0 || d.day() === 6, isToday: d.isSame(dayjs(), 'day') });
    }
    return days;
  }, [startDate, totalDays]);

  useEffect(() => {
    if (timelineRef.current) {
      const todayOffset = dayjs().diff(startDate, 'day');
      const scrollTo = Math.max(0, todayOffset * DAY_WIDTH - 200);
      timelineRef.current.scrollLeft = scrollTo;
    }
  }, [startDate]);

  const getBar = (task: Task) => {
    const start = task.start_date ? dayjs(task.start_date) : null;
    const end = task.due_date ? dayjs(task.due_date) : null;
    if (!start && !end) return null;
    const barStart = start || end!;
    const barEnd = end || start!;
    const left = barStart.diff(startDate, 'day') * DAY_WIDTH;
    const width = Math.max((barEnd.diff(barStart, 'day') + 1) * DAY_WIDTH, DAY_WIDTH);
    return { left, width };
  };

  const toggleGroup = (phase: string) => {
    setCollapsedGroups(prev => { const next = new Set(prev); if (next.has(phase)) next.delete(phase); else next.add(phase); return next; });
  };
  const toggleTask = (taskId: string) => {
    setCollapsedTasks(prev => { const next = new Set(prev); if (next.has(taskId)) next.delete(taskId); else next.add(taskId); return next; });
  };

  const rows: Array<{ type: 'group'; phase: string; label: string; count: number } | { type: 'task'; task: TreeTask }> = [];
  groupedData.forEach(group => {
    if (groupBy === 'phase') rows.push({ type: 'group', phase: group.phase, label: group.label, count: group.tasks.length });
    if (!collapsedGroups.has(group.phase) || groupBy === 'none') {
      group.tasks.forEach(t => rows.push({ type: 'task', task: t }));
    }
  });
  const totalHeight = rows.length * GANTT_ROW_HEIGHT;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
        <Space>
          <Text strong>甘特图视图</Text>
          <Tag>{tasks.length} 个任务</Tag>
        </Space>
        <Space>
          <Text type="secondary">分组:</Text>
          <Select size="small" value={groupBy} onChange={setGroupBy} style={{ width: 120 }}
            options={[{ label: '按阶段', value: 'phase' }, { label: '不分组', value: 'none' }]} />
        </Space>
      </div>

      <div style={{ display: 'flex', flex: 1, overflow: 'hidden', border: '1px solid #e8e8e8', borderRadius: 4 }}>
        {/* Left panel */}
        <div style={{ width: LEFT_PANEL_WIDTH, flexShrink: 0, borderRight: '2px solid #d9d9d9', display: 'flex', flexDirection: 'column' }}>
          <div style={{ height: GANTT_HEADER_HEIGHT, borderBottom: '1px solid #e8e8e8', display: 'flex', alignItems: 'center', padding: '0 12px', background: '#fafafa', fontWeight: 600, fontSize: 13, flexShrink: 0 }}>
            <span style={{ flex: 1 }}>任务名称</span>
            <span style={{ width: 60, textAlign: 'center' }}>负责人</span>
            <span style={{ width: 50, textAlign: 'center' }}>状态</span>
            <span style={{ width: 45, textAlign: 'center' }}>进度</span>
          </div>
          <div ref={leftPanelRef} onScroll={handleLeftScroll} style={{ flex: 1, overflowY: 'auto', overflowX: 'hidden' }}>
            <div style={{ minHeight: totalHeight }}>
              {rows.map((row, idx) => {
                if (row.type === 'group') {
                  const collapsed = collapsedGroups.has(row.phase);
                  return (
                    <div key={`group-${row.phase}`} style={{ height: GANTT_ROW_HEIGHT, display: 'flex', alignItems: 'center', padding: '0 12px', background: '#f5f5f5', cursor: 'pointer', borderBottom: '1px solid #f0f0f0', fontWeight: 600, fontSize: 13 }} onClick={() => toggleGroup(row.phase)}>
                      {collapsed ? <RightOutlined style={{ fontSize: 10, marginRight: 8 }} /> : <DownOutlined style={{ fontSize: 10, marginRight: 8 }} />}
                      <Tag color={phaseColors[row.phase] || 'default'} style={{ marginRight: 8 }}>{row.label}</Tag>
                      <Text type="secondary" style={{ fontSize: 12 }}>({row.count})</Text>
                    </div>
                  );
                }
                const task = row.task;
                const config = taskStatusConfig[task.status] || taskStatusConfig.pending;
                const hasChildren = task.children.length > 0;
                const isCollapsed = collapsedTasks.has(task.id);
                const isMilestone = task.task_type === 'MILESTONE';
                return (
                  <div key={task.id} style={{ height: GANTT_ROW_HEIGHT, display: 'flex', alignItems: 'center', padding: '0 12px', borderBottom: '1px solid #f7f7f7', fontSize: 12, background: idx % 2 === 0 ? '#fff' : '#fafcff' }}>
                    <div style={{ flex: 1, display: 'flex', alignItems: 'center', minWidth: 0, paddingLeft: task.depth * 20 }}>
                      {hasChildren ? (
                        <span style={{ cursor: 'pointer', marginRight: 4, width: 16, textAlign: 'center', flexShrink: 0 }} onClick={() => toggleTask(task.id)}>
                          {isCollapsed ? <RightOutlined style={{ fontSize: 9 }} /> : <DownOutlined style={{ fontSize: 9 }} />}
                        </span>
                      ) : <span style={{ width: 16, marginRight: 4, flexShrink: 0 }} />}
                      {isMilestone && <span style={{ display: 'inline-block', width: 10, height: 10, background: config.barColor, transform: 'rotate(45deg)', marginRight: 6, flexShrink: 0 }} />}
                      <Tooltip title={task.title}>
                        <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontWeight: isMilestone ? 600 : (task.task_type === 'SUBTASK' ? 400 : 500), color: task.is_critical ? '#cf1322' : undefined }}>
                          {(task.code || task.task_code) ? <Text code style={{ fontSize: 11, marginRight: 4 }}>{task.code || task.task_code}</Text> : null}
                          {task.title}
                        </span>
                      </Tooltip>
                    </div>
                    <span style={{ width: 60, textAlign: 'center', flexShrink: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: '#666' }}>
                      {(task.assignee?.name || task.assignee_name) || '-'}
                    </span>
                    <span style={{ width: 50, textAlign: 'center', flexShrink: 0 }}>
                      <Tag color={config.color} style={{ fontSize: 10, padding: '0 4px', margin: 0, lineHeight: '18px' }}>{config.text}</Tag>
                    </span>
                    <span style={{ width: 45, textAlign: 'center', flexShrink: 0, fontSize: 11, color: '#666' }}>{task.progress}%</span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Right timeline */}
        <div ref={timelineRef} onScroll={handleTimelineScroll} style={{ flex: 1, overflow: 'auto' }}>
          <div style={{ minWidth: totalDays * DAY_WIDTH, position: 'relative' }}>
            <div style={{ position: 'sticky', top: 0, zIndex: 10, background: '#fafafa' }}>
              <div style={{ display: 'flex', height: 24, borderBottom: '1px solid #e8e8e8' }}>
                {monthHeaders.map((m, i) => (
                  <div key={i} style={{ width: m.days * DAY_WIDTH, textAlign: 'center', fontSize: 11, fontWeight: 600, lineHeight: '24px', borderRight: '1px solid #e8e8e8', color: '#333' }}>{m.label}</div>
                ))}
              </div>
              <div style={{ display: 'flex', height: GANTT_HEADER_HEIGHT - 24, borderBottom: '1px solid #e8e8e8' }}>
                {dayHeaders.map((d, i) => (
                  <div key={i} style={{ width: DAY_WIDTH, textAlign: 'center', fontSize: 10, lineHeight: `${GANTT_HEADER_HEIGHT - 24}px`, color: d.isToday ? '#fff' : d.isWeekend ? '#bbb' : '#666', background: d.isToday ? '#1677ff' : d.isWeekend ? '#f9f9f9' : 'transparent', borderRight: '1px solid #f0f0f0', fontWeight: d.isToday ? 700 : 400 }}>{d.label}</div>
                ))}
              </div>
            </div>
            <div style={{ position: 'relative' }}>
              {dayHeaders.map((d, i) => d.isWeekend && (
                <div key={`bg-${i}`} style={{ position: 'absolute', left: i * DAY_WIDTH, top: 0, width: DAY_WIDTH, height: totalHeight, background: 'rgba(0,0,0,0.02)', zIndex: 0 }} />
              ))}
              {(() => {
                const todayOffset = dayjs().diff(startDate, 'day');
                if (todayOffset >= 0 && todayOffset <= totalDays) {
                  return <div style={{ position: 'absolute', left: todayOffset * DAY_WIDTH + DAY_WIDTH / 2, top: 0, width: 2, height: totalHeight, background: '#ff4d4f', zIndex: 5, opacity: 0.6 }} />;
                }
                return null;
              })()}
              {rows.map((row, idx) => {
                if (row.type === 'group') {
                  return <div key={`gbar-${row.phase}`} style={{ height: GANTT_ROW_HEIGHT, background: '#f5f5f5', borderBottom: '1px solid #f0f0f0' }} />;
                }
                const task = row.task;
                const bar = getBar(task);
                const config = taskStatusConfig[task.status] || taskStatusConfig.pending;
                const isMilestone = task.task_type === 'MILESTONE';
                return (
                  <div key={task.id} style={{ height: GANTT_ROW_HEIGHT, position: 'relative', borderBottom: '1px solid #f7f7f7', background: idx % 2 === 0 ? '#fff' : '#fafcff' }}>
                    {bar && !isMilestone && (
                      <Tooltip title={<div><div><strong>{task.title}</strong></div><div>{task.start_date || '?'} → {task.due_date || '?'}</div><div>进度: {task.progress}%</div>{(task.assignee?.name || task.assignee_name) && <div>负责人: {task.assignee?.name || task.assignee_name}</div>}</div>}>
                        <div style={{ position: 'absolute', left: bar.left, top: (GANTT_ROW_HEIGHT - 18) / 2, width: bar.width, height: 18, borderRadius: 3, background: config.barColor, opacity: 0.85, zIndex: 2, cursor: 'pointer', overflow: 'hidden', transition: 'opacity 0.2s' }}
                          onMouseEnter={e => (e.currentTarget.style.opacity = '1')} onMouseLeave={e => (e.currentTarget.style.opacity = '0.85')}>
                          {task.progress > 0 && task.progress < 100 && (
                            <div style={{ position: 'absolute', left: 0, top: 0, width: `${task.progress}%`, height: '100%', background: 'rgba(255,255,255,0.3)', borderRadius: '3px 0 0 3px' }} />
                          )}
                          {bar.width > 80 && (
                            <span style={{ position: 'absolute', left: 6, top: 0, lineHeight: '18px', fontSize: 10, color: '#fff', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: bar.width - 12 }}>{task.title}</span>
                          )}
                        </div>
                      </Tooltip>
                    )}
                    {bar && isMilestone && (
                      <Tooltip title={<div><div><strong>🔷 里程碑: {task.title}</strong></div><div>{task.due_date || task.start_date || '-'}</div>{(task.assignee?.name || task.assignee_name) && <div>负责人: {task.assignee?.name || task.assignee_name}</div>}</div>}>
                        <div style={{ position: 'absolute', left: bar.left + (bar.width / 2) - 8, top: (GANTT_ROW_HEIGHT - 16) / 2, width: 16, height: 16, background: config.barColor, transform: 'rotate(45deg)', zIndex: 2, cursor: 'pointer', border: '2px solid rgba(255,255,255,0.8)', boxShadow: '0 1px 3px rgba(0,0,0,0.2)' }} />
                      </Tooltip>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', gap: 16, padding: '8px 0', flexWrap: 'wrap', alignItems: 'center', borderTop: '1px solid #f0f0f0' }}>
        <Text type="secondary" style={{ fontSize: 12 }}>图例:</Text>
        {Object.entries(taskStatusConfig).map(([key, val]) => (
          <Space key={key} size={4}>
            <span style={{ display: 'inline-block', width: 14, height: 10, background: val.barColor, borderRadius: 2 }} />
            <Text style={{ fontSize: 11 }}>{val.text}</Text>
          </Space>
        ))}
        <Space size={4}>
          <span style={{ display: 'inline-block', width: 10, height: 10, background: '#1677ff', transform: 'rotate(45deg)' }} />
          <Text style={{ fontSize: 11 }}>里程碑</Text>
        </Space>
        <Space size={4}>
          <span style={{ display: 'inline-block', width: 2, height: 12, background: '#ff4d4f' }} />
          <Text style={{ fontSize: 11 }}>今天</Text>
        </Space>
      </div>
    </div>
  );
};

// ============ Overview Tab ============

const OverviewTab: React.FC<{ project: Project }> = ({ project }) => {
  return (
    <div>
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label="项目编码"><Text code>{project.code}</Text></Descriptions.Item>
        <Descriptions.Item label="项目名称"><Text strong>{project.name}</Text></Descriptions.Item>
        <Descriptions.Item label="当前阶段"><Tag color={phaseColors[project.phase]}>{project.phase?.toUpperCase()}</Tag></Descriptions.Item>
        <Descriptions.Item label="状态">
          <Badge status={statusColors[project.status] as any} text={
            project.status === 'planning' ? '规划中' :
            project.status === 'active' ? '进行中' :
            project.status === 'completed' ? '已完成' :
            project.status === 'on_hold' ? '暂停' : project.status
          } />
        </Descriptions.Item>
        <Descriptions.Item label="进度"><Progress percent={project.progress} size="small" style={{ width: 200 }} /></Descriptions.Item>
        <Descriptions.Item label="项目经理">{project.manager_name || '-'}</Descriptions.Item>
        <Descriptions.Item label="开始日期">{project.start_date ? dayjs(project.start_date).format('YYYY-MM-DD') : '-'}</Descriptions.Item>
        <Descriptions.Item label="计划结束">{project.planned_end ? dayjs(project.planned_end).format('YYYY-MM-DD') : '-'}</Descriptions.Item>
        <Descriptions.Item label="关联产品" span={2}>{project.product_name || '-'}</Descriptions.Item>
        <Descriptions.Item label="项目描述" span={2}>
          <Paragraph style={{ margin: 0 }}>{project.description || '暂无描述'}</Paragraph>
        </Descriptions.Item>
      </Descriptions>
    </div>
  );
};

// ============ BOM Tab ============

const BOMTab: React.FC<{ projectId: string }> = ({ projectId }) => {
  const queryClient = useQueryClient();
  const [phaseFilter, setPhaseFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [selectedBomId, setSelectedBomId] = useState<string | null>(null);
  const [form] = Form.useForm();

  const { data: bomData, isLoading, isError } = useQuery({
    queryKey: ['project-boms', projectId, phaseFilter, typeFilter],
    queryFn: () => projectBomApi.list(projectId, {
      phase: phaseFilter || undefined,
      bom_type: typeFilter || undefined,
    }),
    retry: false,
  });

  const { data: bomDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['project-bom-detail', projectId, selectedBomId],
    queryFn: () => projectBomApi.get(projectId, selectedBomId!),
    enabled: !!selectedBomId && detailModalOpen,
    retry: false,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateProjectBOMRequest) => projectBomApi.create(projectId, data),
    onSuccess: () => {
      message.success('BOM创建成功');
      setCreateModalOpen(false);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('创建失败'),
  });

  const bomStatusConfig: Record<string, { color: string; text: string }> = {
    draft: { color: 'default', text: '草稿' },
    submitted: { color: 'processing', text: '待审批' },
    approved: { color: 'blue', text: '已审批' },
    rejected: { color: 'error', text: '已驳回' },
    released: { color: 'success', text: '已发布' },
    frozen: { color: 'purple', text: '已冻结' },
  };

  const bomTypeConfig: Record<string, { emoji: string; label: string }> = {
    EBOM: { emoji: '📗', label: '电子BOM' },
    SBOM: { emoji: '📘', label: '结构BOM' },
    MBOM: { emoji: '📙', label: '制造BOM' },
  };

  const bomItemColumns: ColumnsType<any> = [
    { title: '序号', dataIndex: 'sequence', key: 'sequence', width: 60 },
    { title: '物料名称', dataIndex: 'material_name', key: 'material_name', width: 160 },
    { title: '规格', dataIndex: 'specification', key: 'specification', width: 140 },
    { title: '数量', dataIndex: 'quantity', key: 'quantity', width: 80 },
    { title: '单位', dataIndex: 'unit', key: 'unit', width: 60 },
    { title: '供应商', dataIndex: 'supplier', key: 'supplier', width: 120 },
    { title: '单价', dataIndex: 'unit_price', key: 'unit_price', width: 100, render: (v: number) => v != null ? `¥${v.toFixed(2)}` : '-' },
    { title: '位号', dataIndex: 'position', key: 'position', width: 100 },
    { title: '关键物料', dataIndex: 'is_critical', key: 'is_critical', width: 80, render: (v: boolean) => v ? <Tag color="red">是</Tag> : <Tag>否</Tag> },
  ];

  // Group BOMs by phase
  const bomsByPhase = useMemo(() => {
    const boms = bomData?.items || [];
    const grouped: Record<string, ProjectBOM[]> = {};
    boms.forEach(bom => {
      const phase = bom.phase || '未分类';
      if (!grouped[phase]) grouped[phase] = [];
      grouped[phase].push(bom);
    });
    return grouped;
  }, [bomData]);

  if (isError) {
    return <Empty description="BOM数据暂不可用（API开发中）" image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Space>
          <Text strong>BOM管理</Text>
          <Select
            placeholder="阶段筛选"
            allowClear
            style={{ width: 120 }}
            value={phaseFilter || undefined}
            onChange={v => setPhaseFilter(v || '')}
            options={[
              { label: 'Concept', value: 'concept' },
              { label: 'EVT', value: 'evt' },
              { label: 'DVT', value: 'dvt' },
              { label: 'PVT', value: 'pvt' },
              { label: 'MP', value: 'mp' },
            ]}
          />
          <Select
            placeholder="类型筛选"
            allowClear
            style={{ width: 120 }}
            value={typeFilter || undefined}
            onChange={v => setTypeFilter(v || '')}
            options={[
              { label: 'EBOM', value: 'EBOM' },
              { label: 'SBOM', value: 'SBOM' },
              { label: 'MBOM', value: 'MBOM' },
            ]}
          />
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>提交BOM</Button>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 40 }}><Spin /></div>
      ) : Object.keys(bomsByPhase).length === 0 ? (
        <Empty description="暂无BOM数据" />
      ) : (
        Object.entries(bomsByPhase).map(([phase, boms]) => (
          <div key={phase} style={{ marginBottom: 24 }}>
            <div style={{ borderBottom: '2px solid #e8e8e8', paddingBottom: 4, marginBottom: 12 }}>
              <Tag color={phaseColors[phase] || 'default'}>{phase.toUpperCase()} 阶段</Tag>
            </div>
            {boms.map(bom => {
              const typeInfo = bomTypeConfig[bom.bom_type] || { emoji: '📄', label: bom.bom_type };
              const statusInfo = bomStatusConfig[bom.status] || { color: 'default', text: bom.status };
              return (
                <Card key={bom.id} size="small" style={{ marginBottom: 8 }}
                  bodyStyle={{ padding: '12px 16px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div>
                      <Space>
                        <Text strong style={{ fontSize: 14 }}>{typeInfo.emoji} {bom.name || typeInfo.label} v{bom.version}</Text>
                        <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
                        <Text type="secondary">{bom.creator_name || '-'}</Text>
                        <Text type="secondary">{bom.created_at ? dayjs(bom.created_at).format('YYYY-MM-DD') : ''}</Text>
                      </Space>
                      <div style={{ marginTop: 4 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {bom.total_items || 0}个物料
                          {bom.total_cost != null ? ` | ¥${bom.total_cost.toFixed(2)}/台` : ''}
                        </Text>
                      </div>
                    </div>
                    <Space>
                      <Button size="small" icon={<EyeOutlined />} onClick={() => { setSelectedBomId(bom.id); setDetailModalOpen(true); }}>查看详情</Button>
                      <Button size="small" icon={<ExportOutlined />}>导出</Button>
                    </Space>
                  </div>
                </Card>
              );
            })}
          </div>
        ))
      )}

      {/* Create BOM Modal */}
      <Modal
        title="提交BOM"
        open={createModalOpen}
        onCancel={() => { setCreateModalOpen(false); form.resetFields(); }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending}
      >
        <Form form={form} layout="vertical" onFinish={(values) => createMutation.mutate(values)}>
          <Form.Item name="name" label="BOM名称" rules={[{ required: true, message: '请输入BOM名称' }]}>
            <Input placeholder="如：EVT电子BOM" />
          </Form.Item>
          <Form.Item name="bom_type" label="BOM类型" rules={[{ required: true, message: '请选择BOM类型' }]}>
            <Select options={[
              { label: 'EBOM - 电子BOM', value: 'EBOM' },
              { label: 'SBOM - 结构BOM', value: 'SBOM' },
              { label: 'MBOM - 制造BOM', value: 'MBOM' },
            ]} />
          </Form.Item>
          <Form.Item name="phase" label="所属阶段" rules={[{ required: true, message: '请选择阶段' }]}>
            <Select options={PHASES.map(p => ({ label: p.toUpperCase(), value: p }))} />
          </Form.Item>
          <Form.Item name="version" label="版本号">
            <Input placeholder="如：v1.0" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="BOM描述信息" />
          </Form.Item>
        </Form>
      </Modal>

      {/* BOM Detail Modal */}
      <Modal
        title={`BOM详情 - ${bomDetail?.name || ''}`}
        open={detailModalOpen}
        onCancel={() => { setDetailModalOpen(false); setSelectedBomId(null); }}
        width={1100}
        footer={null}
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: 40 }}><Spin /></div>
        ) : bomDetail ? (
          <div>
            <Descriptions size="small" column={3} style={{ marginBottom: 16 }}>
              <Descriptions.Item label="BOM名称">{bomDetail.name}</Descriptions.Item>
              <Descriptions.Item label="类型"><Tag>{bomDetail.bom_type}</Tag></Descriptions.Item>
              <Descriptions.Item label="版本">v{bomDetail.version}</Descriptions.Item>
              <Descriptions.Item label="阶段"><Tag color={phaseColors[bomDetail.phase]}>{bomDetail.phase?.toUpperCase()}</Tag></Descriptions.Item>
              <Descriptions.Item label="状态"><Tag color={bomStatusConfig[bomDetail.status]?.color}>{bomStatusConfig[bomDetail.status]?.text}</Tag></Descriptions.Item>
              <Descriptions.Item label="物料数量">{bomDetail.total_items || bomDetail.items?.length || 0}</Descriptions.Item>
            </Descriptions>
            <Table
              columns={bomItemColumns}
              dataSource={bomDetail.items || []}
              rowKey="id"
              size="small"
              pagination={false}
              scroll={{ y: 400 }}
              locale={{ emptyText: '暂无物料行项' }}
            />
          </div>
        ) : (
          <Empty description="无法加载BOM详情" />
        )}
      </Modal>
    </div>
  );
};

// ============ Documents Tab ============

const DocumentsTab: React.FC<{ projectId: string }> = ({ projectId }) => {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['project-documents', projectId],
    queryFn: () => documentsApi.list({ related_type: 'project', related_id: projectId }),
    retry: false,
  });

  const columns: ColumnsType<Document> = [
    { title: '文档编号', dataIndex: 'code', key: 'code', width: 140, render: (t: string) => <Text code>{t}</Text> },
    { title: '标题', dataIndex: 'title', key: 'title', width: 200 },
    { title: '分类', dataIndex: 'category', key: 'category', width: 100, render: (_, record) => (record.category as any)?.name || (typeof record.category === 'string' ? record.category : '-') },
    { title: '版本', dataIndex: 'version', key: 'version', width: 80 },
    { title: '状态', dataIndex: 'status', key: 'status', width: 100,
      render: (s: string) => <Tag color={s === 'released' ? 'success' : s === 'draft' ? 'default' : 'warning'}>{s === 'released' ? '已发布' : s === 'draft' ? '草稿' : s}</Tag>
    },
    { title: '上传者', dataIndex: 'created_by_name', key: 'created_by_name', width: 100, render: (v: string, record) => v || record.uploader?.name || '-' },
    { title: '更新时间', dataIndex: 'updated_at', key: 'updated_at', width: 160, render: (d: string) => d ? dayjs(d).format('YYYY-MM-DD HH:mm') : '-' },
  ];

  if (isError) {
    return <Empty description="文档数据暂不可用（API开发中）" image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Text strong>图纸文档</Text>
        <Button icon={<UploadOutlined />}>上传文档</Button>
      </div>
      <Table
        columns={columns}
        dataSource={data?.items || []}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={{ pageSize: 10, showTotal: (t) => `共 ${t} 条` }}
        locale={{ emptyText: '暂无文档' }}
      />
    </div>
  );
};

// ============ Deliverables Tab ============

const DeliverablesTab: React.FC<{ projectId: string; currentPhase: string }> = ({ projectId, currentPhase }) => {
  const [selectedPhase, setSelectedPhase] = useState(currentPhase?.toLowerCase() || 'evt');

  const { data, isLoading, isError } = useQuery({
    queryKey: ['project-deliverables', projectId, selectedPhase],
    queryFn: () => deliverablesApi.list(projectId, selectedPhase),
    retry: false,
  });

  const deliverables = data?.items || [];
  const completed = deliverables.filter(d => d.status === 'approved' || d.status === 'submitted').length;
  const total = deliverables.length;
  const percent = total > 0 ? Math.round((completed / total) * 100) : 0;
  const allComplete = total > 0 && completed === total;
  const remaining = total - completed;

  const statusConfig: Record<string, { icon: string; color: string; text: string }> = {
    not_started: { icon: '⬜', color: '#999', text: '未开始' },
    in_progress: { icon: '🟡', color: '#faad14', text: '进行中' },
    submitted: { icon: '✅', color: '#52c41a', text: '已提交' },
    approved: { icon: '✅', color: '#52c41a', text: '已审批' },
    rejected: { icon: '❌', color: '#ff4d4f', text: '已驳回' },
  };

  if (isError) {
    return <Empty description="交付物数据暂不可用（API开发中）" image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Space>
          <Text strong>交付物清单</Text>
          <Select
            value={selectedPhase}
            onChange={setSelectedPhase}
            style={{ width: 120 }}
            options={PHASES.map(p => ({ label: `${p.toUpperCase()} 阶段`, value: p }))}
          />
        </Space>
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 40 }}><Spin /></div>
      ) : deliverables.length === 0 ? (
        <Empty description="暂无交付物" />
      ) : (
        <>
          <Card size="small" style={{ marginBottom: 16, background: '#fafafa' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
              <Text>完成进度: {completed}/{total} ({percent}%)</Text>
              <Progress percent={percent} style={{ flex: 1, maxWidth: 300 }} size="small"
                status={allComplete ? 'success' : 'active'} />
            </div>
          </Card>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {deliverables.map(d => {
              const sc = statusConfig[d.status] || statusConfig.not_started;
              return (
                <div key={d.id} style={{
                  display: 'flex', alignItems: 'center', padding: '10px 16px',
                  border: '1px solid #f0f0f0', borderRadius: 6, background: '#fff',
                }}>
                  <span style={{ fontSize: 16, marginRight: 12 }}>{sc.icon}</span>
                  <div style={{ flex: 1 }}>
                    <Text strong>{d.name}</Text>
                    {d.description && <Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>{d.description}</Text>}
                  </div>
                  <Text type="secondary" style={{ marginRight: 16 }}>{d.assignee_role || d.assignee_name || '-'}</Text>
                  <Tag color={sc.color === '#52c41a' ? 'success' : sc.color === '#ff4d4f' ? 'error' : sc.color === '#faad14' ? 'warning' : 'default'}>
                    {sc.text}
                  </Tag>
                </div>
              );
            })}
          </div>

          <div style={{ marginTop: 16, textAlign: 'right' }}>
            {!allComplete && (
              <Alert
                type="warning"
                showIcon
                icon={<WarningOutlined />}
                message={`还有 ${remaining} 项未完成，无法发起阶段门评审`}
                style={{ marginBottom: 12 }}
              />
            )}
            <Button type="primary" disabled={!allComplete}>
              发起阶段门评审
            </Button>
          </div>
        </>
      )}
    </div>
  );
};

// ============ ECN Tab ============

const ECNTab: React.FC<{ projectId: string; productId?: string }> = ({ productId }) => {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['project-ecns', productId],
    queryFn: () => ecnApi.list({ product_id: productId }),
    enabled: !!productId,
    retry: false,
  });

  const ecnStatusConfig: Record<string, { color: string; text: string }> = {
    draft: { color: 'default', text: '草稿' },
    pending: { color: 'processing', text: '待审批' },
    approved: { color: 'success', text: '已批准' },
    rejected: { color: 'error', text: '已驳回' },
    implemented: { color: 'purple', text: '已实施' },
  };

  const urgencyColors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    urgent: 'red',
  };

  const columns: ColumnsType<ECN> = [
    { title: 'ECN编号', dataIndex: 'code', key: 'code', width: 140, render: (t: string) => <Text code>{t}</Text> },
    { title: '标题', dataIndex: 'title', key: 'title', width: 200 },
    { title: '变更类型', dataIndex: 'change_type', key: 'change_type', width: 100 },
    { title: '紧急度', dataIndex: 'urgency', key: 'urgency', width: 80,
      render: (u: string) => <Tag color={urgencyColors[u] || 'default'}>{u === 'high' ? '高' : u === 'medium' ? '中' : u === 'urgent' ? '紧急' : '低'}</Tag>
    },
    { title: '状态', dataIndex: 'status', key: 'status', width: 100,
      render: (s: string) => { const cfg = ecnStatusConfig[s] || { color: 'default', text: s }; return <Tag color={cfg.color}>{cfg.text}</Tag>; }
    },
    { title: '申请人', key: 'requester', width: 100, render: (_, r) => r.requester?.name || '-' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 160, render: (d: string) => d ? dayjs(d).format('YYYY-MM-DD HH:mm') : '-' },
  ];

  if (isError) {
    return <Empty description="ECN数据暂不可用" image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  if (!productId) {
    return <Empty description="该项目未关联产品，无法查看ECN" image={Empty.PRESENTED_IMAGE_SIMPLE} />;
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Text strong>工程变更通知</Text>
      </div>
      <Table
        columns={columns}
        dataSource={data?.items || []}
        rowKey="id"
        loading={isLoading}
        size="small"
        pagination={{ pageSize: 10, showTotal: (t) => `共 ${t} 条` }}
        locale={{ emptyText: '暂无ECN记录' }}
      />
    </div>
  );
};

// ============ Main ProjectDetail Page ============

const ProjectDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: project, isLoading } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectApi.get(id!),
    enabled: !!id,
  });

  const { data: tasks, isLoading: tasksLoading } = useQuery({
    queryKey: ['project-tasks', id],
    queryFn: () => projectApi.listTasks(id!),
    enabled: !!id,
  });

  const completeTaskMutation = useMutation({
    mutationFn: ({ projectId, taskId }: { projectId: string; taskId: string }) =>
      projectApi.completeTask(projectId, taskId),
    onSuccess: () => {
      message.success('任务已完成');
      queryClient.invalidateQueries({ queryKey: ['project-tasks', id] });
    },
    onError: () => message.error('操作失败'),
  });

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '50vh' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (!project) {
    return (
      <div style={{ padding: 24 }}>
        <Empty description="项目不存在" />
        <div style={{ textAlign: 'center', marginTop: 16 }}>
          <Button onClick={() => navigate('/projects')}>返回项目列表</Button>
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/projects')} style={{ padding: 0, marginBottom: 8 }}>
          返回项目列表
        </Button>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <Title level={3} style={{ margin: 0 }}>
              {project.name}
              {project.code && <Text code style={{ marginLeft: 12, fontSize: 14 }}>{project.code}</Text>}
            </Title>
            <div style={{ marginTop: 8 }}>
              <PhaseProgressBar currentPhase={project.phase} />
            </div>
          </div>
          <Space>
            <Badge status={statusColors[project.status] as any} text={
              project.status === 'planning' ? '规划中' :
              project.status === 'active' ? '进行中' :
              project.status === 'completed' ? '已完成' :
              project.status === 'on_hold' ? '暂停' : project.status
            } />
            <Progress type="circle" percent={project.progress} size={48} />
          </Space>
        </div>
      </div>

      {/* Tabs */}
      <Card>
        <Tabs
          defaultActiveKey="overview"
          items={[
            {
              key: 'overview',
              label: '概览',
              children: <OverviewTab project={project} />,
            },
            {
              key: 'gantt',
              label: `甘特图 (${tasks?.length || 0})`,
              children: tasksLoading ? (
                <div style={{ textAlign: 'center', padding: 40 }}>加载中...</div>
              ) : tasks && tasks.length > 0 ? (
                <div style={{ height: 560 }}>
                  <GanttChart
                    tasks={tasks}
                    projectId={project.id}
                    onCompleteTask={(taskId) =>
                      completeTaskMutation.mutate({ projectId: project.id, taskId })
                    }
                    completingTask={completeTaskMutation.isPending}
                  />
                </div>
              ) : (
                <Empty description="暂无任务" />
              ),
            },
            {
              key: 'bom',
              label: 'BOM',
              children: <BOMTab projectId={project.id} />,
            },
            {
              key: 'documents',
              label: '图纸文档',
              children: <DocumentsTab projectId={project.id} />,
            },
            {
              key: 'deliverables',
              label: '交付物',
              children: <DeliverablesTab projectId={project.id} currentPhase={project.phase} />,
            },
            {
              key: 'ecn',
              label: 'ECN',
              children: <ECNTab projectId={project.id} productId={project.product_id} />,
            },
          ]}
        />
      </Card>
    </div>
  );
};

export default ProjectDetail;
