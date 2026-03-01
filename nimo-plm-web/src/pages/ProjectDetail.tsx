import React, { useState, useMemo, useRef, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useSSE, SSETaskEvent } from '@/hooks/useSSE';
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
  Drawer,
  Timeline,
  Avatar,
  Popconfirm,
  Checkbox,
  Upload,
  Divider,
  Radio,
} from 'antd';
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  RightOutlined,
  DownOutlined,
  PlusOutlined,
  UploadOutlined,
  DownloadOutlined,
  FileExcelOutlined,
  SwapOutlined,
  WarningOutlined,
  UserAddOutlined,
  UserOutlined,
  AuditOutlined,
  CloseCircleOutlined,
  DeleteOutlined,
  SendOutlined,
  LockOutlined,
  ShoppingCartOutlined,
} from '@ant-design/icons';
import { projectApi, Project, Task } from '@/api/projects';
import { projectBomApi, ProjectBOMItem, CreateProjectBOMRequest, BOMItemRequest } from '@/api/projectBom';
import { materialsApi, Material } from '@/api/materials';
import { deliverablesApi } from '@/api/deliverables';
import { ecnApi, ECN } from '@/api/ecn';
import { documentsApi, Document } from '@/api/documents';
import apiClient from '@/api/client';
import { uploadApi } from '@/api/upload';
import { srmApi } from '@/api/srm';
import { skuApi, ProductSKU, FullBOMItem } from '@/api/sku';
import { cmfVariantApi, type AppearancePartWithCMF, type CMFVariant } from '@/api/cmfVariant';
import { partDrawingApi, PartDrawing } from '@/api/partDrawing';
import UserSelect from '@/components/UserSelect';
import CMFEditControl from '@/components/CMFEditControl';
import { EBOMControl, PBOMControl, MBOMControl, type BOMControlConfig } from '@/components/BOM';
import { ROLE_CODES, taskRoleApi, TaskRole } from '@/constants/roles';
import type { ColumnsType } from 'antd/es/table';
import { useIsMobile } from '@/hooks/useIsMobile';
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
  unassigned: { color: 'default', text: '待指派', icon: <UserAddOutlined />, barColor: '#d9d9d9' },
  pending: { color: 'default', text: '待开始', icon: <ClockCircleOutlined />, barColor: '#bfbfbf' },
  in_progress: { color: 'processing', text: '进行中', icon: <ClockCircleOutlined />, barColor: '#1677ff' },
  submitted: { color: 'warning', text: '已提交', icon: <CheckCircleOutlined />, barColor: '#faad14' },
  reviewing: { color: 'warning', text: '审批中', icon: <AuditOutlined />, barColor: '#faad14' },
  completed: { color: 'success', text: '已完成', icon: <CheckCircleOutlined />, barColor: '#52c41a' },
  rejected: { color: 'error', text: '已驳回', icon: <CloseCircleOutlined />, barColor: '#ff4d4f' },
};

