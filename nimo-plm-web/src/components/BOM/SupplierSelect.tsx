import React, { useState, useMemo, useCallback } from 'react';
import { Select, Typography } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { srmApi } from '@/api/srm';

const { Text } = Typography;

interface SupplierSelectProps {
  value?: string;            // supplier_id
  displayName?: string;      // current supplier name (for display before data loads)
  onChange: (supplierId: string | null, supplierName: string) => void;
  category?: string;         // filter by specific category
  categoryNe?: string;       // exclude specific category
  readonly?: boolean;
  placeholder?: string;
  style?: React.CSSProperties;
}

const SupplierSelect: React.FC<SupplierSelectProps> = ({
  value,
  displayName,
  onChange,
  category,
  categoryNe,
  readonly = false,
  placeholder = '选择供应商',
  style,
}) => {
  const [search, setSearch] = useState('');

  // Debounced search query
  const { data: suppliers = [], isFetching } = useQuery({
    queryKey: ['supplier-search', search, category, categoryNe],
    queryFn: async () => {
      const params: Record<string, string | number> = { page: 1, page_size: 20 };
      if (search) params.search = search;
      if (category) params.category = category;
      if (categoryNe) params.category_ne = categoryNe;
      const result = await srmApi.listSuppliers(params as any);
      return result.items || [];
    },
    staleTime: 30_000,
  });

  const options = useMemo(() =>
    suppliers.map(s => ({
      label: s.short_name || s.name,
      value: s.id,
      name: s.name,
      short_name: s.short_name,
    })),
    [suppliers],
  );

  const handleSearch = useCallback((val: string) => {
    setSearch(val);
  }, []);

  const handleChange = useCallback((selectedId: string | null) => {
    if (!selectedId) {
      onChange(null, '');
      return;
    }
    const supplier = suppliers.find(s => s.id === selectedId);
    onChange(selectedId, supplier?.short_name || supplier?.name || '');
  }, [suppliers, onChange]);

  if (readonly) {
    return (
      <Text style={{ fontSize: 12, ...(style || {}) }}>
        {displayName || '-'}
      </Text>
    );
  }

  return (
    <Select
      showSearch
      allowClear
      value={value || undefined}
      placeholder={placeholder}
      style={{ width: '100%', minWidth: 80, ...(style || {}) }}
      size="small"
      filterOption={false}
      onSearch={handleSearch}
      onChange={handleChange}
      loading={isFetching}
      options={options}
      notFoundContent={isFetching ? '搜索中...' : '无匹配供应商'}
      popupMatchSelectWidth={200}
    />
  );
};

export default SupplierSelect;
