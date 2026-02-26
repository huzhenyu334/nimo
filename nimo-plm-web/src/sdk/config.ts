/**
 * PLM Component SDK — Global Configuration Store
 *
 * Stores the config() call's token/baseUrl and provides getters
 * for components and the SDK-internal API client.
 */

import type { SDKConfig } from './types'

// ========== Internal state ==========

let _config: SDKConfig | null = null
let _configured = false

// ========== Public API ==========

/**
 * Save SDK configuration. Must be called before render().
 * Also injects the token into localStorage so the existing apiClient
 * (which reads from localStorage) works without modification.
 */
export function setConfig(config: SDKConfig): void {
  _config = { ...config }
  _configured = true

  // Inject token into localStorage so existing apiClient interceptor picks it up
  if (config.token) {
    localStorage.setItem('access_token', config.token)
  }
}

/**
 * Get the current SDK configuration. Throws if not configured.
 */
export function getConfig(): SDKConfig {
  if (!_configured || !_config) {
    throw new Error('[nimoComponent] SDK not configured. Call nimoComponent.config() first.')
  }
  return _config
}

/**
 * Check if the SDK has been configured.
 */
export function isConfigured(): boolean {
  return _configured
}

/**
 * Get the API base URL. Combines baseUrl + /api/v1 prefix.
 * If baseUrl is provided, we override the apiClient's baseURL.
 */
export function getApiBaseUrl(): string {
  if (!_config?.baseUrl) return '/api/v1'
  // Remove trailing slash
  const base = _config.baseUrl.replace(/\/+$/, '')
  return `${base}/api/v1`
}

/**
 * Get the auth token.
 */
export function getToken(): string {
  return _config?.token || ''
}

/**
 * Check if a component is enabled in componentList.
 */
export function isComponentEnabled(name: string): boolean {
  if (!_config?.componentList) return false
  return _config.componentList.includes(name)
}

/**
 * Update the existing axios apiClient to use SDK's baseUrl.
 * This is called once during config() to reconfigure the shared apiClient.
 */
export async function configureApiClient(): Promise<void> {
  if (!_config) return

  // Dynamically import the existing apiClient and reconfigure its baseURL
  const { apiClient } = await import('@/api/client')

  if (_config.baseUrl) {
    const base = _config.baseUrl.replace(/\/+$/, '')
    apiClient.defaults.baseURL = `${base}/api/v1`
  }

  // The token is already in localStorage from setConfig(),
  // and the request interceptor reads it from there.
}