const GANTT_ROW_HEIGHT = 36;
const GANTT_HEADER_HEIGHT = 50;
const DAY_WIDTH = 28;
const LEFT_PANEL_WIDTH = 650;

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
  onRefresh: () => void;
}> = ({ tasks, projectId, onCompleteTask: _onCompleteTask, completingTask: _completingTask, onRefresh }) => {
  const isMobileView = useIsMobile();
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

  // ===== Mobile: Task List View =====
  if (isMobileView) {
    const mobileStatusIcon: Record<string, string> = {
      completed: '\u2705', in_progress: '\ud83d\udfe2', submitted: '\ud83d\udfe1', reviewing: '\ud83d\udfe1',
      pending: '\u23f3', unassigned: '\u2b1c', rejected: '\ud83d\udd34',
    };
    const mobileGroups = groupedData.filter(g => g.tasks.length > 0);

    return (
      <div className="gantt-mobile-list">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 4px', marginBottom: 4 }}>
          <Text type="secondary" style={{ fontSize: 13 }}>{tasks.length} 个任务</Text>
        </div>
        {mobileGroups.map(group => {
          const collapsed = collapsedGroups.has(group.phase);
          return (
            <div key={group.phase} className="gantt-mobile-phase">
              <div
                className="gantt-mobile-phase-header"
                onClick={() => toggleGroup(group.phase)}
              >
                <RightOutlined className={`gantt-mobile-phase-chevron ${collapsed ? '' : 'expanded'}`} />
                <Tag color={phaseColors[group.phase] || 'default'} style={{ margin: 0 }}>{group.label}</Tag>
                <Text type="secondary" style={{ fontSize: 12, marginLeft: 'auto' }}>({group.tasks.length})</Text>
              </div>
              {!collapsed && (
                <div className="gantt-mobile-phase-body">
                  {group.tasks.map(task => {
                    const config = taskStatusConfig[task.status] || taskStatusConfig.pending;
                    const icon = mobileStatusIcon[task.status] || '\u23f3';
                    const startStr = task.start_date ? dayjs(task.start_date).format('M/D') : '';
                    const endStr = task.due_date ? dayjs(task.due_date).format('M/D') : '';
                    const dateRange = startStr && endStr ? `${startStr}-${endStr}` : (startStr || endStr || '');
                    const hasChildren = task.children.length > 0;
                    const isCollapsed = collapsedTasks.has(task.id);
                    return (
                      <div key={task.id} className="gantt-mobile-task" style={{ paddingLeft: 12 + task.depth * 16 }}>
                        <div className="gantt-mobile-task-row">
                          {hasChildren ? (
                            <span style={{ cursor: 'pointer', width: 18, flexShrink: 0, textAlign: 'center' }} onClick={() => toggleTask(task.id)}>
                              {isCollapsed ? <RightOutlined style={{ fontSize: 10 }} /> : <DownOutlined style={{ fontSize: 10 }} />}
                            </span>
                          ) : <span style={{ width: 18, flexShrink: 0 }} />}
                          <span style={{ fontSize: 14, flexShrink: 0 }}>{icon}</span>
                          <span className="gantt-mobile-task-name" style={{ color: task.is_critical ? '#cf1322' : undefined }}>
                            {task.title}
                          </span>
                          <span className="gantt-mobile-task-date">{dateRange}</span>
                        </div>
                        {/* Mini progress bar */}
                        <div className="gantt-mobile-task-progress-track" style={{ marginLeft: 18 + (hasChildren ? 18 : 0) }}>
                          <div className="gantt-mobile-task-progress-fill" style={{ width: `${task.progress || 0}%`, background: config.barColor }} />
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>
    );
  }

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
            <span style={{ width: 100, textAlign: 'center' }}>负责人</span>
            <span style={{ width: 50, textAlign: 'center' }}>状态</span>
            <span style={{ width: 45, textAlign: 'center' }}>进度</span>
            <span style={{ width: 130, textAlign: 'center' }}>操作</span>
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
                    <span style={{ width: 100, flexShrink: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: '#666' }}>
                      {(task.assignee?.name || task.assignee_name) ? (
                        <Tooltip title={task.assignee?.name || task.assignee_name}>
                          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                            <Avatar size={18} src={task.assignee?.avatar_url} icon={<UserOutlined />} style={{ fontSize: 10, flexShrink: 0 }}>
                              {(task.assignee?.name || task.assignee_name)?.[0]}
                            </Avatar>
                            <span style={{ fontSize: 11, overflow: 'hidden', textOverflow: 'ellipsis' }}>{task.assignee?.name || task.assignee_name}</span>
                          </span>
                        </Tooltip>
                      ) : '-'}
                    </span>
                    <span style={{ width: 50, textAlign: 'center', flexShrink: 0 }}>
                      <Tag color={config.color} style={{ fontSize: 10, padding: '0 4px', margin: 0, lineHeight: '18px' }}>{config.text}</Tag>
                    </span>
                    <span style={{ width: 45, textAlign: 'center', flexShrink: 0, fontSize: 11, color: '#666' }}>{task.progress}%</span>
                    <span style={{ width: 130, textAlign: 'center', flexShrink: 0 }} onClick={(e) => e.stopPropagation()}>
                      <TaskActions task={task} projectId={projectId} onRefresh={onRefresh} />
                    </span>
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
  const isMobileOverview = useIsMobile();
  const statusText = project.status === 'planning' ? '规划中' :
    project.status === 'active' ? '进行中' :
    project.status === 'completed' ? '已完成' :
    project.status === 'on_hold' ? '暂停' : project.status;

  if (isMobileOverview) {
    const tagClass = project.status === 'active' ? 'ds-tag-processing' :
      project.status === 'completed' ? 'ds-tag-success' :
      project.status === 'on_hold' ? 'ds-tag-warning' :
      project.status === 'cancelled' ? 'ds-tag-danger' : 'ds-tag-default';
    return (
      <div className="ds-detail-page" style={{ padding: 0 }}>
        <div className="ds-detail-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
            <div className="ds-card-title" style={{ flex: 1 }}>{project.name}</div>
            <span className={`ds-tag ${tagClass}`}>{statusText}</span>
          </div>
          <div className="ds-card-subtitle" style={{ fontFamily: 'monospace' }}>{project.code}</div>
        </div>
        <div className="ds-detail-section">
          <div className="ds-section-title">基本信息</div>
          <div className="ds-info-row">
            <span className="ds-info-label">当前阶段</span>
            <span className="ds-info-value"><Tag color={phaseColors[project.phase]} style={{ margin: 0 }}>{project.phase?.toUpperCase()}</Tag></span>
          </div>
          <div className="ds-info-row">
            <span className="ds-info-label">进度</span>
            <span className="ds-info-value"><Progress percent={project.progress} size="small" style={{ width: 120 }} /></span>
          </div>
          <div className="ds-info-row">
            <span className="ds-info-label">项目经理</span>
            <span className="ds-info-value">{project.manager_name || '-'}</span>
          </div>
          <div className="ds-info-row">
            <span className="ds-info-label">开始日期</span>
            <span className="ds-info-value">{project.start_date ? dayjs(project.start_date).format('YYYY-MM-DD') : '-'}</span>
          </div>
          <div className="ds-info-row">
            <span className="ds-info-label">计划结束</span>
            <span className="ds-info-value">{project.planned_end ? dayjs(project.planned_end).format('YYYY-MM-DD') : '-'}</span>
          </div>
          <div className="ds-info-row">
            <span className="ds-info-label">关联产品</span>
            <span className="ds-info-value">{project.product_name || '-'}</span>
          </div>
        </div>
        {project.description && (
          <div className="ds-detail-section">
            <div className="ds-section-title">项目描述</div>
            <div style={{ fontSize: 14, color: 'var(--ds-text-primary)', lineHeight: 1.6 }}>{project.description}</div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div>
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label="项目编码"><Text code>{project.code}</Text></Descriptions.Item>
        <Descriptions.Item label="项目名称"><Text strong>{project.name}</Text></Descriptions.Item>
        <Descriptions.Item label="当前阶段"><Tag color={phaseColors[project.phase]}>{project.phase?.toUpperCase()}</Tag></Descriptions.Item>
        <Descriptions.Item label="状态">
          <Badge status={statusColors[project.status] as any} text={statusText} />
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

// ============ BOM Tab - Full Editor ============

const BOM_STATUS_CONFIG: Record<string, { color: string; text: string }> = {
  draft: { color: 'default', text: '草稿' },
  released: { color: 'success', text: '已发布' },
  obsolete: { color: 'default', text: '已废弃' },
  pending_review: { color: 'processing', text: '待审批' },
  published: { color: 'success', text: '已发布' },
  rejected: { color: 'error', text: '已驳回' },
  frozen: { color: 'purple', text: '已冻结' },
};

// Material Search Modal
const MaterialSearchModal: React.FC<{
  open: boolean;
  onClose: () => void;
  onSelect: (material: Material) => void;
}> = ({ open, onClose, onSelect }) => {
  const [search, setSearch] = useState('');
  const { data, isLoading } = useQuery({
    queryKey: ['materials-search', search],
    queryFn: () => materialsApi.list({ search: search || undefined }),
    enabled: open,
  });

  const materials = data?.materials || [];

  const columns: ColumnsType<Material> = [
    { title: '编码', dataIndex: 'code', width: 120, render: (v: string) => <Text code>{v}</Text> },
    { title: '名称', dataIndex: 'name', width: 160 },
    { title: '规格', dataIndex: 'description', width: 200, ellipsis: true },
    { title: '单位', dataIndex: 'unit', width: 60 },
    { title: '标准成本', dataIndex: 'standard_cost', width: 100, render: (v: number) => v != null ? `¥${v.toFixed(2)}` : '-' },
    {
      title: '操作', width: 80, render: (_, record) => (
        <Button size="small" type="link" onClick={() => { onSelect(record); onClose(); }}>选择</Button>
      ),
    },
  ];

  return (
    <Modal title="物料选择" open={open} onCancel={onClose} width={800} footer={null}>
      <Input.Search
        placeholder="按名称/编码/规格搜索"
        allowClear
        onSearch={setSearch}
        onChange={e => { if (!e.target.value) setSearch(''); }}
        style={{ marginBottom: 12 }}
      />
      <Table
        columns={columns}
        dataSource={materials}
        rowKey="id"
        size="small"
        loading={isLoading}
        pagination={{ pageSize: 8, showTotal: (t) => `共 ${t} 条` }}
        scroll={{ y: 350 }}
        locale={{ emptyText: '暂无物料数据' }}
      />
    </Modal>
  );
};

const BOMTab: React.FC<{ projectId: string }> = ({ projectId }) => {
  const queryClient = useQueryClient();
  const isMobile = useIsMobile();
  const [activeTab, setActiveTab] = useState<string>('EBOM');
  const [selectedBomId, setSelectedBomId] = useState<string | null>(null);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [rejectModalOpen, setRejectModalOpen] = useState(false);
  const [rejectComment, setRejectComment] = useState('');
  const [materialModalOpen, setMaterialModalOpen] = useState(false);
  const [editingRowId, setEditingRowId] = useState<string | null>(null);
  const [compareModalOpen, setCompareModalOpen] = useState(false);
  const [compareBom1, setCompareBom1] = useState<string | undefined>(undefined);
  const [compareBom2, setCompareBom2] = useState<string | undefined>(undefined);
  const [compareResult, setCompareResult] = useState<any[] | null>(null);
  const [compareLoading, setCompareLoading] = useState(false);
  const [exportLoading, setExportLoading] = useState(false);
  const [importLoading, setImportLoading] = useState(false);
  const [drawingHistoryOpen, setDrawingHistoryOpen] = useState(false);
  const [drawingHistoryItemId, _setDrawingHistoryItemId] = useState<string>('');
  const [drawingHistoryType, _setDrawingHistoryType] = useState<'2D' | '3D'>('2D');
  const [drawingUploadModalOpen, setDrawingUploadModalOpen] = useState(false);
  const [drawingUploadItemId, _setDrawingUploadItemId] = useState<string>('');
  const [drawingUploadType, _setDrawingUploadType] = useState<'2D' | '3D'>('2D');
  const [drawingChangeDesc, setDrawingChangeDesc] = useState('');
  const [releaseModalOpen, setReleaseModalOpen] = useState(false);
  const [releaseNote, setReleaseNote] = useState('');
  const [form] = Form.useForm();

  // Fetch BOM list
  const { data: bomList = [], isLoading: listLoading } = useQuery({
    queryKey: ['project-boms', projectId],
    queryFn: () => projectBomApi.list(projectId),
    retry: false,
  });

  // Fetch selected BOM detail
  const { data: bomDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['project-bom-detail', projectId, selectedBomId],
    queryFn: () => projectBomApi.get(projectId, selectedBomId!),
    enabled: !!selectedBomId,
    retry: false,
  });

  // Fetch drawings for all items in selected BOM
  const { data: drawingsByBOM = {} } = useQuery({
    queryKey: ['bom-drawings', projectId, selectedBomId],
    queryFn: () => partDrawingApi.listByBOM(projectId, selectedBomId!),
    enabled: !!selectedBomId,
    retry: false,
  });

  // BOMs filtered by active tab type
  const filteredBomList = useMemo(() =>
    bomList.filter(b => b.bom_type === activeTab),
  [bomList, activeTab]);

  // Auto-select first BOM of active tab type
  useEffect(() => {
    if (filteredBomList.length > 0) {
      setSelectedBomId(filteredBomList[0].id);
    } else {
      setSelectedBomId(null);
    }
  }, [filteredBomList]);

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: CreateProjectBOMRequest) => projectBomApi.create(projectId, data),
    onSuccess: (bom) => {
      message.success('BOM创建成功');
      setCreateModalOpen(false);
      form.resetFields();
      setSelectedBomId(bom.id);
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('创建失败'),
  });

  const submitMutation = useMutation({
    mutationFn: () => projectBomApi.submit(projectId, selectedBomId!),
    onSuccess: () => {
      message.success('已提交审批');
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('提交失败'),
  });

  const approveMutation = useMutation({
    mutationFn: () => projectBomApi.approve(projectId, selectedBomId!),
    onSuccess: () => {
      message.success('审批通过');
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('操作失败'),
  });

  const rejectMutation = useMutation({
    mutationFn: (comment: string) => projectBomApi.reject(projectId, selectedBomId!, comment),
    onSuccess: () => {
      message.success('已驳回');
      setRejectModalOpen(false);
      setRejectComment('');
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('操作失败'),
  });

  const freezeMutation = useMutation({
    mutationFn: () => projectBomApi.freeze(projectId, selectedBomId!),
    onSuccess: () => {
      message.success('已冻结');
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('操作失败'),
  });

  const addItemMutation = useMutation({
    mutationFn: (data: BOMItemRequest) => projectBomApi.addItem(projectId, selectedBomId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('添加失败'),
  });

  const updateItemMutation = useMutation({
    mutationFn: ({ itemId, data }: { itemId: string; data: BOMItemRequest }) =>
      projectBomApi.updateItem(projectId, selectedBomId!, itemId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('更新失败'),
  });

  const convertToMBOMMutation = useMutation({
    mutationFn: () => projectBomApi.convertToMBOM(projectId, selectedBomId!),
    onSuccess: () => {
      message.success('已创建MBOM副本');
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: () => message.error('转换失败'),
  });

  const releaseMutation = useMutation({
    mutationFn: (note: string) => projectBomApi.release(projectId, selectedBomId!, note),
    onSuccess: (bom) => {
      message.success(`已发布 ${bom.version}`);
      setReleaseModalOpen(false);
      setReleaseNote('');
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: (err: any) => message.error(err?.response?.data?.message || '发布失败'),
  });

  const createFromMutation = useMutation({
    mutationFn: ({ sourceBomId, targetType }: { sourceBomId: string; targetType: string }) =>
      projectBomApi.createFrom(projectId, sourceBomId, targetType),
    onSuccess: (bom) => {
      message.success(`已创建${bom.bom_type}草稿`);
      setActiveTab(bom.bom_type);
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    },
    onError: (err: any) => message.error(err?.response?.data?.message || '创建失败'),
  });

  const submitToSRMMutation = useMutation({
    mutationFn: () => srmApi.createPRFromBOM({ project_id: projectId, bom_id: selectedBomId! }),
    onSuccess: (pr) => {
      message.success(`已创建采购需求 ${pr.pr_code}`);
    },
    onError: () => message.error('提交到SRM失败'),
  });

  // Export Excel handler
  const handleExportExcel = async () => {
    if (!selectedBomId) return;
    setExportLoading(true);
    try {
      await projectBomApi.exportExcel(projectId, selectedBomId);
      message.success('导出成功');
    } catch {
      message.error('导出失败');
    } finally {
      setExportLoading(false);
    }
  };

  // Import Excel handler
  const handleImportExcel = async (file: File) => {
    if (!selectedBomId) return;
    setImportLoading(true);
    try {
      const result = await projectBomApi.importExcel(projectId, selectedBomId, file);
      message.success(`导入成功：创建${result?.created ?? 0}项，匹配物料${result?.matched ?? 0}项，自动建料${result?.auto_created ?? 0}项，错误${result?.errors ?? 0}项`);
      queryClient.invalidateQueries({ queryKey: ['project-bom-detail', projectId, selectedBomId] });
      queryClient.invalidateQueries({ queryKey: ['project-boms', projectId] });
    } catch {
      message.error('导入失败');
    } finally {
      setImportLoading(false);
    }
  };

  // Download template handler
  const handleDownloadTemplate = async () => {
    try {
      await projectBomApi.downloadTemplate(bomDetail?.bom_type);
    } catch {
      message.error('下载模板失败');
    }
  };

  // Compare BOMs handler
  const handleCompare = async () => {
    if (!compareBom1 || !compareBom2) {
      message.warning('请选择两个BOM进行对比');
      return;
    }
    setCompareLoading(true);
    try {
      const result = await projectBomApi.compareBOMs(compareBom1, compareBom2);
      setCompareResult(result);
    } catch {
      message.error('对比失败');
    } finally {
      setCompareLoading(false);
    }
  };

  // Select material from library
  const handleMaterialSelect = (material: Material) => {
    if (editingRowId) {
      updateItemMutation.mutate({
        itemId: editingRowId,
        data: {
          material_id: material.id,
          name: material.name,
          specification: material.description,
          unit: material.unit || 'pcs',
          unit_price: material.standard_cost || undefined,
          quantity: 1,
        },
      });
    } else {
      addItemMutation.mutate({
        material_id: material.id,
        name: material.name,
        specification: material.description,
        unit: material.unit || 'pcs',
        unit_price: material.standard_cost || undefined,
        quantity: 1,
        item_number: (bomDetail?.items?.length || 0) + 1,
      });
    }
    setEditingRowId(null);
  };

  // 新版图纸上传：创建PartDrawing版本记录
  const handleDrawingVersionUpload = async (file: File) => {
    try {
      const result = await uploadApi.uploadFile(file);
      await partDrawingApi.upload(projectId, drawingUploadItemId, {
        drawing_type: drawingUploadType,
        file_id: result.id,
        file_name: result.filename,
        file_size: file.size,
        change_description: drawingChangeDesc,
      });
      message.success('上传成功');
      setDrawingUploadModalOpen(false);
      setDrawingChangeDesc('');
      queryClient.invalidateQueries({ queryKey: ['bom-drawings', projectId, selectedBomId] });
    } catch {
      message.error('上传失败');
    }
    return false;
  };

  const bomType = bomDetail?.bom_type || 'EBOM';

  // Stats — flatten extended_attrs for table compatibility
  // Strip GORM relation objects to prevent React error #31 (objects as children)
  // Extract material_code from the relation object before stripping
  const items = (bomDetail?.items || []).map(({ material, children, ...rest }) => ({
    ...rest,
    ...(rest.extended_attrs || {}),
    material_code: material?.code || '',
  }));
  const totalItems = items.length;
  const totalCost = items.reduce((sum, item) => {
    const cost = item.extended_cost ?? (item.quantity && item.unit_price ? item.quantity * item.unit_price : 0);
    return sum + (cost || 0);
  }, 0);
  // PBOM stats
  const isPBOM = bomType === 'PBOM';
  const totalTargetPrice = items.reduce((sum, item) => {
    const price = Number(item.extended_attrs?.target_price) || 0;
    return sum + price * (item.quantity || 1);
  }, 0);
  const totalTooling = items.reduce((sum, item) => sum + (Number(item.extended_attrs?.tooling_estimate) || 0), 0);

  // Action buttons based on status
  const renderActions = () => {
    if (!bomDetail) return null;
    const s = bomDetail.status;
    return (
      <Space split={<Divider type="vertical" />}>
        <Space>
          {(s === 'draft' || s === 'rejected') && (
            <Popconfirm title="确认提交审批？" onConfirm={() => submitMutation.mutate()}>
              <Button type="primary" icon={<SendOutlined />} loading={submitMutation.isPending}>提交审批</Button>
            </Popconfirm>
          )}
          {s === 'pending_review' && (
            <>
              <Popconfirm title="确认审批通过？" onConfirm={() => approveMutation.mutate()}>
                <Button type="primary" style={{ background: '#52c41a', borderColor: '#52c41a' }}
                  icon={<CheckCircleOutlined />} loading={approveMutation.isPending}>通过</Button>
              </Popconfirm>
              <Button danger icon={<CloseCircleOutlined />} onClick={() => setRejectModalOpen(true)}>驳回</Button>
            </>
          )}
          {s === 'published' && (
            <Popconfirm title="冻结后BOM不可再修改，确认冻结？" onConfirm={() => freezeMutation.mutate()}>
              <Button icon={<LockOutlined />} loading={freezeMutation.isPending}>冻结</Button>
            </Popconfirm>
          )}
          {s === 'frozen' && <Tag color="purple" icon={<LockOutlined />}>已冻结 - 只读</Tag>}
        </Space>
        <Space>
          <Tooltip title="导出Excel">
            <Button icon={<DownloadOutlined />} loading={exportLoading} onClick={handleExportExcel}>导出Excel</Button>
          </Tooltip>
          <Upload
            accept=".xlsx,.xls,.rep"
            showUploadList={false}
            beforeUpload={(file) => { handleImportExcel(file); return false; }}
            disabled={!(s === 'draft' || s === 'rejected')}
          >
            <Tooltip title={s === 'draft' || s === 'rejected' ? '支持Excel(.xlsx)和PADS(.rep)格式' : '仅草稿/已驳回状态可导入'}>
              <Button icon={<UploadOutlined />} loading={importLoading} disabled={!(s === 'draft' || s === 'rejected')}>导入BOM</Button>
            </Tooltip>
          </Upload>
          <Tooltip title="下载导入模板">
            <Button icon={<FileExcelOutlined />} onClick={handleDownloadTemplate}>下载模板</Button>
          </Tooltip>
          {bomDetail.bom_type === 'EBOM' && (s === 'published' || s === 'frozen') && (
            <Popconfirm title="确认将此EBOM转为MBOM副本？" onConfirm={() => convertToMBOMMutation.mutate()}>
              <Button icon={<SwapOutlined />} loading={convertToMBOMMutation.isPending}>转为MBOM</Button>
            </Popconfirm>
          )}
          <Tooltip title="版本对比">
            <Button icon={<SwapOutlined />} onClick={() => { setCompareModalOpen(true); setCompareResult(null); setCompareBom1(undefined); setCompareBom2(undefined); }}>版本对比</Button>
          </Tooltip>
          {items.length > 0 && (
            <Popconfirm title="确认将此BOM提交到SRM创建采购需求？" onConfirm={() => submitToSRMMutation.mutate()}>
              <Button type="primary" icon={<ShoppingCartOutlined />} loading={submitToSRMMutation.isPending}
                style={{ background: '#722ed1', borderColor: '#722ed1' }}>
                提交到SRM
              </Button>
            </Popconfirm>
          )}
        </Space>
      </Space>
    );
  };

  // Full config for readonly display (show all categories)
  const fullConfig: BOMControlConfig = useMemo(() => ({
    bom_type: activeTab as 'EBOM' | 'PBOM' | 'MBOM',
    visible_categories: [],
    category_config: {},
  }), [activeTab]);

  return (
    <div>
      {/* Top: Tabs + BOM selector + create */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
        <Text strong style={{ fontSize: isMobile ? 14 : 15 }}>BOM管理</Text>
        <Button type="primary" size={isMobile ? 'small' : undefined} icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>
          新建BOM
        </Button>
      </div>
      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key)}
        size={isMobile ? 'small' : undefined}
        items={isMobile ? [
          { key: 'EBOM', label: 'EBOM' },
          { key: 'PBOM', label: 'PBOM' },
          { key: 'MBOM', label: 'MBOM' },
        ] : [
          { key: 'EBOM', label: 'EBOM 工程BOM' },
          { key: 'PBOM', label: 'PBOM 生产BOM' },
          { key: 'MBOM', label: 'MBOM 制造BOM' },
        ]}
        style={{ marginBottom: 8 }}
      />
      {filteredBomList.length > 1 && (
        <div style={{ marginBottom: 12 }}>
          <Select
            value={selectedBomId || undefined}
            onChange={setSelectedBomId}
            style={{ width: isMobile ? '100%' : 300 }}
            placeholder="选择BOM版本"
            loading={listLoading}
            options={filteredBomList.map(b => ({
              label: `${b.bom_type} ${b.version || '草稿'}${b.status === 'obsolete' ? ' (已废弃)' : b.status === 'released' ? ' (当前)' : ''}`,
              value: b.id,
            }))}
          />
        </div>
      )}

      {/* Create from upstream buttons when no BOM of this type exists */}
      {!listLoading && filteredBomList.length === 0 && activeTab === 'PBOM' && (() => {
        const releasedEbom = bomList.find(b => b.bom_type === 'EBOM' && b.status === 'released');
        return releasedEbom ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Empty description="暂无PBOM" />
            <Button
              type="primary"
              style={{ marginTop: 16 }}
              loading={createFromMutation.isPending}
              onClick={() => createFromMutation.mutate({ sourceBomId: releasedEbom.id, targetType: 'PBOM' })}
            >
              从 EBOM {releasedEbom.version} 创建 PBOM
            </Button>
          </div>
        ) : null;
      })()}
      {!listLoading && filteredBomList.length === 0 && activeTab === 'MBOM' && (() => {
        const releasedPbom = bomList.find(b => b.bom_type === 'PBOM' && b.status === 'released');
        return releasedPbom ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Empty description="暂无MBOM" />
            <Button
              type="primary"
              style={{ marginTop: 16 }}
              loading={createFromMutation.isPending}
              onClick={() => createFromMutation.mutate({ sourceBomId: releasedPbom.id, targetType: 'MBOM' })}
            >
              从 PBOM {releasedPbom.version} 创建 MBOM
            </Button>
          </div>
        ) : null;
      })()}

      {/* Version Info Bar */}
      {bomDetail && !isMobile && (
        <Card size="small" style={{ marginBottom: 12 }} styles={{ body: { padding: '10px 16px' } }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
            <Space size={16}>
              {/* Status + Version */}
              <div>
                <Space size={8}>
                  <Tag color={BOM_STATUS_CONFIG[bomDetail.status]?.color} style={{ fontSize: 13 }}>
                    {BOM_STATUS_CONFIG[bomDetail.status]?.text || bomDetail.status}
                  </Tag>
                  <Text strong style={{ fontSize: 15 }}>
                    {bomDetail.bom_type} {bomDetail.version || '-'}
                  </Text>
                </Space>
                {bomDetail.status === 'released' && bomDetail.released_at && (
                  <div style={{ marginTop: 2 }}>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {bomDetail.creator?.name || ''} 发布于 {dayjs(bomDetail.released_at).format('YYYY-MM-DD HH:mm')}
                    </Text>
                  </div>
                )}
                {bomDetail.source_version && (
                  <div style={{ marginTop: 2 }}>
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      来源: {bomDetail.source_version}
                    </Text>
                  </div>
                )}
              </div>
              {/* Stats */}
              <div style={{ borderLeft: '1px solid #f0f0f0', paddingLeft: 16 }}>
                <Text type="secondary" style={{ fontSize: 12 }}>{isPBOM ? '零件数' : '物料数'}</Text>
                <div><Text strong>{totalItems}</Text></div>
              </div>
              {isPBOM ? (
                <>
                  <div>
                    <Text type="secondary" style={{ fontSize: 12 }}>目标成本</Text>
                    <div><Text strong style={{ color: '#cf1322', fontSize: 16 }}>¥{totalTargetPrice.toFixed(2)}</Text></div>
                  </div>
                  <div>
                    <Text type="secondary" style={{ fontSize: 12 }}>模具费</Text>
                    <div><Text strong style={{ color: '#cf1322' }}>¥{totalTooling.toFixed(2)}</Text></div>
                  </div>
                </>
              ) : (
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>总成本</Text>
                  <div><Text strong style={{ color: '#cf1322', fontSize: 16 }}>¥{totalCost.toFixed(2)}</Text></div>
                </div>
              )}
            </Space>
            <Space>
              {/* Release button for draft */}
              {bomDetail.status === 'draft' && totalItems > 0 && (
                <Button type="primary" onClick={() => setReleaseModalOpen(true)}>
                  发布 {bomDetail.bom_type}
                </Button>
              )}
              {renderActions()}
            </Space>
          </div>
        </Card>
      )}
      {/* Mobile: compact summary line */}
      {bomDetail && isMobile && (
        <div className="bom-mobile-summary">
          <Tag color={BOM_STATUS_CONFIG[bomDetail.status]?.color} style={{ margin: 0 }}>
            {BOM_STATUS_CONFIG[bomDetail.status]?.text || bomDetail.status}
          </Tag>
          <span className="bom-mobile-summary-stat">
            <span className="value">{bomDetail.version || '草稿'}</span>
          </span>
          <span className="bom-mobile-summary-stat">
            <span className="label">{isPBOM ? '零件' : '物料'}</span>
            <span className="value">{totalItems}</span>
          </span>
          <span className="bom-mobile-summary-stat">
            <span className="label">成本</span>
            <span className="cost">¥{(isPBOM ? totalTargetPrice : totalCost).toFixed(0)}</span>
          </span>
        </div>
      )}

      {/* Loading state */}
      {(listLoading || detailLoading) && !bomDetail && (
        <div style={{ textAlign: 'center', padding: 60 }}><Spin /></div>
      )}

      {/* Empty state */}
      {!listLoading && filteredBomList.length === 0 && (
        <Empty description={`暂无${activeTab}，请新建`} style={{ padding: 60 }} />
      )}

      {/* BOM Content: new controls in readonly mode */}
      {bomDetail && (
        <>
          {activeTab === 'EBOM' && (
            <EBOMControl
              config={fullConfig}
              value={items}
              onChange={() => {}}
              readonly
              showMaterialCode
            />
          )}
          {activeTab === 'PBOM' && (
            <PBOMControl
              config={fullConfig}
              value={items}
              onChange={() => {}}
              readonly
              showMaterialCode
            />
          )}
          {activeTab === 'MBOM' && (
            <MBOMControl
              config={fullConfig}
              value={items}
              onChange={() => {}}
              readonly
              showMaterialCode
            />
          )}
        </>
      )}

      {/* Mobile bottom action bar */}
      {isMobile && bomDetail && (
        <div className="bom-mobile-action-bar">
          {(bomDetail.status === 'draft' || bomDetail.status === 'rejected') && (
            <Button type="primary" size="small" icon={<SendOutlined />}
              loading={submitMutation.isPending}
              onClick={() => submitMutation.mutate()}>
              提交审批
            </Button>
          )}
          {bomDetail.status === 'pending_review' && (
            <>
              <Button type="primary" size="small" style={{ background: '#52c41a', borderColor: '#52c41a' }}
                icon={<CheckCircleOutlined />} loading={approveMutation.isPending}
                onClick={() => approveMutation.mutate()}>
                通过
              </Button>
              <Button danger size="small" icon={<CloseCircleOutlined />}
                onClick={() => setRejectModalOpen(true)}>
                驳回
              </Button>
            </>
          )}
          {bomDetail.status === 'draft' && totalItems > 0 && (
            <Button type="primary" size="small" onClick={() => setReleaseModalOpen(true)}>
              发布
            </Button>
          )}
          <Button size="small" icon={<DownloadOutlined />} loading={exportLoading} onClick={handleExportExcel}>
            导出
          </Button>
        </div>
      )}

      {/* Create BOM Modal */}
      <Modal
        title="新建BOM"
        open={createModalOpen}
        onCancel={() => { setCreateModalOpen(false); form.resetFields(); }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending}
      >
        <Form form={form} layout="vertical" onFinish={(values) => createMutation.mutate({ ...values, name: values.bom_type })}>
          <Form.Item name="bom_type" label="BOM类型" initialValue={activeTab} rules={[{ required: true, message: '请选择BOM类型' }]}>
            <Select options={[
              { label: 'EBOM - 工程BOM', value: 'EBOM' },
              { label: 'PBOM - 生产BOM', value: 'PBOM' },
              { label: 'MBOM - 制造BOM', value: 'MBOM' },
            ]} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="BOM描述信息（可选）" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Release BOM Modal */}
      <Modal
        title={`发布 ${bomDetail?.bom_type || ''}`}
        open={releaseModalOpen}
        onCancel={() => { setReleaseModalOpen(false); setReleaseNote(''); }}
        onOk={() => releaseMutation.mutate(releaseNote)}
        confirmLoading={releaseMutation.isPending}
        okText="确认发布"
      >
        <div style={{ marginBottom: 12 }}>
          <Text type="secondary">
            发布后BOM将不可编辑，系统会自动分配版本号。
          </Text>
        </div>
        <Input.TextArea
          rows={3}
          placeholder="请输入发布说明..."
          value={releaseNote}
          onChange={(e) => setReleaseNote(e.target.value)}
        />
      </Modal>

      {/* Reject Modal */}
      <Modal
        title="驳回BOM"
        open={rejectModalOpen}
        onCancel={() => { setRejectModalOpen(false); setRejectComment(''); }}
        onOk={() => rejectMutation.mutate(rejectComment)}
        confirmLoading={rejectMutation.isPending}
        okText="确认驳回"
        okButtonProps={{ danger: true }}
      >
        <Input.TextArea
          rows={4}
          placeholder="请输入驳回原因..."
          value={rejectComment}
          onChange={(e) => setRejectComment(e.target.value)}
        />
      </Modal>

      {/* Material Search Modal */}
      <MaterialSearchModal
        open={materialModalOpen}
        onClose={() => { setMaterialModalOpen(false); setEditingRowId(null); }}
        onSelect={handleMaterialSelect}
      />

      {/* Compare BOMs Modal */}
      <Modal
        title="BOM版本对比"
        open={compareModalOpen}
        onCancel={() => setCompareModalOpen(false)}
        width={800}
        footer={compareResult ? [
          <Button key="close" onClick={() => setCompareModalOpen(false)}>关闭</Button>,
        ] : undefined}
        onOk={handleCompare}
        confirmLoading={compareLoading}
        okText="开始对比"
      >
        <Space style={{ marginBottom: 16, width: '100%' }} direction="vertical">
          <Space>
            <Text>BOM A：</Text>
            <Select
              style={{ width: 280 }}
              placeholder="选择第一个BOM"
              value={compareBom1}
              onChange={setCompareBom1}
              options={bomList.map(b => ({
                label: `${b.bom_type} ${b.version || '草稿'}`,
                value: b.id,
              }))}
            />
          </Space>
          <Space>
            <Text>BOM B：</Text>
            <Select
              style={{ width: 280 }}
              placeholder="选择第二个BOM"
              value={compareBom2}
              onChange={setCompareBom2}
              options={bomList.map(b => ({
                label: `${b.bom_type} ${b.version || '草稿'}`,
                value: b.id,
              }))}
            />
          </Space>
        </Space>
        {compareResult && (
          <Table
            dataSource={compareResult}
            rowKey={(_, idx) => String(idx)}
            size="small"
            pagination={false}
            scroll={{ y: 400 }}
            rowClassName={(record) => {
              if (record.change_type === 'added') return 'compare-row-added';
              if (record.change_type === 'removed') return 'compare-row-removed';
              if (record.change_type === 'changed') return 'compare-row-changed';
              return '';
            }}
            columns={[
              { title: '序号', width: 60, render: (_, __, idx) => idx + 1 },
              { title: '物料名称', dataIndex: 'name', width: 140 },
              { title: '规格', dataIndex: 'specification', width: 160, ellipsis: true },
              { title: '变更类型', dataIndex: 'change_type', width: 100,
                render: (v: string) => {
                  const map: Record<string, { color: string; text: string }> = {
                    added: { color: 'success', text: '新增' },
                    removed: { color: 'error', text: '删除' },
                    changed: { color: 'warning', text: '变更' },
                    unchanged: { color: 'default', text: '未变' },
                  };
                  const cfg = map[v] || { color: 'default', text: v };
                  return <Tag color={cfg.color}>{cfg.text}</Tag>;
                },
              },
              { title: '变更详情', dataIndex: 'details', ellipsis: true },
            ]}
          />
        )}
        <style>{`
          .compare-row-added { background: #f6ffed !important; }
          .compare-row-added:hover > td { background: #d9f7be !important; }
          .compare-row-removed { background: #fff1f0 !important; }
          .compare-row-removed:hover > td { background: #ffccc7 !important; }
          .compare-row-changed { background: #fffbe6 !important; }
          .compare-row-changed:hover > td { background: #fff1b8 !important; }
        `}</style>
      </Modal>

      {/* 图纸上传Modal */}
      <Modal
        title={`上传${drawingUploadType}图纸新版本`}
        open={drawingUploadModalOpen}
        onCancel={() => { setDrawingUploadModalOpen(false); setDrawingChangeDesc(''); }}
        footer={null}
        width={400}
      >
        <div style={{ marginBottom: 12 }}>
          <Text type="secondary">变更说明（可选）</Text>
          <Input.TextArea
            rows={2}
            value={drawingChangeDesc}
            onChange={(e) => setDrawingChangeDesc(e.target.value)}
            placeholder="描述本次变更内容..."
            style={{ marginTop: 4 }}
          />
        </div>
        <Upload
          showUploadList={false}
          beforeUpload={handleDrawingVersionUpload}
        >
          <Button icon={<UploadOutlined />} type="primary">选择文件并上传</Button>
        </Upload>
      </Modal>

      {/* 图纸版本历史Drawer */}
      <Drawer
        title={`${drawingHistoryType}图纸版本历史`}
        open={drawingHistoryOpen}
        onClose={() => setDrawingHistoryOpen(false)}
        width={480}
      >
        {(() => {
          const itemDrawings = drawingsByBOM[drawingHistoryItemId];
          const list = itemDrawings?.[drawingHistoryType] || [];
          if (list.length === 0) return <Empty description="暂无图纸版本" />;
          return (
            <Timeline
              items={list.map((d: PartDrawing) => ({
                key: d.id,
                color: d === list[0] ? 'blue' : 'gray',
                children: (
                  <div>
                    <Space>
                      <Tag color={d === list[0] ? 'blue' : 'default'}>{d.version}</Tag>
                      <a href={d.file_url} target="_blank" rel="noreferrer">{d.file_name}</a>
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        {d.file_size ? `${(d.file_size / 1024).toFixed(0)}KB` : ''}
                      </Text>
                    </Space>
                    {d.change_description && (
                      <div style={{ marginTop: 4 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>{d.change_description}</Text>
                      </div>
                    )}
                    <div style={{ marginTop: 2 }}>
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        {d.uploader?.name || '未知'} {dayjs(d.created_at).format('MM-DD HH:mm')}
                      </Text>
                    </div>
                  </div>
                ),
              }))}
            />
          );
        })()}
      </Drawer>
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

// ============ Task Actions Component ============


const TaskActions: React.FC<{
  task: Task;
  projectId: string;
  onRefresh: () => void;
}> = ({ task, projectId, onRefresh }) => {
  const [assignModalOpen, setAssignModalOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [assigneeId, setAssigneeId] = useState('');

  const handleError = (err: unknown) => {
    const axiosErr = err as any;
    const status = axiosErr?.response?.status;
    const errMsg = axiosErr?.response?.data?.error || axiosErr?.response?.data?.message || '操作失败';
    if (status === 400) {
      message.error(`前置任务未完成，${errMsg}`);
    } else {
      message.error(errMsg);
    }
  };

  const handleAssign = async () => {
    if (!assigneeId.trim()) {
      message.warning('请选择负责人');
      return;
    }
    setLoading(true);
    try {
      await apiClient.put(`/projects/${projectId}/tasks/${task.id}`, {
        assignee_id: assigneeId.trim(),
      });
      message.success('指派成功');
      setAssignModalOpen(false);
      setAssigneeId('');
      onRefresh();
    } catch (err) {
      handleError(err);
    } finally {
      setLoading(false);
    }
  };

  const handleStart = async () => {
    setLoading(true);
    try {
      await projectApi.updateTaskStatus(projectId, task.id, 'in_progress');
      message.success('任务已开始');
      onRefresh();
    } catch (err) {
      handleError(err);
    } finally {
      setLoading(false);
    }
  };

  const renderActions = () => {
    switch (task.status) {
      case 'unassigned':
        return (
          <Button size="small" type="primary" onClick={() => setAssignModalOpen(true)} loading={loading}>
            指派
          </Button>
        );
      case 'pending': {
        const hasUnfinishedDeps = task.dependencies?.some(
          d => d.depends_on_status !== 'completed'
        );
        if (hasUnfinishedDeps) {
          return <Tag color="default" icon={<ClockCircleOutlined />}>等待前置任务</Tag>;
        }
        return (
          <Button size="small" type="primary" style={{ background: '#52c41a', borderColor: '#52c41a' }} onClick={handleStart} loading={loading}>
            开始
          </Button>
        );
      }
      case 'in_progress':
        return <Tag color="processing" icon={<ClockCircleOutlined />}>进行中</Tag>;
      case 'submitted':
        return <Tag color="orange" icon={<CheckCircleOutlined />}>已提交</Tag>;
      case 'reviewing':
        return <Tag color="warning" icon={<AuditOutlined />}>审批中</Tag>;
      case 'completed':
        return <Tag color="green" icon={<CheckCircleOutlined />}>已完成</Tag>;
      case 'rejected':
        return (
          <Button size="small" style={{ color: '#fa8c16', borderColor: '#fa8c16' }} onClick={handleStart} loading={loading}>
            重新开始
          </Button>
        );
      default:
        return null;
    }
  };

  return (
    <>
      <Space size={4}>
        {renderActions()}
      </Space>

      {/* Assign Modal */}
      <Modal
        title={`指派任务: ${task.title}`}
        open={assignModalOpen}
        onCancel={() => { setAssignModalOpen(false); setAssigneeId(''); }}
        onOk={handleAssign}
        confirmLoading={loading}
        okText="确认指派"
        cancelText="取消"
      >
        <div style={{ marginBottom: 16 }}>
          <Text type="secondary" style={{ display: 'block', marginBottom: 4 }}>负责人 *</Text>
          <UserSelect
            value={assigneeId || undefined}
            onChange={(val) => setAssigneeId(val as string)}
            mode="single"
            placeholder="选择负责人"
            style={{ width: '100%' }}
          />
        </div>
      </Modal>
    </>
  );
};

// ============ Role Assignment Tab ============

// ROLE_CODES imported from @/constants/roles

const RoleAssignmentTab: React.FC<{ projectId: string }> = ({ projectId }) => {
  const queryClient = useQueryClient();
  const [assignments, setAssignments] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);

  // Fetch tasks to extract unique default_assignee_role values
  const { data: tasks = [] } = useQuery({
    queryKey: ['project-tasks', projectId],
    queryFn: () => projectApi.listTasks(projectId),
    enabled: !!projectId,
  });

  // Fetch task roles for label lookup
  const { data: taskRolesData = [] } = useQuery<TaskRole[]>({
    queryKey: ['task-roles'],
    queryFn: () => taskRoleApi.list(),
  });

  const roleLabelMap = useMemo(() => {
    const map: Record<string, string> = {};
    for (const r of taskRolesData) {
      map[r.code] = r.name;
    }
    for (const rc of ROLE_CODES) {
      if (!map[rc.code]) map[rc.code] = rc.label;
    }
    return map;
  }, [taskRolesData]);

  // Extract unique roles from tasks
  const uniqueRoles = useMemo(() => {
    const roles = new Set<string>();
    for (const t of tasks) {
      const role = (t as any).default_assignee_role;
      if (role) roles.add(role);
    }
    return Array.from(roles).sort();
  }, [tasks]);

  const updateAssignment = (roleCode: string, userId: string) => {
    setAssignments(prev => ({ ...prev, [roleCode]: userId }));
  };

  const handleSave = async () => {
    const validAssignments = Object.entries(assignments)
      .filter(([, userId]) => userId && userId.trim())
      .map(([role, userId]) => ({ role, user_id: userId.trim() }));

    if (validAssignments.length === 0) {
      message.warning('请至少填写一个角色的负责人');
      return;
    }

    setLoading(true);
    try {
      await projectApi.assignRoles(projectId, validAssignments);
      message.success('角色分配成功，已更新对应任务的负责人');
      queryClient.invalidateQueries({ queryKey: ['project-tasks', projectId] });
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
    } catch (err) {
      const axiosErr = err as any;
      const errMsg = axiosErr?.response?.data?.message || '分配失败';
      message.error(errMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Text strong>角色分配</Text>
        <Button type="primary" onClick={handleSave} loading={loading}>
          保存并更新任务
        </Button>
      </div>

      <Alert
        type="info"
        showIcon
        message="为每个角色指定负责人后，将自动更新该角色下所有任务的负责人"
        style={{ marginBottom: 16 }}
      />

      {uniqueRoles.length === 0 ? (
        <Empty description="项目任务中未配置角色，请在研发流程模板中为任务分配角色" />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {uniqueRoles.map(role => (
            <Card key={role} size="small" styles={{ body: { padding: '12px 16px' } }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <div style={{ width: 140, fontWeight: 500 }}>
                  {roleLabelMap[role] || role}
                </div>
                <Tag color="blue">{role}</Tag>
                <UserSelect
                  value={assignments[role] || undefined}
                  onChange={(val) => updateAssignment(role, val as string)}
                  mode="single"
                  placeholder="选择负责人"
                  style={{ flex: 1 }}
                />
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};

// ============ SKU Management Tab ============

// 色块组件
const ColorSwatch: React.FC<{ hex?: string; size?: number }> = ({ hex, size = 14 }) => {
  if (!hex) return null;
  return (
    <span style={{
      display: 'inline-block', width: size, height: size, borderRadius: 3,
      backgroundColor: hex, border: '1px solid #d9d9d9', verticalAlign: 'middle',
    }} />
  );
};

const SKUTab: React.FC<{ projectId: string }> = ({ projectId }) => {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [selectedSKU, setSelectedSKU] = useState<ProductSKU | null>(null);
  const [form] = Form.useForm();

  // 创建弹窗：非外观件勾选状态 + 外观件CMF变体选择
  const [checkedNonAppearance, setCheckedNonAppearance] = useState<Set<string>>(new Set());
  const [selectedVariants, setSelectedVariants] = useState<Record<string, string>>({}); // bomItemId -> variantId

  // List SKUs
  const { data: skus = [], isLoading } = useQuery<ProductSKU[]>({
    queryKey: ['project-skus', projectId],
    queryFn: () => skuApi.listSKUs(projectId),
  });

  // Get PBOM items for create modal
  const { data: bomItems = [] } = useQuery({
    queryKey: ['project-pbom-items', projectId],
    queryFn: async () => {
      const boms = await projectBomApi.list(projectId, { bom_type: 'PBOM' });
      if (boms.length === 0) return [];
      const detail = await projectBomApi.get(projectId, boms[0].id);
      return detail.items || [];
    },
    enabled: createOpen,
  });

  // Get appearance parts + CMF variants for create modal
  const { data: appearanceParts = [] } = useQuery<AppearancePartWithCMF[]>({
    queryKey: ['appearance-parts', projectId],
    queryFn: () => cmfVariantApi.getAppearanceParts(projectId),
    enabled: createOpen,
  });

  // Get full BOM for detail view
  const { data: fullBom = [], isLoading: fullBomLoading } = useQuery<FullBOMItem[]>({
    queryKey: ['sku-full-bom', projectId, selectedSKU?.id],
    queryFn: () => skuApi.getFullBOM(projectId, selectedSKU!.id),
    enabled: !!selectedSKU,
  });

  // Split BOM items into non-appearance and appearance
  const nonAppearanceItems = useMemo(() =>
    bomItems.filter((item: ProjectBOMItem) => !item.extended_attrs?.is_appearance_part && !item.extended_attrs?.is_variant),
  [bomItems]);

  // Initialize non-appearance checkboxes when modal opens
  useEffect(() => {
    if (createOpen && nonAppearanceItems.length > 0) {
      setCheckedNonAppearance(new Set(nonAppearanceItems.map((i: ProjectBOMItem) => i.id)));
      setSelectedVariants({});
    }
  }, [createOpen, nonAppearanceItems]);

  // Create SKU mutation
  const createMutation = useMutation({
    mutationFn: (data: { name: string; bom_items: Array<{ bom_item_id: string; cmf_variant_id?: string }> }) =>
      skuApi.createSKU(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-skus', projectId] });
      setCreateOpen(false);
      form.resetFields();
      message.success('SKU创建成功');
    },
    onError: () => message.error('创建失败'),
  });

  // Delete SKU mutation
  const deleteMutation = useMutation({
    mutationFn: (skuId: string) => skuApi.deleteSKU(projectId, skuId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-skus', projectId] });
      if (selectedSKU) setSelectedSKU(null);
      message.success('已删除');
    },
  });

  // Handle create submit
  const handleCreateSubmit = (values: { name: string }) => {
    const bomItemsToSave: Array<{ bom_item_id: string; cmf_variant_id?: string }> = [];

    // Add checked non-appearance items
    for (const id of checkedNonAppearance) {
      bomItemsToSave.push({ bom_item_id: id });
    }

    // Add appearance items with selected CMF variant
    for (const [bomItemId, variantId] of Object.entries(selectedVariants)) {
      if (variantId) {
        bomItemsToSave.push({ bom_item_id: bomItemId, cmf_variant_id: variantId });
      }
    }

    createMutation.mutate({ name: values.name, bom_items: bomItemsToSave });
  };

  // ========== SKU Detail View ==========
  if (selectedSKU) {
    const detailColumns = [
      { title: '序号', dataIndex: 'item_number', width: 60, align: 'center' as const },
      { title: '零件名称', dataIndex: 'name', width: 160 },
      { title: '材质', dataIndex: 'material_type', width: 100 },
      { title: '数量', dataIndex: 'quantity', width: 70, align: 'right' as const },
      { title: '单位', dataIndex: 'unit', width: 60 },
      {
        title: 'CMF信息',
        width: 280,
        render: (_: any, record: FullBOMItem) => {
          if (!record.is_appearance_part || !record.cmf_variant) return '-';
          const v = record.cmf_variant;
          return (
            <Space size={6}>
              <ColorSwatch hex={v.color_hex} />
              {v.material_code && <Tag style={{ fontSize: 11 }}>{v.material_code}</Tag>}
              {v.finish && <Text style={{ fontSize: 12 }}>{v.finish}</Text>}
              {v.texture && <Text style={{ fontSize: 12 }}>{v.texture}</Text>}
            </Space>
          );
        },
      },
    ];

    return (
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
          <Button size="small" onClick={() => setSelectedSKU(null)}>&lt; 返回</Button>
          <Text strong style={{ fontSize: 15 }}>{selectedSKU.name}</Text>
          {selectedSKU.code && <Tag>{selectedSKU.code}</Tag>}
        </div>

        {fullBomLoading ? <Spin /> : fullBom.length === 0 ? (
          <Empty description="该SKU暂无BOM零件" />
        ) : (
          <Table
            columns={detailColumns}
            dataSource={fullBom}
            rowKey="id"
            size="small"
            pagination={false}
            scroll={{ x: 700 }}
          />
        )}
      </div>
    );
  }

  // ========== SKU List View ==========
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Text strong style={{ fontSize: 15 }}>配色方案 / SKU</Text>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
          新建SKU
        </Button>
      </div>

      {isLoading ? <Spin /> : skus.length === 0 ? (
        <Empty description={'暂无SKU，点击"新建SKU"开始'} />
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: 12 }}>
          {skus.map(sku => (
            <Card
              key={sku.id}
              size="small"
              hoverable
              onClick={() => setSelectedSKU(sku)}
              styles={{ body: { padding: '12px 16px' } }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <Text strong>{sku.name}</Text>
                  {sku.code && <Tag style={{ marginLeft: 8 }}>{sku.code}</Tag>}
                </div>
                <Tag color={sku.status === 'active' ? 'green' : 'default'}>{sku.status === 'active' ? '启用' : '停用'}</Tag>
              </div>
              {sku.description && <Text type="secondary" style={{ fontSize: 12, marginTop: 4, display: 'block' }}>{sku.description}</Text>}
              <div style={{ marginTop: 8, display: 'flex', justifyContent: 'flex-end' }}>
                <Popconfirm title="确认删除此SKU？" onConfirm={(e) => { e?.stopPropagation(); deleteMutation.mutate(sku.id); }}>
                  <Button size="small" type="text" danger icon={<DeleteOutlined />} onClick={(e) => e.stopPropagation()} />
                </Popconfirm>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Create SKU Modal */}
      <Modal
        title="新建SKU"
        open={createOpen}
        onCancel={() => { setCreateOpen(false); form.resetFields(); }}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending}
        width={700}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入SKU名称' }]}>
            <Input placeholder="如：星空黑、冰川白" />
          </Form.Item>
        </Form>

        {/* 非外观件列表 */}
        {nonAppearanceItems.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            <Text strong style={{ fontSize: 13, marginBottom: 8, display: 'block' }}>非外观件（通用零件）</Text>
            <div style={{ maxHeight: 200, overflow: 'auto', border: '1px solid #f0f0f0', borderRadius: 6, padding: 8 }}>
              {nonAppearanceItems.map((item: ProjectBOMItem) => (
                <div key={item.id} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 0' }}>
                  <Checkbox
                    checked={checkedNonAppearance.has(item.id)}
                    onChange={(e) => {
                      const next = new Set(checkedNonAppearance);
                      if (e.target.checked) next.add(item.id); else next.delete(item.id);
                      setCheckedNonAppearance(next);
                    }}
                  />
                  <Text style={{ fontSize: 13 }}>#{item.item_number} {item.name}</Text>
                  {item.extended_attrs?.material_type && <Tag style={{ fontSize: 11 }}>{item.extended_attrs.material_type}</Tag>}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 外观件 + CMF变体选择 */}
        {appearanceParts.length > 0 && (
          <div>
            <Text strong style={{ fontSize: 13, marginBottom: 8, display: 'block' }}>外观件（选择CMF方案）</Text>
            <div style={{ border: '1px solid #f0f0f0', borderRadius: 6, padding: 8 }}>
              {appearanceParts.map((part) => {
                const item = part.bom_item;
                const variants = part.cmf_variants || [];
                return (
                  <div key={item.id} style={{ marginBottom: 12 }}>
                    <Text strong style={{ fontSize: 13 }}>#{item.item_number} {item.name}</Text>
                    {item.extended_attrs?.material_type && <Tag style={{ fontSize: 11, marginLeft: 6 }}>{item.extended_attrs.material_type}</Tag>}
                    {variants.length === 0 ? (
                      <div style={{ padding: '4px 0', color: '#999', fontSize: 12 }}>暂无CMF方案</div>
                    ) : (
                      <Radio.Group
                        value={selectedVariants[item.id] || ''}
                        onChange={(e) => setSelectedVariants(prev => ({ ...prev, [item.id]: e.target.value }))}
                        style={{ display: 'flex', flexDirection: 'column', gap: 4, marginTop: 4 }}
                      >
                        {variants.map((v: CMFVariant) => (
                          <Radio key={v.id} value={v.id} style={{ fontSize: 12 }}>
                            <Space size={6}>
                              <Tag color="processing" style={{ margin: 0, fontSize: 11 }}>V{v.variant_index}</Tag>
                              {v.material_code && <Text style={{ fontSize: 11, color: '#8c8c8c' }}>{v.material_code}</Text>}
                              <ColorSwatch hex={v.color_hex} />
                              {v.finish && <Text style={{ fontSize: 11 }}>{v.finish}</Text>}
                              {v.texture && <Text style={{ fontSize: 11 }}>{v.texture}</Text>}
                            </Space>
                          </Radio>
                        ))}
                      </Radio.Group>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

// ============ Main ProjectDetail Page ============

const ProjectDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isMobileView = useIsMobile();

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
      projectApi.updateTaskStatus(projectId, taskId, 'completed'),
    onSuccess: () => {
      message.success('任务已完成');
      queryClient.invalidateQueries({ queryKey: ['project-tasks', id] });
    },
    onError: () => message.error('操作失败'),
  });

  const refreshTasks = () => {
    queryClient.invalidateQueries({ queryKey: ['project-tasks', id] });
    queryClient.invalidateQueries({ queryKey: ['project', id] });
  };

  // SSE: 实时推送自动刷新
  useSSE({
    onTaskUpdate: useCallback((event: SSETaskEvent) => {
      if (event.project_id === id) {
        queryClient.invalidateQueries({ queryKey: ['project-tasks', id] });
        queryClient.invalidateQueries({ queryKey: ['project', id] });
      }
    }, [id, queryClient]),
    onProjectUpdate: useCallback((event: SSETaskEvent) => {
      if (event.project_id === id) {
        queryClient.invalidateQueries({ queryKey: ['project-tasks', id] });
        queryClient.invalidateQueries({ queryKey: ['project', id] });
      }
    }, [id, queryClient]),
    enabled: !!id,
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
    <div style={{ padding: isMobileView ? 12 : 24 }}>
      {/* Header */}
      {!isMobileView && (
        <div style={{ marginBottom: 24 }}>
          <Button type="link" icon={<ArrowLeftOutlined />} onClick={() => navigate('/projects')} style={{ padding: 0, marginBottom: 8 }}>
            返回项目列表
          </Button>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div style={{ minWidth: 0, flex: 1 }}>
              <Title level={3} style={{ margin: 0 }}>
                {project.name}
                {project.code && <Text code style={{ marginLeft: 8, fontSize: 14 }}>{project.code}</Text>}
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
      )}

      {/* Tabs */}
      <Card bodyStyle={{ padding: isMobileView ? 8 : undefined }}>
        <Tabs
          defaultActiveKey="overview"
          tabBarGutter={isMobileView ? 8 : undefined}
          size={isMobileView ? 'small' : undefined}
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
                    onRefresh={refreshTasks}
                  />
                </div>
              ) : (
                <Empty description="暂无任务" />
              ),
            },
            {
              key: 'bom',
              label: 'BOM管理',
              children: <BOMTab projectId={project.id} />,
            },
            {
              key: 'sku',
              label: 'SKU配色',
              children: <SKUTab projectId={project.id} />,
            },
            {
              key: 'cmf',
              label: 'CMF配色',
              children: <CMFEditControl projectId={project.id} readonly />,
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
            {
              key: 'roles',
              label: '角色指派',
              children: <RoleAssignmentTab projectId={project.id} />,
            },
          ]}
        />
      </Card>
    </div>
  );
};

export default ProjectDetail;
