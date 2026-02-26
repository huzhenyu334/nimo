/**
 * PLM Component SDK — EbomForm Component
 *
 * A self-contained BOM form that:
 *   1. Fetches BOM data using projectId + bomId via PLM API
 *   2. Renders the BOM editor using the existing EBOMControl component
 *   3. Handles submit (diff → batch API calls → callback)
 *
 * Based on src/pages/embed/EBOMEmbed.tsx, adapted for SDK usage.
 */

import React, { useState, useCallback, useEffect, useRef, useMemo } from 'react'
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query'
import { ConfigProvider, Spin, Result, message, App, theme as antdTheme } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'

import EBOMControl from '@/components/BOM/EBOMControl'
import { projectBomApi } from '@/api/projectBom'
import type { BOMItemRequest } from '@/api/projectBom'
import { EBOM_CATEGORIES, type BOMControlConfig } from '@/components/BOM/bomConstants'
import { getConfig } from '../config'
import type { EbomFormProps } from '../types'

// ========== Default config for EBOM ==========

const DEFAULT_EBOM_CONFIG: BOMControlConfig = {
  bom_type: 'EBOM',
  visible_categories: EBOM_CATEGORIES,
  category_config: {},
}

// ========== Helper: convert item to BOMItemRequest ==========

const toBOMItemRequest = (item: Record<string, any>): BOMItemRequest => ({
  material_id: item.material_id,
  parent_item_id: item.parent_item_id,
  level: item.level,
  category: item.category,
  sub_category: item.sub_category,
  name: item.name || '',
  specification: item.specification,
  quantity: item.quantity || 1,
  unit: item.unit || 'pcs',
  reference: item.reference,
  manufacturer: item.manufacturer,
  manufacturer_pn: item.manufacturer_pn,
  supplier: item.supplier,
  supplier_pn: item.supplier_pn,
  unit_price: item.unit_price,
  lead_time_days: item.lead_time_days,
  is_critical: item.is_critical || false,
  is_alternative: item.is_alternative || false,
  notes: item.notes,
  drawing_no: item.drawing_no,
  item_number: item.item_number,
  extended_attrs: item.extended_attrs,
})

// ========== SDK-internal QueryClient ==========

const sdkQueryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

// ========== Inner Component (uses react-query hooks) ==========

