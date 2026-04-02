/**
 * FlowGantt — Unified gantt/timeline grid: CallStack (left) + Timeline (right).
 * Tree-based: step → iterations → body steps, with recursive nesting.
 * Supports mode: 'edit' | 'runtime' | 'debug'.
 */

import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { Typography } from 'antd';
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  LoadingOutlined,
  MinusCircleOutlined,
  PauseCircleOutlined,
  PushpinOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import type { ExecutionPathEntry, ViewMode } from './types';

const { Text } = Typography;

// ── Props ──

export interface FlowGanttProps {
  mode: ViewMode;
  executionPath: ExecutionPathEntry[];
  allSteps: string[];
  bodyStepMap?: Record<string, string[]>;
  /** Map of parentId → { branchKey → stepId[] } for if/switch branches */
  branchStepMap?: Record<string, Record<string, string[]>>;
  /** Map of stepId → control type (loop/foreach/if/switch) */
  controlTypeMap?: Record<string, string>;
  currentStep: string | null;
  selectedStep: string | null;
  pinnedSteps: Set<string>;
  onSelectStep: (stepId: string) => void;
  debugStatus?: string;
  breakpointSteps?: Set<string>;
  onToggleBreakpoint?: (stepId: string) => void;
  /** Map of step ID → planned duration in ms (from YAML planned_duration field) */
  stepPlannedDurations?: Record<string, number>;
  /** Map of stepId → { name, executor, depends_on } for display labels, executor icons, and dependency ordering */
  stepInfoMap?: Record<string, { name?: string; executor?: string; depends_on?: string[] }>;
  /** Optional callback to resolve executor metadata for display */
  getExecutorMeta?: (type: string) => { label: string; icon: string } | undefined;
}

// ── Tree node ──

interface GridNode {
  id: string;
  label: string;
  status: string;
  started_at?: number;
  duration_ms?: number;
  depth: number;
  isIteration?: boolean;
  isBody?: boolean;
  isBranch?: boolean;
  controlType?: string;
  branchKey?: string;
  executorType?: string;
  /** Static step ID for breakpoint operations (body steps may have scoped runtime IDs) */
  breakpointId?: string;
  children: GridNode[];
}

// ── Constants ──

const ROW_HEIGHT = 34;
const INDENT_PX = 16;
const CALL_STACK_WIDTH = 220;
const MAX_VISIBLE_ITERATIONS = 10;
const TRUNCATED_VISIBLE = 5;
const RULER_HEIGHT = 40;
const RULER_UPPER_H = 20;
const RULER_LOWER_H = 20;
const MIN_DURATION = 10_000;        // 10 s
const MAX_DURATION = 63_072_000_000; // ~2 years
const ZOOM_FACTOR = 1.15;

// ── Status helpers ──

const STATUS_ICONS: Record<string, React.ReactNode> = {
  completed: <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 12 }} />,
  running: <LoadingOutlined style={{ color: '#1677ff', fontSize: 12 }} spin />,
  paused_before: <PauseCircleOutlined style={{ color: '#1677ff', fontSize: 12 }} />,
  paused_after: <PauseCircleOutlined style={{ color: '#722ed1', fontSize: 12 }} />,
  paused: <PauseCircleOutlined style={{ color: '#1677ff', fontSize: 12 }} />,
  skipped: <MinusCircleOutlined style={{ color: '#8c8c8c', fontSize: 12 }} />,
  failed: <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 12 }} />,
  waiting: <ClockCircleOutlined style={{ color: '#434343', fontSize: 12 }} />,
};

const STATUS_COLORS: Record<string, { bar: string; text: string }> = {
  completed:     { bar: '#52c41a', text: '#b7eb8f' },
  running:       { bar: '#1677ff', text: '#91caff' },
  paused_before: { bar: '#1677ff', text: '#91caff' },
  paused_after:  { bar: '#1677ff', text: '#91caff' },
  paused:        { bar: '#1677ff', text: '#91caff' },
  skipped:       { bar: '#434343', text: '#8c8c8c' },
  failed:        { bar: '#ff4d4f', text: '#ffa39e' },
  cancelled:     { bar: '#595959', text: '#8c8c8c' },
  waiting:       { bar: '#303030', text: '#595959' },
};

function getStatusColors(s: string) {
  return STATUS_COLORS[s] ?? STATUS_COLORS.waiting;
}

function formatDuration(ms?: number): string {
  if (ms == null) return '';
  const abs = Math.abs(ms);
  if (abs < 1000) return `${abs}ms`;
  if (abs < 60000) return `${(abs / 1000).toFixed(1)}s`;
  return `${Math.floor(abs / 60000)}m${Math.round((abs % 60000) / 1000)}s`;
}

// ── Tick system ──

interface TickCfg { upper: string; step: number | 'month'; fmt: string }

const TICK_TABLE: [number, string, number | 'month', string][] = [
  [180_000,       'minute',  1_000,      ':ss'],
  [180_000,       'minute',  5_000,      ':ss'],
  [180_000,       'minute',  10_000,     ':ss'],
  [7_200_000,     'hour',    60_000,     ':mm'],
  [7_200_000,     'hour',    300_000,    ':mm'],
  [259_200_000,   'day',     3_600_000,  'HH:00'],
  [259_200_000,   'day',     10_800_000, 'HH:00'],
  [7_776_000_000, 'month',   86_400_000, 'D'],
  [7_776_000_000, 'month',   604_800_000,'D'],
  [Infinity,      'year',    'month',    'M月'],
];

function chooseTicks(viewDur: number, w: number): TickCfg {
  const maxTicks = Math.max(1, w / 60);
  for (const [maxDur, upper, step, fmt] of TICK_TABLE) {
    if (viewDur > maxDur) continue;
    if (typeof step === 'number') {
      if (viewDur / step <= maxTicks) return { upper, step, fmt };
    } else {
      return { upper, step, fmt };
    }
  }
  return { upper: 'year', step: 'month', fmt: 'M月' };
}

function fmtTick(ms: number, fmt: string): string {
  const d = new Date(ms);
  switch (fmt) {
    case ':ss': return ':' + String(d.getSeconds()).padStart(2, '0');
    case ':mm': return ':' + String(d.getMinutes()).padStart(2, '0');
    case 'HH:00': return String(d.getHours()).padStart(2, '0') + ':00';
    case 'D': return d.getDate() + '日';
    case 'M月': return (d.getMonth() + 1) + '月';
    default: return '';
  }
}

function fmtUpper(ms: number, unit: string): string {
  const d = new Date(ms);
  switch (unit) {
    case 'minute': return String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0');
    case 'hour': return String(d.getHours()).padStart(2, '0') + ':00';
    case 'day': return (d.getMonth() + 1) + '月' + d.getDate() + '日';
    case 'month': return d.getFullYear() + '年' + (d.getMonth() + 1) + '月';
    case 'year': return d.getFullYear() + '年';
    default: return '';
  }
}

