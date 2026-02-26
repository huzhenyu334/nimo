/**
 * PLM Component SDK — Entry Point
 *
 * Usage pattern (similar to Feishu Open Platform Web Components):
 *
 *   <script src="https://cdn.example.com/nimo-plm-components.js"></script>
 *   <script>
 *     // Step 1: Configure
 *     nimoComponent.config({
 *       appId: 'your-app-id',
 *       token: 'bearer-token',
 *       baseUrl: 'https://plm-api.example.com',
 *       componentList: ['EbomForm'],
 *     }).then(() => {
 *       // Step 2: Render
 *       const result = nimoComponent.render('EbomForm', {
 *         projectId: 'proj-123',
 *         bomId: 'bom-456',
 *         mode: 'edit',
 *         onSubmit: (result) => console.log('Submitted:', result),
 *       }, document.getElementById('container'));
 *
 *       // Step 3: Cleanup when done
 *       // result.destroy();
 *     });
 *   </script>
 */

import React from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { setConfig, getConfig, configureApiClient, isConfigured, isComponentEnabled } from './config'
import type { SDKConfig, EbomFormProps, RenderResult, ComponentHandle } from './types'

// SDK version — injected at build time or fallback
const SDK_VERSION = '__SDK_VERSION__'

// ========== Component registry ==========

type ComponentModule = {
  default: React.ComponentType<any>
}

const COMPONENT_REGISTRY: Record<string, () => Promise<ComponentModule>> = {
  EbomForm: () => import('./components/EbomForm'),
}

// Cache loaded components
const loadedComponents: Record<string, React.ComponentType<any>> = {}

// Track active roots for cleanup
const activeRoots: Map<HTMLElement, Root> = new Map()

// ========== config() ==========

async function config(sdkConfig: SDKConfig): Promise<void> {
  // Validate required fields
  if (!sdkConfig.appId) throw new Error('[nimoComponent] config: appId is required')
  // token is optional — cross-system callers may use other auth mechanisms
  if (!sdkConfig.baseUrl) throw new Error('[nimoComponent] config: baseUrl is required')
  if (!sdkConfig.componentList || sdkConfig.componentList.length === 0) {
    throw new Error('[nimoComponent] config: componentList is required and must not be empty')
  }

  // Validate component names
  for (const name of sdkConfig.componentList) {
    if (!COMPONENT_REGISTRY[name]) {
      console.warn(`[nimoComponent] Unknown component: "${name}". Available: ${Object.keys(COMPONENT_REGISTRY).join(', ')}`)
    }
  }

  // Store config + inject token into localStorage + reconfigure apiClient
  setConfig(sdkConfig)
  await configureApiClient()

  // Pre-load enabled components
  const loadPromises = sdkConfig.componentList
    .filter((name) => COMPONENT_REGISTRY[name])
    .map(async (name) => {
      const mod = await COMPONENT_REGISTRY[name]()
      loadedComponents[name] = mod.default
    })

  await Promise.all(loadPromises)

  console.log(`[nimoComponent] SDK v${SDK_VERSION} configured. Components: ${sdkConfig.componentList.join(', ')}`)
}

// ========== render() ==========

