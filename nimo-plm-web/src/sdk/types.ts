/**
 * PLM Component SDK — Type Definitions
 *
 * Modeled after Feishu Open Platform Web Component pattern:
 *   script -> config() -> render()
 */

// ========== SDK Global Config ==========

export interface SDKConfig {
  /** Application ID for identification */
  appId: string
  /** Bearer token for PLM API authentication (optional for cross-system callers) */
  token?: string
  /** PLM API base URL, e.g. "https://plm.example.com" */
  baseUrl: string
  /** List of component names to enable, e.g. ["EbomForm"] */
  componentList: string[]
  /** Locale: "zh-CN" | "en-US" (default: "zh-CN") */
  locale?: string
  /** Theme: "light" | "dark" (default: "light") */
  theme?: string
}

// ========== EbomForm Component Props ==========

export interface EbomFormProps {
  /** Project ID to load BOM data */
  projectId: string
  /** BOM ID to load (required for fetching data) */
  bomId?: string
  /** Mode: 'edit' allows modification, 'view' is read-only */
  mode: 'edit' | 'view'
  /** ACP integration context */
  acpContext?: {
    instanceId: string
    stepName: string
  }
  /** Whether to show cost columns (default: true) */
  showCostColumn?: boolean
  /** Whether to allow Excel import (default: true in edit mode) */
  allowImport?: boolean
  /** Whether BOM categories are expanded by default (default: true) */
  defaultExpanded?: boolean
  /** Called when component finishes loading data */
  onReady?: (info: { title: string; totalItems: number }) => void
  /** Called after submit completes */
  onSubmit?: (result: { success: boolean; data: any; message?: string }) => void
  /** Called when data dirty state changes */
  onChange?: (dirty: boolean) => void
  /** Called on error */
  onError?: (error: { code: string; message: string }) => void
  /** Register a handle for external submit control (used by ACP) */
  onRegisterHandle?: (handle: { submit: () => Promise<{ success: boolean; data: any; message?: string }> }) => void
}

// ========== Render Result / Component Handle ==========

export interface RenderResult {
  /** Trigger form submission. Returns result object. */
  submit: () => Promise<{ success: boolean; data: any; message?: string }>
  /** Unmount and destroy the rendered component */
  destroy: () => void
}

export interface ComponentHandle {
  /** Trigger form submission. Returns result object. */
  submit: () => Promise<{ success: boolean; data: any; message?: string }>
  /** Unmount and destroy the rendered component */
  destroy: () => void
}