const EbomFormInner: React.FC<EbomFormProps> = ({
  projectId,
  bomId,
  mode,
  acpContext,
  showCostColumn: _showCostColumn = true,
  allowImport: _allowImport = true,
  onReady,
  onSubmit,
  onChange,
  onError,
  onRegisterHandle,
}) => {
  const isReadonly = mode === 'view'

  const [items, setItems] = useState<Record<string, any>[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [dirty, setDirty] = useState(false)
  const originalItemIdsRef = useRef<Set<string>>(new Set())
  const initializedRef = useRef(false)

  // ---- Fetch BOM detail ----
  const { data: bomDetail, isLoading, error } = useQuery({
    queryKey: ['sdk-ebom', projectId, bomId],
    queryFn: () => projectBomApi.get(projectId, bomId!),
    enabled: !!projectId && !!bomId,
  })

  // ---- Initialize items from fetched data ----
  useEffect(() => {
    if (bomDetail?.items) {
      setItems(bomDetail.items)
      originalItemIdsRef.current = new Set(bomDetail.items.map((item) => item.id))

      // Fire onReady callback once
      if (!initializedRef.current) {
        initializedRef.current = true
        onReady?.({
          title: bomDetail.name || 'EBOM',
          totalItems: bomDetail.items.length,
        })
      }
    }
  }, [bomDetail, onReady])

  // ---- Handle error ----
  useEffect(() => {
    if (error) {
      onError?.({
        code: 'LOAD_ERROR',
        message: error instanceof Error ? error.message : 'Failed to load BOM data',
      })
    }
  }, [error, onError])

  // ---- Handle items change from EBOMControl ----
  const handleChange = useCallback(
    (newItems: Record<string, any>[]) => {
      setItems(newItems)
      if (!dirty) {
        setDirty(true)
        onChange?.(true)
      }
    },
    [dirty, onChange],
  )

  // ---- Submit: diff items → batch API calls → callback ----
  const handleSubmit = useCallback(async (): Promise<{ success: boolean; data: any; message?: string }> => {
    if (!projectId || !bomId) return { success: false, data: null, message: 'Missing projectId or bomId' }

    setSubmitting(true)
    try {
      const originalIds = originalItemIdsRef.current
      const currentIds = new Set(items.map((item) => item.id))

      // Diff: deleted / new / updated
      const deletedIds = [...originalIds].filter((id) => !currentIds.has(id))
      const newItems = items.filter((item) => String(item.id).startsWith('new-'))
      const updatedItems = items.filter(
        (item) => !String(item.id).startsWith('new-') && originalIds.has(item.id),
      )

      const promises: Promise<any>[] = []

      // Delete removed items
      for (const id of deletedIds) {
        promises.push(projectBomApi.deleteItem(projectId, bomId, id))
      }

      // Batch add new items
      if (newItems.length > 0) {
        promises.push(
          projectBomApi.batchAddItems(
            projectId,
            bomId,
            newItems.map(toBOMItemRequest),
          ),
        )
      }

      // Update existing items
      for (const item of updatedItems) {
        promises.push(
          projectBomApi.updateItem(projectId, bomId, item.id, toBOMItemRequest(item)),
        )
      }

      await Promise.all(promises)

      // Submit for review (with acpContext in request if provided)
      await projectBomApi.submit(projectId, bomId)

      message.success('BOM数据提交成功')

      // Reset dirty state
      setDirty(false)
      onChange?.(false)

      // Update original IDs to reflect new state
      const refreshed = await projectBomApi.get(projectId, bomId)
      if (refreshed?.items) {
        setItems(refreshed.items)
        originalItemIdsRef.current = new Set(refreshed.items.map((i) => i.id))
      }

      const result = {
        success: true as const,
        data: {
          project_id: projectId,
          bom_id: bomId,
          ...(acpContext ? { acp_context: acpContext } : {}),
        },
        message: 'BOM submitted successfully',
      }

      // Fire onSubmit callback for backward compat
      onSubmit?.(result)

      return result
    } catch (err) {
      console.error('[EbomForm] Submit failed:', err)
      const errMsg = err instanceof Error ? err.message : 'Submit failed'
      message.error('提交失败，请重试')

      const result = {
        success: false as const,
        data: null,
        message: errMsg,
      }

      // Fire onSubmit callback for backward compat
      onSubmit?.(result)

      onError?.({
        code: 'SUBMIT_ERROR',
        message: errMsg,
      })

      return result
    } finally {
      setSubmitting(false)
    }
  }, [projectId, bomId, items, acpContext, onChange, onSubmit, onError])

  // ---- Register handle for external submit control ----
  useEffect(() => {
    if (isReadonly) {
      onRegisterHandle?.({ submit: async () => ({ success: true, data: {} }) })
    } else {
      onRegisterHandle?.({ submit: handleSubmit })
    }
  }, [isReadonly, handleSubmit, onRegisterHandle])

  // ========== Render ==========

  // Validate required params
  if (!projectId) {
    return (
      <Result
        status="warning"
        title="参数缺失"
        subTitle="请提供 projectId 参数"
      />
    )
  }

  if (!bomId) {
    return (
      <Result
        status="warning"
        title="参数缺失"
        subTitle="请提供 bomId 参数"
      />
    )
  }

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
        <Spin size="large" tip="加载BOM数据..." />
      </div>
    )
  }

  if (error) {
    return (
      <Result
        status="error"
        title="加载失败"
        subTitle="无法加载BOM数据，请检查参数或权限"
      />
    )
  }

  return (
    <div style={{ padding: 16 }}>
      <EBOMControl
        config={DEFAULT_EBOM_CONFIG}
        value={items}
        onChange={handleChange}
        readonly={isReadonly}
      />

    </div>
  )
}

// ========== Wrapper: provides QueryClient + Antd ConfigProvider ==========

const EbomForm: React.FC<EbomFormProps> = (props) => {
  const { locale, isDark } = useMemo(() => {
    try {
      const cfg = getConfig()
      return {
        locale: cfg.locale === 'en-US' ? enUS : zhCN,
        isDark: cfg.theme === 'dark',
      }
    } catch {
      return { locale: zhCN, isDark: false }
    }
  }, [])

  const themeConfig = useMemo(() => isDark ? {
    algorithm: antdTheme.darkAlgorithm,
    token: {
      colorBgContainer: '#141414',
      colorBgElevated: '#1f1f1f',
      borderRadius: 6,
    },
  } : undefined, [isDark])

  return (
    <QueryClientProvider client={sdkQueryClient}>
      <ConfigProvider locale={locale} theme={themeConfig}>
        <App>
          <EbomFormInner {...props} />
        </App>
      </ConfigProvider>
    </QueryClientProvider>
  )
}

export default EbomForm