function render(
  componentName: string,
  props: EbomFormProps | Record<string, any>,
  container: HTMLElement,
): RenderResult {
  // Validate state
  if (!isConfigured()) {
    throw new Error('[nimoComponent] SDK not configured. Call nimoComponent.config() first.')
  }

  if (!isComponentEnabled(componentName)) {
    throw new Error(
      `[nimoComponent] Component "${componentName}" is not in componentList. ` +
      `Add it to the config({ componentList: [...] }) call.`,
    )
  }

  const Component = loadedComponents[componentName]
  if (!Component) {
    throw new Error(
      `[nimoComponent] Component "${componentName}" not loaded. ` +
      `This usually means config() hasn't finished yet.`,
    )
  }

  if (!container || !(container instanceof HTMLElement)) {
    throw new Error('[nimoComponent] render: container must be a valid HTMLElement')
  }

  // If the container already has an active root, destroy it first
  const existingRoot = activeRoots.get(container)
  if (existingRoot) {
    existingRoot.unmount()
    activeRoots.delete(container)
  }

  // Mutable ref to capture the component's submit handle
  let handleRef: { submit: () => Promise<any> } | null = null

  // Inject dark theme CSS overrides if configured
  let darkStyleEl: HTMLStyleElement | null = null
  if (getConfig()?.theme === 'dark') {
    darkStyleEl = document.createElement('style')
    darkStyleEl.setAttribute('data-nimo-dark', 'true')
    darkStyleEl.textContent = `
      /* SDK dark theme overrides for BOM components */
      [data-nimo-sdk] { color-scheme: dark; }
      [data-nimo-sdk] .ant-table { background: #141414 !important; }
      [data-nimo-sdk] .ant-table-thead > tr > th,
      [data-nimo-sdk] .ant-table-thead > tr > td { background: #1d1d1d !important; color: rgba(255,255,255,0.85) !important; border-color: #303030 !important; }
      [data-nimo-sdk] .ant-table-tbody > tr > td { border-color: #303030 !important; }
      [data-nimo-sdk] .ant-table-tbody > tr:hover > td { background: #262626 !important; }
      [data-nimo-sdk] .ant-table-tbody > tr.ant-table-row:hover > td { background: #262626 !important; }
      [data-nimo-sdk] .ant-collapse > .ant-collapse-item > .ant-collapse-header { color: rgba(255,255,255,0.85) !important; }
      [data-nimo-sdk] .ant-collapse-content { background: #141414 !important; border-color: #303030 !important; }
      [data-nimo-sdk] .ant-btn-default { background: #1d1d1d !important; border-color: #424242 !important; color: rgba(255,255,255,0.85) !important; }
      [data-nimo-sdk] .ant-select-selector { background: #1d1d1d !important; border-color: #424242 !important; color: rgba(255,255,255,0.85) !important; }
      [data-nimo-sdk] .ant-input { background: #1d1d1d !important; border-color: #424242 !important; color: rgba(255,255,255,0.85) !important; }
      [data-nimo-sdk] .ant-input-number { background: #1d1d1d !important; border-color: #424242 !important; color: rgba(255,255,255,0.85) !important; }
      [data-nimo-sdk] .ant-tag { border-color: #424242 !important; }
      [data-nimo-sdk] [style*="background: #fafafa"],
      [data-nimo-sdk] [style*="background:#fafafa"],
      [data-nimo-sdk] [style*="background: rgb(250"] { background: #1d1d1d !important; }
      [data-nimo-sdk] [style*="background: #fff7e6"] { background: #2a2000 !important; }
      [data-nimo-sdk] [style*="background: #f6ffed"] { background: #0a2000 !important; }
      [data-nimo-sdk] [style*="background: #e6f7ff"],
      [data-nimo-sdk] [style*="background: #e6f4ff"] { background: #001d3d !important; }
      [data-nimo-sdk] [style*="background: #fff0f6"] { background: #2a0012 !important; }
      [data-nimo-sdk] [style*="background: #f9f0ff"] { background: #1a002a !important; }
      [data-nimo-sdk] [style*="background: #e6fffb"] { background: #002a20 !important; }
      [data-nimo-sdk] [style*="background: #fff2e8"] { background: #2a1500 !important; }
      [data-nimo-sdk] [style*="border-bottom: 1px solid #e8e8e8"],
      [data-nimo-sdk] [style*="border-bottom: 1px solid rgb(232"] { border-color: #303030 !important; }
    `
    document.head.appendChild(darkStyleEl)
    container.setAttribute('data-nimo-sdk', 'true')
  }

  // Create React root and render with onRegisterHandle to capture the submit function
  const root = createRoot(container)
  const enhancedProps = {
    ...props,
    onRegisterHandle: (h: { submit: () => Promise<any> }) => {
      handleRef = h
    },
  }
  root.render(React.createElement(Component, enhancedProps))
  activeRoots.set(container, root)

  // Return ComponentHandle with submit + destroy
  return {
    async submit() {
      if (handleRef) return handleRef.submit()
      return { success: false, data: null, message: 'Component not ready' }
    },
    destroy() {
      root.unmount()
      activeRoots.delete(container)
      if (darkStyleEl) {
        darkStyleEl.remove()
        darkStyleEl = null
      }
      container.removeAttribute('data-nimo-sdk')
      container.innerHTML = ''
    },
  }
}

// ========== Expose on window ==========

interface NimoComponentSDK {
  version: string
  config: (sdkConfig: SDKConfig) => Promise<void>
  render: (
    componentName: string,
    props: EbomFormProps | Record<string, any>,
    container: HTMLElement,
  ) => RenderResult
}

const nimoComponent: NimoComponentSDK = {
  version: SDK_VERSION,
  config,
  render,
}

// Attach to window
;(window as any).nimoComponent = nimoComponent

// Also export for ESM usage
export default nimoComponent
export { config, render }
export type { SDKConfig, EbomFormProps, RenderResult, ComponentHandle }
