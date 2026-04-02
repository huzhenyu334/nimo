export type ViewMode = 'edit' | 'runtime' | 'debug';

export interface ExecutionPathEntry {
  step: string;
  status: string;
  duration_ms?: number;
  started_at?: number;
  parent_step_id?: string;
  iteration_type?: 'foreach' | 'loop';
  iteration_index?: number;
  iteration_item?: string;
}
