# PLM Frontend Phase 4 - Workflow UI

## Task 1: Create src/api/workflow.ts

Create the file `src/api/workflow.ts` with this exact content:

```typescript
import apiClient from './client';

export interface TaskActionLog {
  id: string;
  action: string;
  from_status: string;
  to_status: string;
  operator_id: string;
  operator_type: string;
  comment?: string;
  event_data?: any;
  created_at: string;
}

export interface RoleAssignment {
  role_code: string;
  user_id: string;
  feishu_user_id?: string;
}

export interface ReviewOutcome {
  outcome_code: string;
  outcome_name: string;
  outcome_type: string;
  rollback_to_task_code?: string;
}

export const workflowApi = {
  assignTask: async (projectId: string, taskId: string, data: { assignee_id: string; feishu_user_id?: string }) => {
    const response = await apiClient.post(`/projects/${projectId}/tasks/${taskId}/assign`, data);
    return response.data;
  },

  startTask: async (projectId: string, taskId: string) => {
    const response = await apiClient.post(`/projects/${projectId}/tasks/${taskId}/start`);
    return response.data;
  },

  completeTask: async (projectId: string, taskId: string) => {
    const response = await apiClient.post(`/projects/${projectId}/tasks/${taskId}/complete`);
    return response.data;
  },

  submitReview: async (projectId: string, taskId: string, data: { outcome_code: string; comment?: string }) => {
    const response = await apiClient.post(`/projects/${projectId}/tasks/${taskId}/review`, data);
    return response.data;
  },

  assignPhaseRoles: async (projectId: string, phase: string, data: { assignments: RoleAssignment[] }) => {
    const response = await apiClient.post(`/projects/${projectId}/phases/${phase}/assign-roles`, data);
    return response.data;
  },

  getTaskHistory: async (projectId: string, taskId: string): Promise<TaskActionLog[]> => {
    const response = await apiClient.get(`/projects/${projectId}/tasks/${taskId}/history`);
    return response.data.data || [];
  },
};
```

## Task 2: Modify src/pages/ProjectDetail.tsx

### 2a. Add imports

Add these imports at the top:
- `Drawer, Timeline` from 'antd'
- `UserAddOutlined, AuditOutlined, CloseCircleOutlined, HistoryOutlined` from '@ant-design/icons'
- `import { workflowApi, TaskActionLog } from '@/api/workflow';`

### 2b. Update taskStatusConfig

Replace the existing `taskStatusConfig` object with:

```typescript
const taskStatusConfig: Record<string, { color: string; text: string; icon: React.ReactNode; barColor: string }> = {
  unassigned: { color: 'default', text: '待指派', icon: <UserAddOutlined />, barColor: '#d9d9d9' },
  pending: { color: 'default', text: '待处理', icon: <ClockCircleOutlined />, barColor: '#bfbfbf' },
  in_progress: { color: 'processing', text: '进行中', icon: <ClockCircleOutlined />, barColor: '#1677ff' },
  reviewing: { color: 'warning', text: '审批中', icon: <AuditOutlined />, barColor: '#faad14' },
  completed: { color: 'success', text: '已完成', icon: <CheckCircleOutlined />, barColor: '#52c41a' },
  rejected: { color: 'error', text: '已驳回', icon: <CloseCircleOutlined />, barColor: '#ff4d4f' },
  ready: { color: 'blue', text: '就绪', icon: <PlayCircleOutlined />, barColor: '#69b1ff' },
  blocked: { color: 'error', text: '阻塞', icon: <ExclamationCircleOutlined />, barColor: '#ff4d4f' },
  needs_review: { color: 'warning', text: '待审批', icon: <ExclamationCircleOutlined />, barColor: '#faad14' },
};
```

### 2c. Add a TaskActions component (inside ProjectDetail.tsx, before the main ProjectDetail component)

Create a component called `TaskActions` that takes props: `{ task: Task; projectId: string; onRefresh: () => void }`.