function upperFloor(ms: number, unit: string): number {
  const d = new Date(ms);
  switch (unit) {
    case 'minute': return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours(), d.getMinutes()).getTime();
    case 'hour':   return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours()).getTime();
    case 'day':    return new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
    case 'month':  return new Date(d.getFullYear(), d.getMonth()).getTime();
    case 'year':   return new Date(d.getFullYear(), 0).getTime();
    default: return ms;
  }
}

function upperCeil(ms: number, unit: string): number {
  const d = new Date(ms);
  switch (unit) {
    case 'minute': return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours(), d.getMinutes() + 1).getTime();
    case 'hour':   return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours() + 1).getTime();
    case 'day':    return new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1).getTime();
    case 'month':  return new Date(d.getFullYear(), d.getMonth() + 1).getTime();
    case 'year':   return new Date(d.getFullYear() + 1, 0).getTime();
    default: return ms + 1;
  }
}

function genLowerTicks(s: number, e: number, cfg: TickCfg): number[] {
  const out: number[] = [];
  if (typeof cfg.step === 'number') {
    const first = Math.ceil(s / cfg.step) * cfg.step;
    for (let t = first; t <= e; t += cfg.step) out.push(t);
  } else {
    const d = new Date(s);
    let cur = new Date(d.getFullYear(), d.getMonth(), 1).getTime();
    if (cur < s) { const c = new Date(cur); cur = new Date(c.getFullYear(), c.getMonth() + 1, 1).getTime(); }
    while (cur <= e) { out.push(cur); const c = new Date(cur); cur = new Date(c.getFullYear(), c.getMonth() + 1, 1).getTime(); }
  }
  return out;
}

function genUpperGroups(s: number, e: number, unit: string): { s: number; e: number; label: string }[] {
  const out: { s: number; e: number; label: string }[] = [];
  let cur = upperFloor(s, unit);
  while (cur < e) {
    const next = upperCeil(cur, unit);
    out.push({ s: Math.max(cur, s), e: Math.min(next, e), label: fmtUpper(cur, unit) });
    cur = next;
  }
  return out;
}

function createHatchPattern(ctx: CanvasRenderingContext2D): CanvasPattern | null {
  const c = document.createElement('canvas');
  c.width = c.height = 8;
  const g = c.getContext('2d');
  if (!g) return null;
  g.strokeStyle = 'rgba(56,189,248,0.4)';
  g.lineWidth = 2;
  g.beginPath(); g.moveTo(0, 8); g.lineTo(8, 0); g.stroke();
  g.beginPath(); g.moveTo(-4, 4); g.lineTo(4, -4); g.stroke();
  g.beginPath(); g.moveTo(4, 12); g.lineTo(12, 4); g.stroke();
  return ctx.createPattern(c, 'repeat');
}

// ── Component ──

