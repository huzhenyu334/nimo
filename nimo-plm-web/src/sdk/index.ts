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
import { setConfig, configureApiClient, isConfigured, isComponentEnabled } from './config'
import type { SDKConfig, EbomFormProps, RenderResult } from './types'

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

  // Create React root and render
  const root = createRoot(container)
  root.render(React.createElement(Component, props))
  activeRoots.set(container, root)

  // Return destroy handle
  return {
    destroy() {
      root.unmount()
      activeRoots.delete(container)
      // Clear container contents
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
export type { SDKConfig, EbomFormProps, RenderResult }