It should manage state for:
- `assignModalOpen` (boolean)
- `rejectModalOpen` (boolean) 
- `historyDrawerOpen` (boolean)
- `historyData` (TaskActionLog[])
- `historyLoading` (boolean)
- `loading` (boolean - for action buttons)
- `assigneeId` (string)
- `rejectComment` (string)

Render action buttons based on task.status:
- **unassigned**: Blue "指派" button → opens assign modal
- **pending**: Green "开始" button → calls workflowApi.startTask, then onRefresh
- **in_progress**: Green "完成" button → calls workflowApi.completeTask, if response.data.status === 'reviewing', show message.info('任务已提交审批'), then onRefresh
- **reviewing**: Two buttons - green "通过" (calls workflowApi.submitReview with outcome_code:'pass') and red "驳回" (opens reject modal)
- **completed**: Show a green CheckCircleOutlined icon with "已完成" text
- **rejected**: Orange "重新开始" button → calls workflowApi.startTask, then onRefresh

Plus a small grey "历史" button (HistoryOutlined) that loads history and opens a Drawer.

All API calls in try-catch. On error, extract error message: `(err as any)?.response?.data?.error || (err as any)?.response?.data?.message || '操作失败'`. If status 400, prepend "前置任务未完成，" to the message.

The Assign Modal should have an Input for assignee_id and optional feishu_user_id.
The Reject Modal should have a TextArea for comment.
The History Drawer should use Timeline component showing each log entry with: action name, from_status → to_status, operator, time, comment.

Use `Space size={4}` or `Space wrap` for compact button layout. Buttons should be size="small".

### 2d. Add a RoleAssignmentTab component

Create a component `RoleAssignmentTab` that takes `{ projectId: string }`.

It should display a simple form where users can:
- Select a phase from a dropdown (concept, evt, dvt, pvt, mp)
- Add role assignments (role_code + user_id pairs)
- Submit via workflowApi.assignPhaseRoles

Implementation:
- State: selectedPhase, assignments array [{role_code, user_id}], loading
- Predefined role codes: ['project_manager', 'hardware_engineer', 'software_engineer', 'mechanical_engineer', 'quality_engineer', 'reviewer']
- Display a table-like form with rows for each role
- "保存" button to submit

### 2e. Add TaskActions column to the GanttChart left panel

In the GanttChart component's left panel header, add a new "操作" column (width: 120px).

In each task row, add the TaskActions component. Pass the task, projectId, and a refresh function.

The refresh function should call `queryClient.invalidateQueries({ queryKey: ['project-tasks', projectId] })`. Since GanttChart doesn't have access to queryClient, pass `onRefresh` as a prop to GanttChart.

Update GanttChart props to include `onRefresh: () => void`.

### 2f. Add the "角色指派" tab

In the main ProjectDetail component's Tabs items array, add a new tab:
```
{
  key: 'roles',
  label: '角色指派',
  children: <RoleAssignmentTab projectId={project.id} />,
}
```

### 2g. Wire up onRefresh in ProjectDetail

In the main ProjectDetail component, create:
```typescript
const refreshTasks = () => {
  queryClient.invalidateQueries({ queryKey: ['project-tasks', id] });
};
```

Pass this to GanttChart as onRefresh prop.

## Task 3: Build and Deploy

```bash
cd /home/claw/.openclaw/workspace/nimo-plm-web
npm run build
```

If build succeeds:
```bash
cp dist/index.html /home/claw/.openclaw/workspace/web/plm/
rm -rf /home/claw/.openclaw/workspace/web/plm/assets
cp -r dist/assets /home/claw/.openclaw/workspace/web/plm/assets
```

## Important Notes

- Keep ALL existing functionality intact (Gantt chart, BOM, Documents, etc.)
- Use `@/` alias for imports (maps to src/)
- Buttons should be compact (size="small", Space compact)
- All error handling should be user-friendly with message.error()
- The existing `completeTaskMutation` in ProjectDetail can stay but TaskActions will use workflowApi directly