export default function FlowGantt({
  mode,
  executionPath,
  allSteps,
  bodyStepMap = {},
  branchStepMap = {},
  controlTypeMap = {},
  currentStep,
  selectedStep,
  pinnedSteps,
  onSelectStep,
  debugStatus,
  breakpointSteps = new Set<string>(),
  onToggleBreakpoint,
  stepPlannedDurations = {},
  stepInfoMap = {},
  getExecutorMeta,
}: FlowGanttProps) {
  // ── Mode flags ──
  const showBreakpoints = mode === 'debug';
  const showPlannedBars = mode === 'debug' || mode === 'edit';
  const showExecutionBars = mode !== 'edit';
  const showDebugStatus = mode === 'debug';
  const showCurrentHighlight = mode === 'debug';

  // ── Inject stripe animation CSS once (not in render to avoid animation reset) ──
  // ── rAF-driven neon pulse (bypasses React + CSS animation entirely) ──
  useEffect(() => {
    let frame: number;
    const animate = () => {
      const t = performance.now();
      // Breathing cycle: 1.5s period, sine wave
      const breath = (Math.sin(t / 1500 * Math.PI * 2) + 1) / 2; // 0→1→0
      const glowSize = 6 + breath * 14; // 6px→20px
      const glowAlpha = 0.3 + breath * 0.4; // 0.3→0.7


      const bars = document.querySelectorAll('.gantt-neon-bar');
      bars.forEach((el) => {
        (el as HTMLElement).style.boxShadow =
          `0 0 ${glowSize}px ${glowSize / 3}px rgba(22,119,255,${glowAlpha}), inset 0 0 6px rgba(255,255,255,0.1)`;
      });

      // Stripe scroll: move background-position in reverse direction (bar grows right, stripes scroll left)
      const stripeOffset = (t % 200) / 200 * 20; // 0→20px over 200ms (4x original)
      const stripes = document.querySelectorAll('.gantt-neon-stripe');
      stripes.forEach((el) => {
        (el as HTMLElement).style.backgroundPosition = `${-stripeOffset}px 0px`;
      });



      // Cursor pulse: faster, 0.8s period
      const cursorBreath = (Math.sin(t / 800 * Math.PI * 2) + 1) / 2;
      const cursorAlpha = 0.6 + cursorBreath * 0.4;
      const cursors = document.querySelectorAll('.gantt-neon-cursor');
      cursors.forEach((el) => {
        (el as HTMLElement).style.opacity = String(cursorAlpha);
        (el as HTMLElement).style.boxShadow =
          `0 0 8px 3px rgba(140,200,255,${cursorAlpha}), 0 0 16px 6px rgba(22,119,255,${cursorAlpha * 0.5})`;
      });

      frame = requestAnimationFrame(animate);
    };
    frame = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(frame);
  }, []);

  const scrollRef = useRef<HTMLDivElement>(null);
  const canvasBoxRef = useRef<HTMLDivElement>(null);
  const barBoxRef = useRef<HTMLDivElement>(null);
  const barCvRef = useRef<HTMLCanvasElement>(null);
  const rulerCvRef = useRef<HTMLCanvasElement>(null);
  const dragRef = useRef<{ sx: number; vp: { startMs: number; endMs: number }; moved: boolean } | null>(null);
  const hatchRef = useRef<CanvasPattern | null>(null);

  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [viewport, setViewport] = useState({ startMs: 0, endMs: MIN_DURATION });
  const [cvW, setCvW] = useState(600);
  const [dragging, setDragging] = useState(false);

  const toggle = useCallback((id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }, []);

  // ── Lookup maps ──

  const entryMap = useMemo(() => {
    const m = new Map<string, ExecutionPathEntry>();
    for (const e of executionPath) m.set(e.step, e);
    return m;
  }, [executionPath]);

  // Group iteration entries: parent → iterIndex → entries[]
  const iterByParent = useMemo(() => {
    const m = new Map<string, Map<number, ExecutionPathEntry[]>>();
    for (const e of executionPath) {
      if (!e.parent_step_id || e.iteration_type == null) continue;
      let byIdx = m.get(e.parent_step_id);
      if (!byIdx) { byIdx = new Map(); m.set(e.parent_step_id, byIdx); }
      const idx = e.iteration_index ?? 0;
      let arr = byIdx.get(idx);
      if (!arr) { arr = []; byIdx.set(idx, arr); }
      arr.push(e);
    }
    return m;
  }, [executionPath]);

  // ── Build tree (supports infinite nesting: loop, foreach, if, switch) ──

  const tree = useMemo(() => {
    const bodyIdSet = new Set<string>();
    for (const bodies of Object.values(bodyStepMap)) {
      for (const b of bodies) bodyIdSet.add(b);
    }

    const getStatus = (stepId: string): string => {
      const e = entryMap.get(stepId);
      let s = e?.status ?? 'waiting';
      if (showDebugStatus && debugStatus === 'ended' && (s === 'running' || s === 'pending')) s = 'cancelled';
      return s;
    };

    /** Build children for a step that has nested content (body or branches).
     *  Recursively handles infinite nesting of any control type.
     *  scopePrefix: when inside a loop iteration, scoped IDs like "loop1-1-" are prepended
     *  to look up runtime data (entryMap) with the correct scoped step ID. */
    const buildStepChildren = (stepId: string, depth: number, scopePrefix = ''): GridNode[] => {
      const ct = controlTypeMap[stepId];
      const hasBranches = branchStepMap[stepId];

      // if/switch → show branches as groups, each with their children
      if ((ct === 'if' || ct === 'switch') && hasBranches) {
        return buildBranchChildren(stepId, depth, scopePrefix);
      }

      // loop/foreach → check for iteration data first, then static body
      if (ct === 'loop' || ct === 'foreach') {
        // When inside a scope (e.g. nested loop), check scoped parent ID for iterations
        const scopedParent = scopePrefix ? `${scopePrefix}${stepId}` : stepId;
        if (iterByParent.has(scopedParent)) {
          return buildIterChildren(scopedParent, depth);
        }
        if (iterByParent.has(stepId)) {
          return buildIterChildren(stepId, depth);
        }
        // Static body (no runtime iteration data yet)
        return buildBodySteps(stepId, 0, depth, scopePrefix);
      }

      // Generic: if bodyStepMap has children, show them
      if (bodyStepMap[stepId]) {
        const scopedParent = scopePrefix ? `${scopePrefix}${stepId}` : stepId;
        if (iterByParent.has(scopedParent)) {
          return buildIterChildren(scopedParent, depth);
        }
        if (iterByParent.has(stepId)) {
          return buildIterChildren(stepId, depth);
        }
        return buildBodySteps(stepId, 0, depth, scopePrefix);
      }

      return [];
    };

    /** Build branch group nodes for if/switch.
     *  scopePrefix is passed through so scoped IDs resolve correctly. */
    const buildBranchChildren = (parentId: string, depth: number, scopePrefix = ''): GridNode[] => {
      const branches = branchStepMap[parentId];
      if (!branches) return [];

      return Object.entries(branches).map(([branchKey, childIds]) => {
        const scopedParentId = scopePrefix ? `${scopePrefix}${parentId}` : parentId;
        const branchNodeId = `${scopedParentId}-branch-${branchKey}`;
        const branchChildren: GridNode[] = childIds.map((childId) => {
          const scopedChildId = scopePrefix ? `${scopePrefix}${childId}` : childId;
          const entry = entryMap.get(scopedChildId) || entryMap.get(childId);
          const nested = buildStepChildren(childId, depth + 2, scopePrefix);
          const info = stepInfoMap[childId];
          return {
            id: entry ? scopedChildId : childId,
            label: info?.name || childId,
            status: entry ? (entry.status ?? 'waiting') : getStatus(scopedChildId) !== 'waiting' ? getStatus(scopedChildId) : getStatus(childId),
            started_at: entry?.started_at,
            duration_ms: entry?.duration_ms,
            depth: depth + 1,
            isBody: true,
            controlType: controlTypeMap[childId],
            executorType: info?.executor,
            breakpointId: childId,
            children: nested,
          };
        });

        // Aggregate status from children
        let branchStatus = 'waiting';
        for (const c of branchChildren) {
          if (c.status === 'skipped') { if (branchStatus === 'waiting') branchStatus = 'skipped'; continue; }
          if (c.status === 'failed') { branchStatus = 'failed'; break; }
          if (c.status === 'running' || c.status.startsWith('paused')) { branchStatus = 'running'; break; }
          if (c.status === 'completed') branchStatus = 'completed';
        }

        return {
          id: branchNodeId,
          label: branchKey,
          status: branchStatus,
          depth,
          isBranch: true,
          branchKey,
          children: branchChildren,
        };
      });
    };

    /** Build body step nodes (for loop/foreach).
     *  scopePrefix propagates the iteration scope for nested control nodes. */
    const buildBodySteps = (parentId: string, iterIdx: number, depth: number, outerScope = ''): GridNode[] => {
      const bodies = bodyStepMap[parentId] ?? [];
      // For if/switch, body is all branch IDs flattened — use branchStepMap instead
      if (branchStepMap[parentId]) return [];
      // Build scope prefix for this iteration: e.g. "nested_loop-1-"
      const iterScope = `${outerScope ? outerScope : ''}${parentId}-${iterIdx + 1}-`;
      return bodies.map((bodyId) => {
        const scopedId = `${iterScope}${bodyId}`;
        const entry = entryMap.get(scopedId);
        const nodeId = entry ? scopedId : bodyId;

        const nested = buildStepChildren(bodyId, depth + 1, iterScope);

        const bodyInfo = stepInfoMap[bodyId];
        return {
          id: nodeId,
          label: bodyInfo?.name || bodyId,
          status: entry ? getStatus(scopedId) : getStatus(bodyId),
          started_at: entry?.started_at,
          duration_ms: entry?.duration_ms,
          depth,
          isBody: true,
          controlType: controlTypeMap[bodyId],
          executorType: bodyInfo?.executor,
          breakpointId: bodyId,
          children: nested,
        };
      });
    };

    /** Build iteration group nodes (loop/foreach runtime data) */
    const buildIterChildren = (parentId: string, depth: number): GridNode[] => {
      const byIdx = iterByParent.get(parentId);
      if (!byIdx) return [];
      const iterType = executionPath.find((e) => e.parent_step_id === parentId)?.iteration_type;
      const sortedIndices = [...byIdx.keys()].sort((a, b) => a - b);

      return sortedIndices.map((idx) => {
        const entries = byIdx.get(idx)!;
        const first = entries[0];
        const itemVal = first?.iteration_item;
        const label = iterType === 'foreach'
          ? `Item #${idx + 1}${itemVal ? `: ${itemVal.length > 15 ? itemVal.slice(0, 15) + '…' : itemVal}` : ''}`
          : `Iter #${idx + 1}`;

        const iterEntryId = `${parentId}-iter-${idx + 1}`;
        const iterEntry = entryMap.get(iterEntryId);
        const iterStart = iterEntry?.started_at ?? first?.started_at;
        const iterDuration = iterEntry?.duration_ms ?? first?.duration_ms;

        let iterStatus = 'completed';
        for (const e of entries) {
          if (e.status === 'failed') { iterStatus = 'failed'; break; }
          if (e.status === 'running') { iterStatus = 'running'; break; }
        }

        const bodyChildren = buildBodySteps(parentId, idx, depth + 1);
        return {
          id: `${parentId}-iter-group-${idx}`,
          label,
          status: iterStatus,
          started_at: iterStart,
          duration_ms: iterDuration,
          depth,
          isIteration: true,
          children: bodyChildren,
        };
      });
    };

    // ── Topological sort by depends_on ──
    // Steps with no dependencies come first; dependents come after their dependencies.
    const topoSorted = (() => {
      const filtered = allSteps.filter((id) => !bodyIdSet.has(id));
      const idxMap = new Map(filtered.map((id, i) => [id, i]));
      const adj = new Map<string, string[]>(); // step → steps that depend on it
      const inDeg = new Map<string, number>();
      for (const id of filtered) {
        adj.set(id, []);
        inDeg.set(id, 0);
      }
      for (const id of filtered) {
        const deps = stepInfoMap[id]?.depends_on;
        if (deps) {
          for (const raw of deps) {
            // depends_on may use "step" or "step.output_field" syntax
            const dep = raw.includes('.') ? raw.split('.')[0] : raw;
            if (idxMap.has(dep)) {
              adj.get(dep)!.push(id);
              inDeg.set(id, (inDeg.get(id) ?? 0) + 1);
            }
          }
        }
      }
      // Kahn's algorithm — stable: among zero-indegree nodes, preserve original YAML order
      const queue = filtered.filter((id) => (inDeg.get(id) ?? 0) === 0);
      const result: string[] = [];
      while (queue.length > 0) {
        const cur = queue.shift()!;
        result.push(cur);
        for (const next of adj.get(cur) ?? []) {
          const d = (inDeg.get(next) ?? 1) - 1;
          inDeg.set(next, d);
          if (d === 0) {
            // Insert into queue maintaining original YAML order (stable sort)
            const nextIdx = idxMap.get(next) ?? Infinity;
            let insertAt = queue.length;
            for (let i = 0; i < queue.length; i++) {
              if ((idxMap.get(queue[i]) ?? Infinity) > nextIdx) {
                insertAt = i;
                break;
              }
            }
            queue.splice(insertAt, 0, next);
          }
        }
      }
      // Safety: append any unreachable steps (cycle protection)
      for (const id of filtered) {
        if (!result.includes(id)) result.push(id);
      }
      return result;
    })();

    const nodes: GridNode[] = [];
    for (const stepId of topoSorted) {
      const entry = entryMap.get(stepId);

      const stepInfo = stepInfoMap[stepId];
      const node: GridNode = {
        id: stepId,
        label: stepInfo?.name || stepId,
        status: getStatus(stepId),
        started_at: entry?.started_at,
        duration_ms: entry?.duration_ms,
        depth: 0,
        controlType: controlTypeMap[stepId],
        executorType: stepInfo?.executor,
        children: buildStepChildren(stepId, 1),
      };

      nodes.push(node);
    }
    return nodes;
  }, [allSteps, bodyStepMap, branchStepMap, controlTypeMap, entryMap, iterByParent, executionPath, debugStatus, showDebugStatus, stepInfoMap]);

  // ── Global time span ──

  const { epochBase, totalSpan } = useMemo(() => {
    let minStart = Infinity;
    let maxEnd = 0;
    const walk = (nodes: GridNode[]) => {
      for (const n of nodes) {
        if (n.started_at) {
          if (n.started_at < minStart) minStart = n.started_at;
          const end = n.started_at + (n.duration_ms ?? 0);
          if (end > maxEnd) maxEnd = end;
        }
        if (n.status === 'waiting' && stepPlannedDurations[n.id]) {
          const pEnd = (maxEnd || Date.now()) + stepPlannedDurations[n.id];
          if (pEnd > maxEnd) maxEnd = pEnd;
        }
        walk(n.children);
      }
    };
    walk(tree);
    const base = isFinite(minStart) ? minStart : Date.now();
    return { epochBase: base, totalSpan: Math.max(maxEnd - base, 1000) };
  }, [tree, stepPlannedDurations]);

  // ── Flatten visible rows ──

  const visibleRows = useMemo(() => {
    const rows: GridNode[] = [];
    const seen = new Set<string>();
    const addNodes = (nodes: GridNode[]) => {
      for (const node of nodes) {
        if (seen.has(node.id)) continue; // dedup safety
        seen.add(node.id);
        rows.push(node);
        if (node.children.length > 0 && expanded.has(node.id)) {
          if (node.children.length > MAX_VISIBLE_ITERATIONS && node.children[0]?.isIteration) {
            const first = node.children.slice(0, TRUNCATED_VISIBLE);
            const last = node.children.slice(-TRUNCATED_VISIBLE);
            addNodes(first);
            const ellipsisId = `${node.id}-ellipsis`;
            if (!seen.has(ellipsisId)) {
              seen.add(ellipsisId);
              rows.push({
                id: ellipsisId,
                label: `… ${node.children.length - TRUNCATED_VISIBLE * 2} more …`,
                status: 'completed',
                depth: node.depth + 1,
                isIteration: true,
                children: [],
              });
            }
            addNodes(last);
          } else {
            addNodes(node.children);
          }
        }
      }
    };
    addNodes(tree);
    return rows;
  }, [tree, expanded]);

  // ── Fit-all viewport ──

  const fitAll = useCallback(() => {
    const pad = totalSpan * 0.1;
    setViewport({ startMs: epochBase - pad, endMs: epochBase + totalSpan + pad });
  }, [epochBase, totalSpan]);

  useEffect(() => { fitAll(); }, [fitAll]);

  // ── Auto-scroll ──

  useEffect(() => {
    if (!currentStep || !scrollRef.current) return;
    const el = scrollRef.current.querySelector(`[data-step="${currentStep}"]`);
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  }, [currentStep]);

  // ── ResizeObserver ──

  useEffect(() => {
    const box = canvasBoxRef.current;
    if (!box) return;
    const ro = new ResizeObserver((es) => {
      const w = Math.floor(es[0].contentRect.width);
      if (w > 0) setCvW(w);
    });
    ro.observe(box);
    return () => ro.disconnect();
  }, []);

  // ── Draw ruler ──

  const drawRuler = useCallback(() => {
    const cv = rulerCvRef.current;
    if (!cv) return;
    const ctx = cv.getContext('2d');
    if (!ctx) return;
    const dpr = window.devicePixelRatio || 1;
    const w = cvW;
    cv.width = w * dpr; cv.height = RULER_HEIGHT * dpr;
    cv.style.width = w + 'px'; cv.style.height = RULER_HEIGHT + 'px';
    ctx.scale(dpr, dpr);

    const vDur = viewport.endMs - viewport.startMs;
    const px = w / vDur;
    const ticks = chooseTicks(vDur, w);

    ctx.clearRect(0, 0, w, RULER_HEIGHT);

    // Upper groups with alternating backgrounds
    const groups = genUpperGroups(viewport.startMs, viewport.endMs, ticks.upper);
    groups.forEach((g, i) => {
      const x1 = (g.s - viewport.startMs) * px;
      const x2 = (g.e - viewport.startMs) * px;
      ctx.fillStyle = i % 2 === 0 ? '#1a1a2e' : '#16162a';
      ctx.fillRect(x1, 0, x2 - x1, RULER_UPPER_H);
      if (x2 - x1 > 40) {
        ctx.fillStyle = 'rgba(255,255,255,0.85)'; ctx.font = '11px sans-serif';
        ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
        ctx.fillText(g.label, (x1 + x2) / 2, RULER_UPPER_H / 2);
      }
      if (i > 0) {
        ctx.strokeStyle = '#333'; ctx.lineWidth = 1;
        ctx.beginPath(); ctx.moveTo(x1, 0); ctx.lineTo(x1, RULER_HEIGHT); ctx.stroke();
      }
    });

    // Lower background
    ctx.fillStyle = '#0f0f20';
    ctx.fillRect(0, RULER_UPPER_H, w, RULER_LOWER_H);

    // Lower ticks: draw lines for all ticks, text centered in cells (skip origin)
    const lt = genLowerTicks(viewport.startMs, viewport.endMs, ticks);
    ctx.font = '10px sans-serif'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
    for (const t of lt) {
      const x = (t - viewport.startMs) * px;
      ctx.strokeStyle = '#333'; ctx.lineWidth = 1;
      ctx.beginPath(); ctx.moveTo(x, RULER_UPPER_H); ctx.lineTo(x, RULER_HEIGHT); ctx.stroke();
    }
    // Text labels centered in cell to the RIGHT of each tick line
    // e.g. "2月" label goes between Feb 1 and Mar 1 tick lines (February's cell)
    for (let i = 0; i < lt.length - 1; i++) {
      const x0 = (lt[i] - viewport.startMs) * px;
      const x1 = (lt[i + 1] - viewport.startMs) * px;
      ctx.fillStyle = 'rgba(255,255,255,0.7)';
      ctx.fillText(fmtTick(lt[i], ticks.fmt), (x0 + x1) / 2, RULER_UPPER_H + RULER_LOWER_H / 2);
    }

    // Bottom border
    ctx.strokeStyle = '#252540'; ctx.lineWidth = 1;
    ctx.beginPath(); ctx.moveTo(0, RULER_HEIGHT - 0.5); ctx.lineTo(w, RULER_HEIGHT - 0.5); ctx.stroke();
  }, [viewport, cvW]);

  // ── Draw bars ──

  const drawBars = useCallback(() => {
    const cv = barCvRef.current;
    if (!cv) return;
    const ctx = cv.getContext('2d');
    if (!ctx) return;
    const dpr = window.devicePixelRatio || 1;
    const w = cvW;
    const h = Math.max(visibleRows.length * ROW_HEIGHT, 1);
    cv.width = w * dpr; cv.height = h * dpr;
    cv.style.width = w + 'px'; cv.style.height = h + 'px';
    ctx.scale(dpr, dpr);

    const vDur = viewport.endMs - viewport.startMs;
    const px = w / vDur;
    const ticks = chooseTicks(vDur, w);

    // Background
    ctx.fillStyle = '#0d0d1a';
    ctx.fillRect(0, 0, w, h);

    // Grid lines from lower ticks
    const lt = genLowerTicks(viewport.startMs, viewport.endMs, ticks);
    ctx.strokeStyle = 'rgba(255,255,255,0.04)'; ctx.lineWidth = 1;
    for (const t of lt) {
      const x = Math.round((t - viewport.startMs) * px) + 0.5;
      ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, h); ctx.stroke();
    }

    // Upper group boundary lines
    const groups = genUpperGroups(viewport.startMs, viewport.endMs, ticks.upper);
    ctx.strokeStyle = 'rgba(255,255,255,0.08)'; ctx.lineWidth = 1;
    for (let i = 1; i < groups.length; i++) {
      const x = Math.round((groups[i].s - viewport.startMs) * px) + 0.5;
      ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, h); ctx.stroke();
    }

    // Ensure hatch pattern
    if (!hatchRef.current) hatchRef.current = createHatchPattern(ctx);

    // Sequential cursor for planned bars in edit mode (no execution data)
    let plannedCursor = epochBase;
    visibleRows.forEach((node, ri) => {
      const y = ri * ROW_HEIGHT;
      if (node.id.endsWith('-ellipsis')) return;
      const isRunning = node.status === 'running' || node.status.startsWith('paused');
      const isCurrent = showCurrentHighlight && debugStatus !== 'ended' && node.id === currentStep;
      const barH = node.isIteration || node.isBody ? 14 : 18;
      const barY = y + (ROW_HEIGHT - barH) / 2;
      const colors = getStatusColors(node.status);

      // Row highlight
      if (isCurrent) { ctx.fillStyle = 'rgba(22,119,255,0.15)'; ctx.fillRect(0, y, w, ROW_HEIGHT); }
      else if (node.id === selectedStep) { ctx.fillStyle = 'rgba(26,39,68,0.5)'; ctx.fillRect(0, y, w, ROW_HEIGHT); }

      // Running bars are rendered as CSS overlays with barber-pole animation
      if (isRunning) {
        // skip — CSS overlay renders these
      } else if (showExecutionBars && node.started_at) {
        const bS = node.started_at;
        const bE = node.duration_ms ? bS + node.duration_ms : bS;
        const x1 = (bS - viewport.startMs) * px;
        const x2 = (bE - viewport.startMs) * px;
        const bW = Math.max(x2 - x1, 3);
        if (x1 + bW < 0 || x1 > w) return; // off-screen

        // Bar fill
        ctx.globalAlpha = node.status === 'completed' ? 0.8 : node.status === 'skipped' ? 0.4 : 0.9;
        ctx.fillStyle = colors.bar;
        ctx.beginPath(); ctx.roundRect(x1, barY, bW, barH, 3); ctx.fill();
        ctx.globalAlpha = 1.0;

        // Current cursor — drawn by CSS overlay (gantt-neon-cursor), not canvas

        // Duration label — inside bar if wide enough, otherwise to the right
        if (node.duration_ms != null && node.duration_ms > 0) {
          const durText = formatDuration(node.duration_ms);
          ctx.font = '10px sans-serif';
          ctx.textBaseline = 'middle';
          ctx.shadowColor = 'rgba(0,0,0,0.9)'; ctx.shadowBlur = 4;
          const textW = ctx.measureText(durText).width;
          if (bW > textW + 8) {
            ctx.fillStyle = '#fff'; ctx.textAlign = 'left';
            ctx.fillText(durText, x1 + 4, barY + barH / 2);
          } else {
            ctx.fillStyle = 'rgba(255,255,255,0.7)'; ctx.textAlign = 'left';
            ctx.fillText(durText, x1 + bW + 4, barY + barH / 2);
          }
          ctx.shadowBlur = 0;
        }
      } else if (node.status === 'waiting' && showPlannedBars) {
        const plannedMs = stepPlannedDurations[node.id];
        if (plannedMs && plannedMs > 0) {
          // In edit mode (no execution bars), use sequential cursor; in debug, after actual bars
          const pStart = showExecutionBars ? epochBase + totalSpan : plannedCursor;
          if (!showExecutionBars) plannedCursor += plannedMs;
          const x1 = (pStart - viewport.startMs) * px;
          const x2 = (pStart + plannedMs - viewport.startMs) * px;
          const bW = Math.max(x2 - x1, 50);
          if (x1 + bW >= 0 && x1 <= w) {
            // Sky-blue hatched fill
            if (hatchRef.current) {
              ctx.fillStyle = hatchRef.current;
              ctx.beginPath(); ctx.roundRect(x1, barY, bW, barH, 3); ctx.fill();
            }
            // Dashed border
            ctx.strokeStyle = 'rgba(56,189,248,0.6)'; ctx.lineWidth = 1;
            ctx.setLineDash([4, 3]);
            ctx.beginPath(); ctx.roundRect(x1, barY, bW, barH, 3); ctx.stroke();
            ctx.setLineDash([]);
            // Label
            if (bW > 40) {
              ctx.fillStyle = 'rgba(56,189,248,0.9)'; ctx.font = '10px sans-serif';
              ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
              ctx.fillText('预计 ' + formatDuration(plannedMs), x1 + 6, barY + barH / 2);
            }
          }
        } else if (showExecutionBars) {
          // Dashed placeholder for waiting without planned (only in runtime/debug)
          const phX = Math.max(0, (epochBase + totalSpan - viewport.startMs) * px);
          if (phX < w) {
            ctx.strokeStyle = '#303040'; ctx.lineWidth = 1; ctx.setLineDash([4, 4]);
            ctx.strokeRect(phX, barY, 60, barH); ctx.setLineDash([]);
            ctx.fillStyle = '#444'; ctx.font = 'italic 10px sans-serif';
            ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
            ctx.fillText('waiting', phX + 6, barY + barH / 2);
          }
        }
      }

      // Row divider
      ctx.strokeStyle = 'rgba(255,255,255,0.03)'; ctx.lineWidth = 1;
      ctx.beginPath(); ctx.moveTo(0, y + ROW_HEIGHT - 0.5); ctx.lineTo(w, y + ROW_HEIGHT - 0.5); ctx.stroke();
    });
  }, [viewport, cvW, visibleRows, currentStep, selectedStep, debugStatus, epochBase, totalSpan, stepPlannedDurations, showExecutionBars, showPlannedBars, showCurrentHighlight]);

  // ── Redraw on changes (synchronous, before browser paint) ──

  useLayoutEffect(() => {
    drawRuler();
    drawBars();
  }, [drawRuler, drawBars]);

  // ── Live redraw timer for running steps ──

  const hasRunningSteps = useMemo(() => {
    const check = (nodes: GridNode[]): boolean =>
      nodes.some((n) => n.status === 'running' || n.status.startsWith('paused') || check(n.children));
    return check(tree);
  }, [tree]);

  const [tick, setTick] = useState(0);
  useEffect(() => {
    if (!hasRunningSteps) return;
    // Fast tick (100ms) for smooth CSS bar growth; canvas redraws every 5th tick
    let count = 0;
    const id = setInterval(() => {
      count++;
      setTick((t) => t + 1);
      if (count % 5 === 0) { drawRuler(); drawBars(); }
    }, 100);
    return () => clearInterval(id);
  }, [hasRunningSteps, drawRuler, drawBars]);

  // ── Running bars overlay data (CSS-driven animation) ──

  const runningBars = useMemo(() => {
    if (!showExecutionBars) return [];
    const now = Date.now();
    const vDur = viewport.endMs - viewport.startMs;
    if (vDur <= 0) return [];

    const bars: { id: string; left: number; width: number; top: number; height: number; duration: number; isCurrent: boolean }[] = [];
    visibleRows.forEach((node, ri) => {
      const isRunning = node.status === 'running' || node.status.startsWith('paused');
      if (!isRunning || !node.started_at) return;

      const barH = node.isIteration || node.isBody ? 14 : 18;
      const y = ri * ROW_HEIGHT;
      const barY = y + (ROW_HEIGHT - barH) / 2;

      const bS = node.started_at;
      const bE = node.duration_ms ? bS + node.duration_ms : now;
      const x1 = (bS - viewport.startMs) / vDur * cvW;
      const x2 = (bE - viewport.startMs) / vDur * cvW;
      const bW = Math.max(x2 - x1, 3);

      if (x1 + bW < 0 || x1 > cvW) return;

      const displayDuration = now - node.started_at;
      const isCurrent = showCurrentHighlight && debugStatus !== 'ended' && node.id === currentStep;

      bars.push({ id: node.id, left: x1, width: bW, top: barY, height: barH, duration: displayDuration, isCurrent });
    });
    return bars;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visibleRows, viewport, cvW, showExecutionBars, showCurrentHighlight, debugStatus, currentStep, tick]);

  // ── Wheel zoom (anchor-based, non-passive) ──

  useEffect(() => {
    const onWheel = (e: WheelEvent) => {
      if (Math.abs(e.deltaY) < 1) return;
      // Ctrl+wheel = zoom timeline; plain wheel = normal scroll (don't intercept)
      if (!e.ctrlKey && !e.metaKey) return;
      e.preventDefault();
      // Use canvasBoxRef (ruler area) to get consistent left offset
      const rect = canvasBoxRef.current?.getBoundingClientRect();
      if (!rect) return;
      const mouseX = Math.max(0, Math.min(e.clientX - rect.left, cvW));
      setViewport((vp) => {
        const vDur = vp.endMs - vp.startMs;
        const anchor = vp.startMs + (mouseX / cvW) * vDur;
        const factor = e.deltaY < 0 ? ZOOM_FACTOR : 1 / ZOOM_FACTOR;
        const newDur = Math.max(MIN_DURATION, Math.min(MAX_DURATION, vDur / factor));
        const ratio = mouseX / cvW;
        const newStart = anchor - ratio * newDur;
        return { startMs: newStart, endMs: newStart + newDur };
      });
    };
    const ruler = canvasBoxRef.current;
    const barBox = barBoxRef.current;
    ruler?.addEventListener('wheel', onWheel, { passive: false });
    barBox?.addEventListener('wheel', onWheel, { passive: false });
    return () => {
      ruler?.removeEventListener('wheel', onWheel);
      barBox?.removeEventListener('wheel', onWheel);
    };
  }, [cvW]);

  // ── Drag pan ──

  const onMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button !== 0) return;
    setDragging(true);
    dragRef.current = { sx: e.clientX, vp: { ...viewport }, moved: false };
  }, [viewport]);

  const onMouseMove = useCallback((e: React.MouseEvent) => {
    if (!dragging || !dragRef.current) return;
    const dx = e.clientX - dragRef.current.sx;
    if (Math.abs(dx) > 3) dragRef.current.moved = true;
    const vDur = dragRef.current.vp.endMs - dragRef.current.vp.startMs;
    const dMs = -(dx / cvW) * vDur;
    setViewport({ startMs: dragRef.current.vp.startMs + dMs, endMs: dragRef.current.vp.endMs + dMs });
  }, [dragging, cvW]);

  const didDragRef = useRef(false);

  const onMouseUp = useCallback(() => {
    didDragRef.current = dragRef.current?.moved ?? false;
    setDragging(false);
    dragRef.current = null;
  }, []);

  useEffect(() => {
    if (!dragging) return;
    const up = () => {
      didDragRef.current = dragRef.current?.moved ?? false;
      setDragging(false);
      dragRef.current = null;
    };
    window.addEventListener('mouseup', up);
    return () => window.removeEventListener('mouseup', up);
  }, [dragging]);

  // ── Canvas click → select step ──

  const onCanvasClick = useCallback((e: React.MouseEvent) => {
    if (didDragRef.current) { didDragRef.current = false; return; }
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const y = e.clientY - rect.top;
    const ri = Math.floor(y / ROW_HEIGHT);
    if (ri >= 0 && ri < visibleRows.length && !visibleRows[ri].id.endsWith('-ellipsis')) {
      onSelectStep(visibleRows[ri].id);
    }
  }, [visibleRows, onSelectStep]);

  // ── Render left row (DOM) ──

  const renderLeftRow = (node: GridNode) => {
    const isCurrent = showCurrentHighlight && debugStatus !== 'ended' && node.id === currentStep;
    const isSelected = node.id === selectedStep;
    const executorMeta = node.executorType && getExecutorMeta ? getExecutorMeta(node.executorType) : null;
    const isPinned = pinnedSteps.has(node.id);
    const hasChildren = node.children.length > 0;
    const isExp = expanded.has(node.id);
    const isEllipsis = node.id.endsWith('-ellipsis');

    return (
      <div
        key={node.id}
        data-step={node.id}
        onClick={() => !isEllipsis && onSelectStep(node.id)}
        style={{
          display: 'flex',
          height: ROW_HEIGHT,
          alignItems: 'center',
          gap: 4,
          cursor: isEllipsis ? 'default' : 'pointer',
          background: isCurrent ? 'rgba(22,119,255,0.25)' : isSelected ? '#1a2744' : '#141422',
          borderLeft: isCurrent ? '3px solid #1677ff' : isSelected ? '3px solid #4488ff' : '3px solid transparent',
          paddingLeft: node.depth * INDENT_PX + 8,
          paddingRight: 8,
          overflow: 'hidden',
          borderRight: '1px solid #252540',
          transition: 'background 0.15s',
        }}
      >
        {/* Breakpoint dot (debug mode only) */}
        {showBreakpoints && !isEllipsis && !node.isIteration && !node.isBranch && onToggleBreakpoint ? (() => {
          const bpId = node.breakpointId ?? node.id;
          const hasBP = breakpointSteps.has(bpId);
          return (
          <span
            onClick={(e) => { e.stopPropagation(); onToggleBreakpoint(bpId); }}
            onMouseEnter={(e) => {
              const dot = e.currentTarget.querySelector('span') as HTMLElement;
              if (dot && !hasBP) {
                dot.style.background = 'rgba(231,76,60,0.35)';
                dot.style.border = '1.5px solid #e74c3c80';
              }
            }}
            onMouseLeave={(e) => {
              const dot = e.currentTarget.querySelector('span') as HTMLElement;
              if (dot && !hasBP) {
                dot.style.background = 'transparent';
                dot.style.border = '1.5px solid transparent';
              }
            }}
            style={{
              width: 20, height: 20, display: 'flex', alignItems: 'center', justifyContent: 'center',
              cursor: 'pointer', flexShrink: 0, marginLeft: -4, marginRight: -4,
            }}
            title={hasBP ? '移除断点' : '设置断点'}
          >
            <span style={{
              width: 9, height: 9, borderRadius: '50%', display: 'block',
              background: hasBP ? '#e74c3c' : 'transparent',
              border: hasBP ? 'none' : '1.5px solid transparent',
              transition: 'all 0.15s',
            }} />
          </span>
          );
        })() : !showBreakpoints ? null : (
          <span style={{ width: 12, flexShrink: 0 }} />
        )}

        {/* Expand/collapse arrow */}
        {hasChildren && !isEllipsis ? (
          <span
            onClick={(e) => { e.stopPropagation(); toggle(node.id); }}
            style={{
              width: 14, height: 14, display: 'flex', alignItems: 'center', justifyContent: 'center',
              cursor: 'pointer', color: '#888', fontSize: 8, userSelect: 'none', flexShrink: 0, borderRadius: 2,
            }}
            onMouseEnter={(e) => (e.currentTarget.style.background = '#303050')}
            onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
          >
            {isExp ? '▼' : '▶'}
          </span>
        ) : (
          <span style={{ width: 14, flexShrink: 0 }} />
        )}

        {/* Status icon */}
        {!isEllipsis && (STATUS_ICONS[node.status] ?? STATUS_ICONS.waiting)}

        {/* Executor icon */}
        {!isEllipsis && !node.isIteration && !node.isBranch && executorMeta && (
          <span style={{ fontSize: 12, flexShrink: 0, lineHeight: 1 }} title={executorMeta.label}>
            {executorMeta.icon}
          </span>
        )}

        {/* Control type badge */}
        {!isEllipsis && node.controlType && (
          <span style={{
            fontSize: 8, fontWeight: 700, letterSpacing: 0.5,
            padding: '1px 4px', borderRadius: 3, flexShrink: 0,
            textTransform: 'uppercase',
            ...(node.controlType === 'if' || node.controlType === 'switch'
              ? { background: 'rgba(245,158,11,0.15)', color: '#f59e0b', border: '1px solid rgba(245,158,11,0.3)' }
              : { background: 'rgba(22,119,255,0.15)', color: '#1677ff', border: '1px solid rgba(22,119,255,0.3)' }),
          }}>
            {node.controlType}
          </span>
        )}

        {/* Branch key badge */}
        {!isEllipsis && node.isBranch && node.branchKey && (
          <span style={{
            fontSize: 8, fontWeight: 700, letterSpacing: 0.3,
            padding: '1px 5px', borderRadius: 3, flexShrink: 0,
            background: node.branchKey === 'true' ? 'rgba(82,196,26,0.15)' : node.branchKey === 'false' ? 'rgba(255,77,79,0.15)' : 'rgba(139,92,246,0.15)',
            color: node.branchKey === 'true' ? '#52c41a' : node.branchKey === 'false' ? '#ff4d4f' : '#8b5cf6',
            border: `1px solid ${node.branchKey === 'true' ? 'rgba(82,196,26,0.3)' : node.branchKey === 'false' ? 'rgba(255,77,79,0.3)' : 'rgba(139,92,246,0.3)'}`,
          }}>
            {node.branchKey}
          </span>
        )}

        {/* Step label */}
        <Text
          ellipsis
          style={{
            flex: 1,
            fontSize: node.isIteration || node.isBody || node.isBranch ? 10 : 12,
            color: isEllipsis ? '#888'
              : isCurrent ? '#fff'
              : node.isBranch ? '#e0b0ff'
              : node.isIteration ? '#ccccff'
              : node.isBody ? '#ddd'
              : node.status === 'waiting' ? '#aaa' : '#fff',
            fontWeight: isCurrent ? 600 : 400,
            fontStyle: node.isIteration || node.isBody || node.isBranch ? 'italic' : undefined,
          }}
        >
          {node.label}
        </Text>

        {/* Pinned */}
        {isPinned && <PushpinOutlined style={{ color: '#faad14', fontSize: 11, flexShrink: 0 }} />}

        {/* Count badge (collapsed) */}
        {hasChildren && !isExp && !isEllipsis && (
          <span style={{
            fontSize: 9, color: '#777', background: '#252540',
            padding: '0 4px', borderRadius: 6, whiteSpace: 'nowrap', flexShrink: 0,
          }}>
            {node.children.length}
          </span>
        )}

        {/* Duration */}
        {!isEllipsis && (node.duration_ms != null || (node.status === 'running' && node.started_at)) && (
          <Text style={{ color: node.status === 'running' ? '#91caff' : '#555', fontSize: 10, whiteSpace: 'nowrap', flexShrink: 0 }}>
            {formatDuration(
              node.status === 'running' && node.started_at
                ? Date.now() - node.started_at
                : node.duration_ms
            )}
          </Text>
        )}
      </div>
    );
  };

  // ── Layout ──

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, minHeight: 0, overflow: 'hidden', background: '#0d0d1a' }}>
      {/* Header */}
      <div style={{ display: 'flex', flexShrink: 0, borderBottom: '1px solid #252540' }}>
        <div style={{
          width: CALL_STACK_WIDTH, flexShrink: 0, padding: '8px 12px',
          fontSize: 11, fontWeight: 600, color: '#888', textTransform: 'uppercase',
          letterSpacing: 1, background: '#141422', borderRight: '1px solid #252540',
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        }}>
          <span>{mode === 'edit' ? '步骤列表' : '执行序列'}</span>
          <Text style={{ color: '#555', fontSize: 10, textTransform: 'none', letterSpacing: 0, fontWeight: 400 }}>
            {showExecutionBars
              ? `${executionPath.filter((e) => e.status === 'completed' || e.status === 'skipped').length}/${allSteps.length}`
              : `${allSteps.length} 步骤`
            }
          </Text>
        </div>
        <div ref={canvasBoxRef} style={{ flex: 1, height: RULER_HEIGHT, position: 'relative', overflow: 'hidden' }}>
          <canvas ref={rulerCvRef} style={{ display: 'block' }} />
        </div>
      </div>

      {/* Body: single scroll container with left DOM + right canvas */}
      <div ref={scrollRef} className="debug-step-grid-scroll" style={{ flex: 1, overflowY: 'auto', overflowX: 'hidden' }}>
        <div style={{ display: 'flex', minHeight: '100%' }}>
          {/* Left column: step labels (DOM) */}
          <div style={{ width: CALL_STACK_WIDTH, flexShrink: 0 }}>
            {visibleRows.map(renderLeftRow)}
            {visibleRows.length === 0 && (
              <div style={{ padding: 24, textAlign: 'center' }}>
                <Text style={{ color: '#444', fontStyle: 'italic' }}>
                  {mode === 'edit' ? '无步骤' : '等待调试启动...'}
                </Text>
              </div>
            )}
          </div>
          {/* Right column: canvas timeline */}
          <div ref={barBoxRef} style={{ flex: 1, minWidth: 0, position: 'relative' }}>
            <canvas
              ref={barCvRef}
              style={{ display: 'block', cursor: dragging ? 'grabbing' : 'grab' }}
              onMouseDown={onMouseDown}
              onMouseMove={onMouseMove}
              onMouseUp={onMouseUp}
              onClick={onCanvasClick}
            />
            {/* Running bars — CSS overlay with barber-pole animation */}
            {runningBars.length > 0 && (
              <div style={{ position: 'absolute', top: 0, left: 0, width: cvW, pointerEvents: 'none' }}>
                {runningBars.map((bar) => {
                  const durText = bar.duration > 0 ? formatDuration(bar.duration) : '';
                  const showInside = bar.width > 40;
                  return (
                    <React.Fragment key={bar.id}>
                      <div className="gantt-neon-bar" style={{
                        position: 'absolute',
                        left: bar.left,
                        top: bar.top,
                        width: bar.width,
                        height: bar.height,
                        borderRadius: 3,
                        transition: 'width 0.1s linear',
                        background: 'linear-gradient(90deg, #1677ff, #3b8bff)',
                        overflow: 'hidden',
                      }}>
                        {/* Scrolling stripe overlay */}
                        <div className="gantt-neon-stripe" style={{
                          position: 'absolute',
                          top: 0, left: 0, right: 0, bottom: 0,
                          backgroundImage: 'linear-gradient(45deg, rgba(255,255,255,0.25) 25%, transparent 25%, transparent 50%, rgba(255,255,255,0.25) 50%, rgba(255,255,255,0.25) 75%, transparent 75%, transparent)',
                          backgroundSize: '20px 20px',
                          pointerEvents: 'none',
                        }} />
                        {showInside && durText && (
                          <span style={{
                            fontSize: 10, color: '#fff', whiteSpace: 'nowrap',
                            textShadow: '0 0 6px rgba(22,119,255,0.8), 0 1px 3px rgba(0,0,0,0.9)',
                            position: 'absolute', zIndex: 2,
                            left: 4, top: '50%', transform: 'translateY(-50%)',
                          }}>
                            {durText}
                          </span>
                        )}
                      </div>


                      {!showInside && durText && (
                        <span style={{
                          position: 'absolute',
                          left: bar.left + bar.width + 4,
                          top: bar.top,
                          lineHeight: bar.height + 'px',
                          fontSize: 10,
                          color: 'rgba(255,255,255,0.7)',
                          whiteSpace: 'nowrap',
                          textShadow: '0 1px 3px rgba(0,0,0,0.9)',
                          transition: 'left 0.5s ease-out',
                        }}>
                          {durText}
                        </span>
                      )}

                    </React.Fragment>
                  );
                })}
                {/* Single neon cursor — only for the current running bar */}
                {(() => {
                  const cur = runningBars.find(b => b.isCurrent);
                  if (!cur) return null;
                  return (
                    <div className="gantt-neon-cursor" style={{
                      position: 'absolute',
                      left: cur.left + cur.width - 1,
                      top: cur.top - 1,
                      width: 3,
                      height: cur.height + 2,
                      transition: 'left 0.1s linear',
                      background: 'rgba(200,230,255,0.9)',
                      borderRadius: 2,
                      zIndex: 10,
                      pointerEvents: 'none',
                    }} />
                  );
                })()}
              </div>
            )}
          </div>
        </div>
      </div>


    </div>
  );
}
