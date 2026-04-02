package erp

import sdk "github.com/bitfantasy/acp-module-sdk"

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func intPtr(i int) *int { return &i }

// ---------------------------------------------------------------------------
// customerEntity — erp_customers
// ---------------------------------------------------------------------------

func customerEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "customers",
		Label:      "客户",
		Table:      "erp_customers",
		PrimaryKey: "id",
		Icon:       "TeamOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "客户编码", Type: "string", ReadOnly: true, Unique: true, Width: 100},
			{Name: "name", Label: "客户名称", Type: "string", Required: true},
			{Name: "short_name", Label: "简称", Type: "string", Width: 80, Component: "virtual"},
			{
				Name: "type", Label: "类型", Type: "select", Default: "enterprise",
				Options: []sdk.FieldOption{
					{Value: "enterprise", Label: "企业", Color: "blue"},
					{Value: "individual", Label: "个人", Color: "green"},
					{Value: "distributor", Label: "经销商", Color: "purple"},
					{Value: "oem", Label: "OEM", Color: "orange"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{Name: "tax_id", Label: "税号", Type: "string", HideInList: true},
			{Name: "contact_name", Label: "联系人", Type: "string", Width: 90},
			{Name: "contact_phone", Label: "联系电话", Type: "string", Width: 120},
			{Name: "contact_email", Label: "邮箱", Type: "string", HideInList: true},
			{Name: "billing_address", Label: "开票地址", Type: "text", HideInList: true},
			{Name: "shipping_address", Label: "收货地址", Type: "json", HideInList: true},
			{
				Name: "payment_terms", Label: "付款条件", Type: "select", Default: "NET30",
				Options: []sdk.FieldOption{
					{Value: "NET30", Label: "NET30", Color: "blue"},
					{Value: "NET60", Label: "NET60", Color: "cyan"},
					{Value: "COD", Label: "货到付款", Color: "orange"},
					{Value: "prepaid", Label: "预付", Color: "green"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "credit_limit", Label: "信用额度", Type: "number", Precision: intPtr(2), Unit: "¥", HideInList: true},
			{
				Name: "currency", Label: "默认币种", Type: "select", Default: "CNY",
				Options: []sdk.FieldOption{
					{Value: "CNY", Label: "CNY", Color: "red"},
					{Value: "USD", Label: "USD", Color: "green"},
				},
				Width: 70,
			},
			{
				Name: "status", Label: "状态", Type: "select", Default: "active",
				Options: []sdk.FieldOption{
					{Value: "active", Label: "活跃", Color: "green"},
					{Value: "inactive", Label: "停用", Color: "default"},
					{Value: "blocked", Label: "冻结", Color: "red"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
			{Name: "tags", Label: "标签", Type: "json", HideInList: true},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
			{Name: "updated_at", Label: "更新时间", Type: "datetime", ReadOnly: true, HideInList: true},
		},
		ListColumns:  []string{"code", "name", "short_name", "type", "contact_name", "contact_phone", "payment_terms", "status", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"name", "code", "short_name", "contact_name"},
		Filters:      []string{"type", "status", "payment_terms"},
	}
}

// ---------------------------------------------------------------------------
// warehouseEntity — erp_warehouses
// ---------------------------------------------------------------------------

func warehouseEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "warehouses",
		Label:      "仓库",
		Table:      "erp_warehouses",
		PrimaryKey: "id",
		Icon:       "HomeOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "仓库编码", Type: "string", ReadOnly: true, Unique: true, Width: 80},
			{Name: "name", Label: "仓库名称", Type: "string", Required: true},
			{
				Name: "type", Label: "类型", Type: "select", Required: true,
				Options: []sdk.FieldOption{
					{Value: "raw_material", Label: "原材料仓", Color: "blue"},
					{Value: "semi_finished", Label: "半成品仓", Color: "cyan"},
					{Value: "finished_goods", Label: "成品仓", Color: "green"},
					{Value: "returns", Label: "退货仓", Color: "orange"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  100,
			},
			{Name: "address", Label: "地址", Type: "text", HideInList: true},
			{Name: "manager_id", Label: "仓管员", Type: "string", Width: 90},
			{Name: "is_default", Label: "默认仓库", Type: "boolean"},
			{
				Name: "status", Label: "状态", Type: "select", Default: "active",
				Options: []sdk.FieldOption{
					{Value: "active", Label: "启用", Color: "green"},
					{Value: "inactive", Label: "停用", Color: "default"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
		},
		ListColumns:  []string{"code", "name", "type", "manager_id", "is_default", "status"},
		DefaultSort:  "code",
		DefaultOrder: "asc",
		Searchable:   []string{"name", "code"},
		Filters:      []string{"type", "status"},
	}
}

// ---------------------------------------------------------------------------
// locationEntity — erp_locations
// ---------------------------------------------------------------------------

func locationEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "locations",
		Label:      "库位",
		Table:      "erp_locations",
		PrimaryKey: "id",
		Icon:       "EnvironmentOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "warehouse_id", Label: "所属仓库", Type: "relation", Required: true, RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name"},
			{Name: "code", Label: "库位编码", Type: "string", ReadOnly: true, Unique: true, Width: 100},
			{Name: "name", Label: "库位名称", Type: "string", Required: true},
			{
				Name: "type", Label: "类型", Type: "select", Default: "storage",
				Options: []sdk.FieldOption{
					{Value: "storage", Label: "存储", Color: "blue"},
					{Value: "picking", Label: "拣货", Color: "green"},
					{Value: "staging", Label: "暂存", Color: "orange"},
					{Value: "quarantine", Label: "隔离", Color: "red"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{Name: "capacity", Label: "容量上限", Type: "integer"},
			{
				Name: "status", Label: "状态", Type: "select", Default: "active",
				Options: []sdk.FieldOption{
					{Value: "active", Label: "启用", Color: "green"},
					{Value: "inactive", Label: "停用", Color: "default"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
		},
		ListColumns:  []string{"code", "name", "warehouse_id", "type", "capacity", "status"},
		DefaultSort:  "code",
		DefaultOrder: "asc",
		Searchable:   []string{"name", "code"},
		Filters:      []string{"warehouse_id", "type", "status"},
	}
}

// ---------------------------------------------------------------------------
// quotationEntity — erp_quotations
// ---------------------------------------------------------------------------

func quotationEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "quotations",
		Label:      "销售报价",
		Table:      "erp_quotations",
		PrimaryKey: "id",
		Icon:       "FileTextOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "报价编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "customer_id", Label: "客户", Type: "relation", Required: true, RefEntity: "customers", RefModule: "erp", RefDisplay: "name", Width: 120},
			{Name: "contact_name", Label: "联系人", Type: "string", HideInList: true},
			{
				Name: "currency", Label: "币种", Type: "select", Default: "CNY",
				Options: []sdk.FieldOption{
					{Value: "CNY", Label: "CNY", Color: "red"},
					{Value: "USD", Label: "USD", Color: "green"},
				},
				Width: 70,
			},
			{Name: "exchange_rate", Label: "汇率", Type: "number", Precision: intPtr(6), HideInList: true},
			{Name: "subtotal", Label: "小计", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, Width: 100},
			{Name: "tax_amount", Label: "税额", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, HideInList: true},
			{Name: "total", Label: "总计", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, Width: 110},
			{Name: "valid_until", Label: "有效期", Type: "date", Width: 110},
			{
				Name: "payment_terms", Label: "付款条件", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "NET30", Label: "NET30", Color: "blue"},
					{Value: "NET60", Label: "NET60", Color: "cyan"},
					{Value: "COD", Label: "货到付款", Color: "orange"},
					{Value: "prepaid", Label: "预付", Color: "green"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
				HideInList: true,
			},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
			{
				Name: "status", Label: "状态", Type: "select", Default: "draft",
				Options: []sdk.FieldOption{
					{Value: "draft", Label: "草稿", Color: "default"},
					{Value: "sent", Label: "已发送", Color: "processing"},
					{Value: "accepted", Label: "已接受", Color: "success"},
					{Value: "rejected", Label: "已拒绝", Color: "error"},
					{Value: "expired", Label: "已过期", Color: "warning"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{Name: "created_by", Label: "创建人", Type: "string", ReadOnly: true, Width: 90, HideInForm: true},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"code", "customer_id", "currency", "subtotal", "total", "valid_until", "status", "created_by", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code"},
		Filters:      []string{"customer_id", "status", "currency"},
		Relations: []sdk.RelationDef{
			{Name: "items", Label: "报价明细", Type: "has_many", Target: "quotation_items", ForeignKey: "quotation_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// quotationItemEntity — erp_quotation_items
// ---------------------------------------------------------------------------

func quotationItemEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "quotation_items",
		Label:      "报价明细",
		Table:      "erp_quotation_items",
		PrimaryKey: "id",
		Icon:       "UnorderedListOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "quotation_id", Label: "报价单", Type: "relation", Required: true, RefEntity: "quotations", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "product_id", Label: "产品", Type: "relation", Required: true, RefEntity: "products", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "sku_id", Label: "SKU", Type: "string", HideInList: true},
			{Name: "description", Label: "描述", Type: "string", Width: 160},
			{Name: "quantity", Label: "数量", Type: "number", Required: true, Precision: intPtr(4), Width: 80},
			{Name: "unit_price", Label: "单价", Type: "number", Required: true, Precision: intPtr(4), Unit: "¥", Width: 100},
			{Name: "discount_pct", Label: "折扣%", Type: "number", Precision: intPtr(2), Width: 70},
			{Name: "tax_rate", Label: "税率%", Type: "number", Precision: intPtr(2), Width: 70},
			{Name: "line_total", Label: "行合计", Type: "number", ReadOnly: true, Precision: intPtr(2), Unit: "¥", Width: 100},
			{Name: "sort_order", Label: "排序", Type: "integer", HideInList: true, HideInForm: true},
		},
		ListColumns:  []string{"product_id", "description", "quantity", "unit_price", "discount_pct", "tax_rate", "line_total"},
		DefaultSort:  "sort_order",
		DefaultOrder: "asc",
		Searchable:   []string{"description"},
		Filters:      []string{"quotation_id"},
	}
}

// ---------------------------------------------------------------------------
// salesOrderEntity — erp_sales_orders
// ---------------------------------------------------------------------------

func salesOrderEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "sales_orders",
		Label:      "销售订单",
		Table:      "erp_sales_orders",
		PrimaryKey: "id",
		Icon:       "ShoppingCartOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "订单编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "quotation_id", Label: "来源报价", Type: "relation", RefEntity: "quotations", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "customer_id", Label: "客户", Type: "relation", Required: true, RefEntity: "customers", RefModule: "erp", RefDisplay: "name", Width: 120},
			{Name: "shipping_address", Label: "收货地址", Type: "text", HideInList: true},
			{
				Name: "currency", Label: "币种", Type: "select", Default: "CNY",
				Options: []sdk.FieldOption{
					{Value: "CNY", Label: "CNY", Color: "red"},
					{Value: "USD", Label: "USD", Color: "green"},
				},
				Width: 70,
			},
			{Name: "subtotal", Label: "小计", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, HideInList: true},
			{Name: "tax_amount", Label: "税额", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, HideInList: true},
			{Name: "total", Label: "总计", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, Width: 110},
			{
				Name: "payment_terms", Label: "付款条件", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "NET30", Label: "NET30", Color: "blue"},
					{Value: "NET60", Label: "NET60", Color: "cyan"},
					{Value: "COD", Label: "货到付款", Color: "orange"},
					{Value: "prepaid", Label: "预付", Color: "green"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
				HideInList: true,
			},
			{Name: "expected_date", Label: "期望交货日", Type: "date", Width: 110},
			{Name: "shipping_method", Label: "物流方式", Type: "string", HideInList: true},
			{
				Name: "priority", Label: "优先级", Type: "select", Default: "normal",
				Options: []sdk.FieldOption{
					{Value: "normal", Label: "普通", Color: "blue"},
					{Value: "urgent", Label: "紧急", Color: "orange"},
					{Value: "critical", Label: "加急", Color: "red"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{
				Name: "status", Label: "状态", Type: "select", Default: "draft",
				Options: []sdk.FieldOption{
					{Value: "draft", Label: "草稿", Color: "default"},
					{Value: "confirmed", Label: "已确认", Color: "processing"},
					{Value: "producing", Label: "生产中", Color: "blue"},
					{Value: "ready", Label: "待发货", Color: "cyan"},
					{Value: "shipped", Label: "已发货", Color: "purple"},
					{Value: "delivered", Label: "已签收", Color: "success"},
					{Value: "closed", Label: "已关闭", Color: "default"},
					{Value: "cancelled", Label: "已取消", Color: "error"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"draft":     "#d9d9d9",
						"confirmed": "#1677ff",
						"producing": "#0958d9",
						"ready":     "#13c2c2",
						"shipped":   "#722ed1",
						"delivered": "#52c41a",
						"closed":    "#8c8c8c",
						"cancelled": "#f5222d",
					},
				},
				Width: 80,
			},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
			{Name: "created_by", Label: "创建人", Type: "string", ReadOnly: true, Width: 90, HideInForm: true},
			{Name: "confirmed_at", Label: "确认时间", Type: "datetime", ReadOnly: true, HideInList: true},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"code", "customer_id", "total", "expected_date", "priority", "status", "created_by", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code"},
		Filters:      []string{"customer_id", "status", "priority"},
		Relations: []sdk.RelationDef{
			{Name: "items", Label: "订单明细", Type: "has_many", Target: "sales_order_items", ForeignKey: "order_id", Display: "table"},
			{Name: "shipments", Label: "发货单", Type: "has_many", Target: "shipments", ForeignKey: "order_id", Display: "table"},
			{Name: "invoices", Label: "发票", Type: "has_many", Target: "sales_invoices", ForeignKey: "order_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// salesOrderItemEntity — erp_sales_order_items
// ---------------------------------------------------------------------------

func salesOrderItemEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "sales_order_items",
		Label:      "订单明细",
		Table:      "erp_sales_order_items",
		PrimaryKey: "id",
		Icon:       "UnorderedListOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "order_id", Label: "销售订单", Type: "relation", Required: true, RefEntity: "sales_orders", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "product_id", Label: "产品", Type: "relation", Required: true, RefEntity: "products", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "sku_id", Label: "SKU", Type: "string", HideInList: true},
			{Name: "bom_id", Label: "BOM", Type: "relation", RefEntity: "boms", RefModule: "plm", RefDisplay: "name", HideInList: true},
			{Name: "description", Label: "描述", Type: "string", Width: 160},
			{Name: "quantity", Label: "订单数量", Type: "number", Required: true, Precision: intPtr(4), Width: 90},
			{Name: "delivered_qty", Label: "已发货数量", Type: "number", ReadOnly: true, Precision: intPtr(4), Width: 100},
			{Name: "unit_price", Label: "单价", Type: "number", Required: true, Precision: intPtr(4), Unit: "¥", Width: 100},
			{Name: "discount_pct", Label: "折扣%", Type: "number", Precision: intPtr(2), Width: 70},
			{Name: "tax_rate", Label: "税率%", Type: "number", Precision: intPtr(2), Width: 70},
			{Name: "line_total", Label: "行合计", Type: "number", ReadOnly: true, Precision: intPtr(2), Unit: "¥", Width: 100},
			{Name: "expected_date", Label: "行级交期", Type: "date", Width: 110},
			{
				Name: "status", Label: "状态", Type: "select", Default: "pending",
				Options: []sdk.FieldOption{
					{Value: "pending", Label: "待处理", Color: "default"},
					{Value: "allocated", Label: "已分配", Color: "processing"},
					{Value: "producing", Label: "生产中", Color: "blue"},
					{Value: "ready", Label: "待发货", Color: "cyan"},
					{Value: "shipped", Label: "已发货", Color: "purple"},
					{Value: "delivered", Label: "已签收", Color: "success"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
		},
		ListColumns:  []string{"product_id", "description", "quantity", "delivered_qty", "unit_price", "line_total", "expected_date", "status"},
		DefaultSort:  "created_at",
		DefaultOrder: "asc",
		Searchable:   []string{"description"},
		Filters:      []string{"order_id", "status"},
	}
}

// ---------------------------------------------------------------------------
// shipmentEntity — erp_shipments
// ---------------------------------------------------------------------------

func shipmentEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "shipments",
		Label:      "发货单",
		Table:      "erp_shipments",
		PrimaryKey: "id",
		Icon:       "SendOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "发货编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "order_id", Label: "销售订单", Type: "relation", Required: true, RefEntity: "sales_orders", RefModule: "erp", RefDisplay: "code", Width: 140},
			{Name: "warehouse_id", Label: "出库仓库", Type: "relation", Required: true, RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "shipping_address", Label: "收货地址", Type: "text", HideInList: true},
			{Name: "carrier", Label: "承运商", Type: "string", Width: 100},
			{Name: "tracking_no", Label: "物流单号", Type: "string", Width: 140},
			{Name: "shipped_at", Label: "发货时间", Type: "datetime", Width: 160},
			{Name: "delivered_at", Label: "签收时间", Type: "datetime", HideInList: true},
			{
				Name: "status", Label: "状态", Type: "select", Default: "draft",
				Options: []sdk.FieldOption{
					{Value: "draft", Label: "草稿", Color: "default"},
					{Value: "picking", Label: "拣货中", Color: "processing"},
					{Value: "packed", Label: "已打包", Color: "cyan"},
					{Value: "shipped", Label: "已发货", Color: "purple"},
					{Value: "delivered", Label: "已签收", Color: "success"},
					{Value: "returned", Label: "已退回", Color: "error"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"draft":     "#d9d9d9",
						"picking":   "#1677ff",
						"packed":    "#13c2c2",
						"shipped":   "#722ed1",
						"delivered": "#52c41a",
						"returned":  "#f5222d",
					},
				},
				Width: 80,
			},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
		},
		ListColumns:  []string{"code", "order_id", "warehouse_id", "carrier", "tracking_no", "shipped_at", "status"},
		DefaultSort:  "shipped_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "tracking_no"},
		Filters:      []string{"order_id", "warehouse_id", "status"},
		Relations: []sdk.RelationDef{
			{Name: "items", Label: "发货明细", Type: "has_many", Target: "shipment_items", ForeignKey: "shipment_id", Display: "table"},
			{Name: "oqc", Label: "出货检验", Type: "has_many", Target: "oqc_inspections", ForeignKey: "shipment_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// shipmentItemEntity — erp_shipment_items
// ---------------------------------------------------------------------------

func shipmentItemEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "shipment_items",
		Label:      "发货明细",
		Table:      "erp_shipment_items",
		PrimaryKey: "id",
		Icon:       "UnorderedListOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "shipment_id", Label: "发货单", Type: "relation", Required: true, RefEntity: "shipments", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "order_item_id", Label: "订单行", Type: "relation", Required: true, RefEntity: "sales_order_items", RefModule: "erp", RefDisplay: "id"},
			{Name: "product_id", Label: "产品", Type: "relation", RefEntity: "products", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "quantity", Label: "发货数量", Type: "number", Required: true, Precision: intPtr(4), Width: 90},
			{Name: "lot_number", Label: "批次号", Type: "string", Width: 100},
			{Name: "serial_numbers", Label: "序列号", Type: "json", HideInList: true},
		},
		ListColumns:  []string{"product_id", "quantity", "lot_number"},
		DefaultSort:  "id",
		DefaultOrder: "asc",
		Searchable:   []string{"lot_number"},
		Filters:      []string{"shipment_id"},
	}
}

// ---------------------------------------------------------------------------
// returnEntity — erp_returns
// ---------------------------------------------------------------------------

func returnEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "returns",
		Label:      "退货单",
		Table:      "erp_returns",
		PrimaryKey: "id",
		Icon:       "RollbackOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "退货编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "order_id", Label: "原始订单", Type: "relation", Required: true, RefEntity: "sales_orders", RefModule: "erp", RefDisplay: "code", Width: 140},
			{Name: "customer_id", Label: "客户", Type: "relation", Required: true, RefEntity: "customers", RefModule: "erp", RefDisplay: "name", Width: 120},
			{Name: "reason", Label: "退货原因", Type: "string", Required: true},
			{
				Name: "type", Label: "类型", Type: "select", Required: true,
				Options: []sdk.FieldOption{
					{Value: "refund", Label: "退款", Color: "red"},
					{Value: "exchange", Label: "换货", Color: "orange"},
					{Value: "repair", Label: "维修", Color: "blue"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{Name: "total_amount", Label: "退款金额", Type: "number", Precision: intPtr(2), Unit: "¥", Width: 100},
			{
				Name: "status", Label: "状态", Type: "select", Default: "requested",
				Options: []sdk.FieldOption{
					{Value: "requested", Label: "已申请", Color: "default"},
					{Value: "approved", Label: "已批准", Color: "processing"},
					{Value: "received", Label: "已收货", Color: "cyan"},
					{Value: "inspected", Label: "已检验", Color: "blue"},
					{Value: "completed", Label: "已完成", Color: "success"},
					{Value: "rejected", Label: "已拒绝", Color: "error"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"requested": "#d9d9d9",
						"approved":  "#1677ff",
						"received":  "#13c2c2",
						"inspected": "#0958d9",
						"completed": "#52c41a",
						"rejected":  "#f5222d",
					},
				},
				Width: 80,
			},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"code", "order_id", "customer_id", "type", "reason", "total_amount", "status", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "reason"},
		Filters:      []string{"customer_id", "type", "status"},
	}
}

// ---------------------------------------------------------------------------
// inventoryEntity — erp_inventory
// ---------------------------------------------------------------------------

func inventoryEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "inventory",
		Label:      "库存",
		Table:      "erp_inventory",
		PrimaryKey: "id",
		Icon:       "InboxOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "material_id", Label: "物料", Type: "relation", Required: true, RefEntity: "materials", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "warehouse_id", Label: "仓库", Type: "relation", Required: true, RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "location_id", Label: "库位", Type: "relation", RefEntity: "locations", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "lot_number", Label: "批次号", Type: "string", Width: 100},
			{Name: "quantity", Label: "现有数量", Type: "number", Precision: intPtr(4), Width: 100},
			{Name: "reserved_qty", Label: "预留数量", Type: "number", Precision: intPtr(4), Width: 100},
			{Name: "available_qty", Label: "可用数量", Type: "number", ReadOnly: true, Precision: intPtr(4), Width: 100, Component: "virtual"},
			{Name: "unit_cost", Label: "单位成本", Type: "number", Precision: intPtr(4), Unit: "¥", Width: 100},
			{
				Name: "status", Label: "状态", Type: "select", Default: "available",
				Options: []sdk.FieldOption{
					{Value: "available", Label: "可用", Color: "green"},
					{Value: "quarantine", Label: "隔离", Color: "orange"},
					{Value: "damaged", Label: "受损", Color: "red"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
			{Name: "expiry_date", Label: "有效期", Type: "date", Width: 110},
			{Name: "updated_at", Label: "更新时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"material_id", "warehouse_id", "location_id", "lot_number", "quantity", "reserved_qty", "available_qty", "unit_cost", "status"},
		DefaultSort:  "updated_at",
		DefaultOrder: "desc",
		Searchable:   []string{"lot_number"},
		Filters:      []string{"material_id", "warehouse_id", "status"},
	}
}

// ---------------------------------------------------------------------------
// inventoryTransactionEntity — erp_inventory_transactions
// ---------------------------------------------------------------------------

func inventoryTransactionEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "inventory_transactions",
		Label:      "库存事务",
		Table:      "erp_inventory_transactions",
		PrimaryKey: "id",
		Icon:       "SwapOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "事务编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{
				Name: "type", Label: "类型", Type: "select", Required: true,
				Options: []sdk.FieldOption{
					{Value: "receive", Label: "入库", Color: "green"},
					{Value: "issue", Label: "出库", Color: "red"},
					{Value: "transfer", Label: "调拨", Color: "blue"},
					{Value: "adjust", Label: "调整", Color: "orange"},
					{Value: "scrap", Label: "报废", Color: "default"},
					{Value: "return", Label: "退库", Color: "purple"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{Name: "material_id", Label: "物料", Type: "relation", Required: true, RefEntity: "materials", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "from_warehouse_id", Label: "源仓库", Type: "relation", RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "from_location_id", Label: "源库位", Type: "relation", RefEntity: "locations", RefModule: "erp", RefDisplay: "name", HideInList: true},
			{Name: "to_warehouse_id", Label: "目标仓库", Type: "relation", RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "to_location_id", Label: "目标库位", Type: "relation", RefEntity: "locations", RefModule: "erp", RefDisplay: "name", HideInList: true},
			{Name: "quantity", Label: "数量", Type: "number", Required: true, Precision: intPtr(4), Width: 90},
			{Name: "unit_cost", Label: "单位成本", Type: "number", Precision: intPtr(4), Unit: "¥", Width: 100},
			{Name: "lot_number", Label: "批次号", Type: "string", Width: 100},
			{
				Name: "reference_type", Label: "来源类型", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "po", Label: "采购订单", Color: "blue"},
					{Value: "so", Label: "销售订单", Color: "green"},
					{Value: "wo", Label: "生产工单", Color: "purple"},
					{Value: "adjust", Label: "手工调整", Color: "orange"},
					{Value: "scrap", Label: "报废", Color: "red"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "reference_id", Label: "关联单据", Type: "string", HideInList: true},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
			{Name: "created_by", Label: "操作人", Type: "string", ReadOnly: true, Width: 90, HideInForm: true},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"code", "type", "material_id", "from_warehouse_id", "to_warehouse_id", "quantity", "lot_number", "reference_type", "created_by", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "lot_number"},
		Filters:      []string{"type", "material_id", "reference_type"},
	}
}

// ---------------------------------------------------------------------------
// serialNumberEntity — erp_serial_numbers
// ---------------------------------------------------------------------------

func serialNumberEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "serial_numbers",
		Label:      "序列号",
		Table:      "erp_serial_numbers",
		PrimaryKey: "id",
		Icon:       "BarcodeOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "serial_number", Label: "序列号", Type: "string", Required: true, Unique: true, Width: 140},
			{Name: "material_id", Label: "物料", Type: "relation", RefEntity: "materials", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "product_id", Label: "产品", Type: "relation", RefEntity: "products", RefModule: "plm", RefDisplay: "name", Width: 120},
			{
				Name: "status", Label: "状态", Type: "select", Default: "in_stock",
				Options: []sdk.FieldOption{
					{Value: "in_stock", Label: "在库", Color: "green"},
					{Value: "sold", Label: "已售", Color: "blue"},
					{Value: "returned", Label: "已退", Color: "orange"},
					{Value: "scrapped", Label: "报废", Color: "red"},
					{Value: "in_repair", Label: "维修中", Color: "purple"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{Name: "warehouse_id", Label: "当前仓库", Type: "relation", RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "lot_number", Label: "所属批次", Type: "string", Width: 100},
			{Name: "manufactured_at", Label: "生产日期", Type: "date", Width: 110},
			{Name: "sold_to", Label: "销售客户", Type: "relation", RefEntity: "customers", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "sold_at", Label: "销售日期", Type: "date", Width: 110},
			{Name: "warranty_until", Label: "保修截止", Type: "date", Width: 110},
		},
		ListColumns:  []string{"serial_number", "product_id", "material_id", "status", "warehouse_id", "lot_number", "manufactured_at", "sold_to", "warranty_until"},
		DefaultSort:  "manufactured_at",
		DefaultOrder: "desc",
		Searchable:   []string{"serial_number", "lot_number"},
		Filters:      []string{"status", "product_id", "warehouse_id"},
	}
}

// ---------------------------------------------------------------------------
// mrpResultEntity — erp_mrp_results
// ---------------------------------------------------------------------------

func mrpResultEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "mrp_results",
		Label:      "MRP结果",
		Table:      "erp_mrp_results",
		PrimaryKey: "id",
		Icon:       "CalculatorOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "run_id", Label: "运算批次", Type: "string", ReadOnly: true, Width: 120},
			{Name: "material_id", Label: "物料", Type: "relation", Required: true, RefEntity: "materials", RefModule: "plm", RefDisplay: "name", Width: 120},
			{
				Name: "demand_source", Label: "需求来源", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "so", Label: "销售订单", Color: "green"},
					{Value: "wo", Label: "生产工单", Color: "blue"},
					{Value: "forecast", Label: "预测", Color: "orange"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "demand_id", Label: "需求单据", Type: "string", HideInList: true},
			{Name: "gross_requirement", Label: "毛需求", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "on_hand", Label: "现有库存", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "on_order", Label: "在途数量", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "net_requirement", Label: "净需求", Type: "number", ReadOnly: true, Precision: intPtr(4), Width: 90},
			{
				Name: "action", Label: "建议动作", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "purchase", Label: "采购", Color: "blue"},
					{Value: "produce", Label: "生产", Color: "green"},
					{Value: "none", Label: "无需处理", Color: "default"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "suggested_qty", Label: "建议数量", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "suggested_date", Label: "建议日期", Type: "date", Width: 110},
			{Name: "bom_id", Label: "展开BOM", Type: "relation", RefEntity: "boms", RefModule: "plm", RefDisplay: "name", HideInList: true},
			{Name: "bom_level", Label: "BOM层级", Type: "integer", Width: 70},
			{
				Name: "status", Label: "状态", Type: "select", Default: "suggested",
				Options: []sdk.FieldOption{
					{Value: "suggested", Label: "建议", Color: "default"},
					{Value: "confirmed", Label: "已确认", Color: "processing"},
					{Value: "executed", Label: "已执行", Color: "success"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"run_id", "material_id", "demand_source", "gross_requirement", "on_hand", "net_requirement", "action", "suggested_qty", "suggested_date", "status"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"run_id"},
		Filters:      []string{"demand_source", "action", "status"},
	}
}

// ---------------------------------------------------------------------------
// workOrderEntity — erp_work_orders
// ---------------------------------------------------------------------------

func workOrderEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "work_orders",
		Label:      "生产工单",
		Table:      "erp_work_orders",
		PrimaryKey: "id",
		Icon:       "ToolOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "工单编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "product_id", Label: "产品", Type: "relation", Required: true, RefEntity: "products", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "bom_id", Label: "BOM", Type: "relation", Required: true, RefEntity: "boms", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "order_id", Label: "销售订单", Type: "relation", RefEntity: "sales_orders", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "mrp_result_id", Label: "MRP建议", Type: "relation", RefEntity: "mrp_results", RefModule: "erp", RefDisplay: "id", HideInList: true},
			{Name: "planned_qty", Label: "计划数量", Type: "number", Required: true, Precision: intPtr(4), Width: 90},
			{Name: "completed_qty", Label: "完工数量", Type: "number", ReadOnly: true, Precision: intPtr(4), Width: 90},
			{Name: "scrap_qty", Label: "报废数量", Type: "number", ReadOnly: true, Precision: intPtr(4), Width: 90},
			{Name: "warehouse_id", Label: "入库仓", Type: "relation", RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "planned_start", Label: "计划开始", Type: "date", Width: 110},
			{Name: "planned_end", Label: "计划完成", Type: "date", Width: 110},
			{Name: "actual_start", Label: "实际开始", Type: "datetime", ReadOnly: true, HideInList: true},
			{Name: "actual_end", Label: "实际完成", Type: "datetime", ReadOnly: true, HideInList: true},
			{
				Name: "priority", Label: "优先级", Type: "select", Default: "normal",
				Options: []sdk.FieldOption{
					{Value: "low", Label: "低", Color: "default"},
					{Value: "normal", Label: "普通", Color: "blue"},
					{Value: "high", Label: "高", Color: "orange"},
					{Value: "urgent", Label: "紧急", Color: "red"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{
				Name: "status", Label: "状态", Type: "select", Default: "draft",
				Options: []sdk.FieldOption{
					{Value: "draft", Label: "草稿", Color: "default"},
					{Value: "released", Label: "已下达", Color: "processing"},
					{Value: "in_progress", Label: "生产中", Color: "blue"},
					{Value: "completed", Label: "已完工", Color: "success"},
					{Value: "closed", Label: "已关闭", Color: "default"},
					{Value: "cancelled", Label: "已取消", Color: "error"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"draft":       "#d9d9d9",
						"released":    "#1677ff",
						"in_progress": "#0958d9",
						"completed":   "#52c41a",
						"closed":      "#8c8c8c",
						"cancelled":   "#f5222d",
					},
				},
				Width: 80,
			},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
			{Name: "created_by", Label: "创建人", Type: "string", ReadOnly: true, Width: 90, HideInForm: true},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"code", "product_id", "bom_id", "planned_qty", "completed_qty", "planned_start", "planned_end", "priority", "status", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code"},
		Filters:      []string{"product_id", "status", "priority"},
		Relations: []sdk.RelationDef{
			{Name: "material_issues", Label: "物料领用", Type: "has_many", Target: "wo_material_issues", ForeignKey: "work_order_id", Display: "table"},
			{Name: "reports", Label: "报工记录", Type: "has_many", Target: "wo_reports", ForeignKey: "work_order_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// woMaterialIssueEntity — erp_wo_material_issues
// ---------------------------------------------------------------------------

func woMaterialIssueEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "wo_material_issues",
		Label:      "工单领料",
		Table:      "erp_wo_material_issues",
		PrimaryKey: "id",
		Icon:       "ExportOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "work_order_id", Label: "生产工单", Type: "relation", Required: true, RefEntity: "work_orders", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "material_id", Label: "物料", Type: "relation", Required: true, RefEntity: "materials", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "bom_item_id", Label: "BOM行", Type: "relation", RefEntity: "bom_items", RefModule: "plm", RefDisplay: "name", HideInList: true},
			{Name: "required_qty", Label: "需求数量", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "issued_qty", Label: "领用数量", Type: "number", Required: true, Precision: intPtr(4), Width: 90},
			{Name: "warehouse_id", Label: "领料仓库", Type: "relation", RefEntity: "warehouses", RefModule: "erp", RefDisplay: "name", Width: 100},
			{Name: "lot_number", Label: "批次", Type: "string", Width: 100},
			{Name: "issued_at", Label: "领料时间", Type: "datetime", Width: 160},
			{Name: "issued_by", Label: "领料人", Type: "string", Width: 90},
		},
		ListColumns:  []string{"material_id", "required_qty", "issued_qty", "warehouse_id", "lot_number", "issued_at", "issued_by"},
		DefaultSort:  "issued_at",
		DefaultOrder: "desc",
		Searchable:   []string{"lot_number"},
		Filters:      []string{"work_order_id", "material_id"},
	}
}

// ---------------------------------------------------------------------------
// woReportEntity — erp_wo_reports
// ---------------------------------------------------------------------------

func woReportEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "wo_reports",
		Label:      "工单报工",
		Table:      "erp_wo_reports",
		PrimaryKey: "id",
		Icon:       "FormOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "work_order_id", Label: "生产工单", Type: "relation", Required: true, RefEntity: "work_orders", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "operation", Label: "工序名称", Type: "string", Required: true, Width: 120},
			{Name: "operator_id", Label: "操作员", Type: "string", Width: 90},
			{Name: "good_qty", Label: "良品数量", Type: "number", Required: true, Precision: intPtr(4), Width: 90},
			{Name: "defect_qty", Label: "不良数量", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "scrap_qty", Label: "报废数量", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "start_time", Label: "开始时间", Type: "datetime", Width: 160},
			{Name: "end_time", Label: "结束时间", Type: "datetime", Width: 160},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
		},
		ListColumns:  []string{"operation", "operator_id", "good_qty", "defect_qty", "scrap_qty", "start_time", "end_time"},
		DefaultSort:  "start_time",
		DefaultOrder: "desc",
		Searchable:   []string{"operation"},
		Filters:      []string{"work_order_id"},
	}
}

// ---------------------------------------------------------------------------
// accountEntity — erp_accounts
// ---------------------------------------------------------------------------

func accountEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "accounts",
		Label:      "会计科目",
		Table:      "erp_accounts",
		PrimaryKey: "id",
		Icon:       "AccountBookOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "科目编码", Type: "string", ReadOnly: true, Unique: true, Width: 100},
			{Name: "name", Label: "科目名称", Type: "string", Required: true},
			{
				Name: "type", Label: "类型", Type: "select", Required: true,
				Options: []sdk.FieldOption{
					{Value: "asset", Label: "资产", Color: "blue"},
					{Value: "liability", Label: "负债", Color: "orange"},
					{Value: "equity", Label: "权益", Color: "purple"},
					{Value: "revenue", Label: "收入", Color: "green"},
					{Value: "expense", Label: "费用", Color: "red"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  80,
			},
			{Name: "parent_id", Label: "父科目", Type: "relation", RefEntity: "accounts", RefModule: "erp", RefDisplay: "name"},
			{Name: "level", Label: "层级", Type: "integer", Width: 60},
			{Name: "is_leaf", Label: "末级科目", Type: "boolean"},
			{
				Name: "currency", Label: "核算币种", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "CNY", Label: "CNY", Color: "red"},
					{Value: "USD", Label: "USD", Color: "green"},
				},
				Width: 70,
			},
			{
				Name: "status", Label: "状态", Type: "select", Default: "active",
				Options: []sdk.FieldOption{
					{Value: "active", Label: "启用", Color: "green"},
					{Value: "inactive", Label: "停用", Color: "default"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
		},
		ListColumns:  []string{"code", "name", "type", "parent_id", "level", "is_leaf", "currency", "status"},
		DefaultSort:  "code",
		DefaultOrder: "asc",
		Searchable:   []string{"code", "name"},
		Filters:      []string{"type", "status"},
	}
}

// ---------------------------------------------------------------------------
// journalEntryEntity — erp_journal_entries
// ---------------------------------------------------------------------------

func journalEntryEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "journal_entries",
		Label:      "会计凭证",
		Table:      "erp_journal_entries",
		PrimaryKey: "id",
		Icon:       "AuditOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "凭证编号", Type: "string", ReadOnly: true, Unique: true, Width: 130},
			{Name: "period", Label: "会计期间", Type: "string", Width: 90},
			{Name: "entry_date", Label: "记账日期", Type: "date", Required: true, Width: 110},
			{
				Name: "source_type", Label: "来源类型", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "sales_invoice", Label: "销售发票", Color: "green"},
					{Value: "purchase_invoice", Label: "采购发票", Color: "blue"},
					{Value: "receipt", Label: "收款", Color: "cyan"},
					{Value: "payment", Label: "付款", Color: "orange"},
					{Value: "manual", Label: "手工", Color: "default"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "source_id", Label: "来源单据", Type: "string", HideInList: true},
			{Name: "description", Label: "摘要", Type: "text", Width: 200},
			{Name: "total_debit", Label: "借方合计", Type: "number", ReadOnly: true, Precision: intPtr(2), Unit: "¥", Width: 110},
			{Name: "total_credit", Label: "贷方合计", Type: "number", ReadOnly: true, Precision: intPtr(2), Unit: "¥", Width: 110},
			{
				Name: "status", Label: "状态", Type: "select", Default: "draft",
				Options: []sdk.FieldOption{
					{Value: "draft", Label: "草稿", Color: "default"},
					{Value: "posted", Label: "已过账", Color: "success"},
					{Value: "reversed", Label: "已冲销", Color: "error"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
			{Name: "posted_by", Label: "过账人", Type: "string", ReadOnly: true, HideInList: true},
			{Name: "posted_at", Label: "过账时间", Type: "datetime", ReadOnly: true, HideInList: true},
			{Name: "created_by", Label: "创建人", Type: "string", ReadOnly: true, Width: 90, HideInForm: true},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"code", "period", "entry_date", "source_type", "description", "total_debit", "total_credit", "status", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "description"},
		Filters:      []string{"source_type", "status", "period"},
		Relations: []sdk.RelationDef{
			{Name: "lines", Label: "分录明细", Type: "has_many", Target: "journal_lines", ForeignKey: "entry_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// journalLineEntity — erp_journal_lines
// ---------------------------------------------------------------------------

func journalLineEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "journal_lines",
		Label:      "凭证分录",
		Table:      "erp_journal_lines",
		PrimaryKey: "id",
		Icon:       "UnorderedListOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "entry_id", Label: "凭证", Type: "relation", Required: true, RefEntity: "journal_entries", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "account_id", Label: "科目", Type: "relation", Required: true, RefEntity: "accounts", RefModule: "erp", RefDisplay: "name", Width: 120},
			{Name: "debit", Label: "借方金额", Type: "number", Precision: intPtr(2), Unit: "¥", Width: 110},
			{Name: "credit", Label: "贷方金额", Type: "number", Precision: intPtr(2), Unit: "¥", Width: 110},
			{
				Name: "currency", Label: "原币币种", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "CNY", Label: "CNY", Color: "red"},
					{Value: "USD", Label: "USD", Color: "green"},
				},
				Width: 70,
			},
			{Name: "original_amount", Label: "原币金额", Type: "number", Precision: intPtr(2), Unit: "¥", Width: 100},
			{Name: "description", Label: "行摘要", Type: "string", Width: 200},
			{Name: "customer_id", Label: "辅助-客户", Type: "relation", RefEntity: "customers", RefModule: "erp", RefDisplay: "name", HideInList: true},
			{Name: "supplier_id", Label: "辅助-供应商", Type: "string", HideInList: true},
			{Name: "department_id", Label: "辅助-部门", Type: "string", HideInList: true},
		},
		ListColumns:  []string{"account_id", "description", "debit", "credit", "currency", "original_amount"},
		DefaultSort:  "id",
		DefaultOrder: "asc",
		Searchable:   []string{"description"},
		Filters:      []string{"entry_id", "account_id"},
	}
}

// ---------------------------------------------------------------------------
// salesInvoiceEntity — erp_sales_invoices
// ---------------------------------------------------------------------------

func salesInvoiceEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "sales_invoices",
		Label:      "销售发票",
		Table:      "erp_sales_invoices",
		PrimaryKey: "id",
		Icon:       "DollarOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "发票编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "order_id", Label: "销售订单", Type: "relation", Required: true, RefEntity: "sales_orders", RefModule: "erp", RefDisplay: "code", Width: 140},
			{Name: "customer_id", Label: "客户", Type: "relation", Required: true, RefEntity: "customers", RefModule: "erp", RefDisplay: "name", Width: 120},
			{Name: "invoice_date", Label: "开票日期", Type: "date", Required: true, Width: 110},
			{Name: "due_date", Label: "到期日", Type: "date", Width: 110},
			{
				Name: "currency", Label: "币种", Type: "select", Default: "CNY",
				Options: []sdk.FieldOption{
					{Value: "CNY", Label: "CNY", Color: "red"},
					{Value: "USD", Label: "USD", Color: "green"},
				},
				Width: 70,
			},
			{Name: "subtotal", Label: "小计", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, HideInList: true},
			{Name: "tax_amount", Label: "税额", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, HideInList: true},
			{Name: "total", Label: "总计", Type: "number", Precision: intPtr(2), Unit: "¥", ReadOnly: true, Width: 110},
			{Name: "paid_amount", Label: "已收金额", Type: "number", ReadOnly: true, Precision: intPtr(2), Unit: "¥", Width: 100},
			{Name: "balance", Label: "未收余额", Type: "number", ReadOnly: true, Precision: intPtr(2), Unit: "¥", Width: 100, Component: "virtual"},
			{
				Name: "status", Label: "状态", Type: "select", Default: "draft",
				Options: []sdk.FieldOption{
					{Value: "draft", Label: "草稿", Color: "default"},
					{Value: "issued", Label: "已开票", Color: "processing"},
					{Value: "partially_paid", Label: "部分收款", Color: "cyan"},
					{Value: "paid", Label: "已收齐", Color: "success"},
					{Value: "overdue", Label: "逾期", Color: "error"},
					{Value: "cancelled", Label: "已作废", Color: "default"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"draft":          "#d9d9d9",
						"issued":         "#1677ff",
						"partially_paid": "#13c2c2",
						"paid":           "#52c41a",
						"overdue":        "#f5222d",
						"cancelled":      "#8c8c8c",
					},
				},
				Width: 90,
			},
			{Name: "journal_entry_id", Label: "关联凭证", Type: "relation", RefEntity: "journal_entries", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
		},
		ListColumns:  []string{"code", "order_id", "customer_id", "invoice_date", "due_date", "total", "paid_amount", "balance", "status"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code"},
		Filters:      []string{"customer_id", "status"},
		Relations: []sdk.RelationDef{
			{Name: "allocations", Label: "核销记录", Type: "has_many", Target: "receipt_allocations", ForeignKey: "invoice_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// receiptEntity — erp_receipts
// ---------------------------------------------------------------------------

func receiptEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "receipts",
		Label:      "收款记录",
		Table:      "erp_receipts",
		PrimaryKey: "id",
		Icon:       "BankOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "收款编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "customer_id", Label: "客户", Type: "relation", Required: true, RefEntity: "customers", RefModule: "erp", RefDisplay: "name", Width: 120},
			{Name: "amount", Label: "收款金额", Type: "number", Required: true, Precision: intPtr(2), Unit: "¥", Width: 110},
			{
				Name: "currency", Label: "币种", Type: "select", Default: "CNY",
				Options: []sdk.FieldOption{
					{Value: "CNY", Label: "CNY", Color: "red"},
					{Value: "USD", Label: "USD", Color: "green"},
				},
				Width: 70,
			},
			{
				Name: "payment_method", Label: "收款方式", Type: "select", Required: true,
				Options: []sdk.FieldOption{
					{Value: "bank_transfer", Label: "银行转账", Color: "blue"},
					{Value: "check", Label: "支票", Color: "cyan"},
					{Value: "cash", Label: "现金", Color: "green"},
					{Value: "online", Label: "在线支付", Color: "purple"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "bank_account", Label: "收款银行账号", Type: "string", HideInList: true},
			{Name: "reference_no", Label: "银行流水号", Type: "string", Width: 140},
			{Name: "received_date", Label: "收款日期", Type: "date", Required: true, Width: 110},
			{
				Name: "status", Label: "状态", Type: "select", Default: "draft",
				Options: []sdk.FieldOption{
					{Value: "draft", Label: "草稿", Color: "default"},
					{Value: "confirmed", Label: "已确认", Color: "processing"},
					{Value: "reconciled", Label: "已核销", Color: "success"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
			{Name: "journal_entry_id", Label: "关联凭证", Type: "relation", RefEntity: "journal_entries", RefModule: "erp", RefDisplay: "code", HideInList: true},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
		},
		ListColumns:  []string{"code", "customer_id", "amount", "payment_method", "reference_no", "received_date", "status"},
		DefaultSort:  "received_date",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "reference_no"},
		Filters:      []string{"customer_id", "payment_method", "status"},
		Relations: []sdk.RelationDef{
			{Name: "allocations", Label: "核销明细", Type: "has_many", Target: "receipt_allocations", ForeignKey: "receipt_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// receiptAllocationEntity — erp_receipt_allocations
// ---------------------------------------------------------------------------

func receiptAllocationEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "receipt_allocations",
		Label:      "收款核销",
		Table:      "erp_receipt_allocations",
		PrimaryKey: "id",
		Icon:       "LinkOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "receipt_id", Label: "收款", Type: "relation", Required: true, RefEntity: "receipts", RefModule: "erp", RefDisplay: "code", Width: 140},
			{Name: "invoice_id", Label: "发票", Type: "relation", Required: true, RefEntity: "sales_invoices", RefModule: "erp", RefDisplay: "code", Width: 140},
			{Name: "amount", Label: "核销金额", Type: "number", Required: true, Precision: intPtr(2), Unit: "¥", Width: 110},
		},
		ListColumns:  []string{"receipt_id", "invoice_id", "amount"},
		DefaultSort:  "id",
		DefaultOrder: "asc",
		Searchable:   []string{},
		Filters:      []string{"receipt_id", "invoice_id"},
	}
}

// ---------------------------------------------------------------------------
// oqcInspectionEntity — erp_oqc_inspections
// ---------------------------------------------------------------------------

func oqcInspectionEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "oqc_inspections",
		Label:      "出货检验",
		Table:      "erp_oqc_inspections",
		PrimaryKey: "id",
		Icon:       "SafetyCertificateOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "检验编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{Name: "shipment_id", Label: "发货单", Type: "relation", Required: true, RefEntity: "shipments", RefModule: "erp", RefDisplay: "code", Width: 140},
			{Name: "product_id", Label: "产品", Type: "relation", RefEntity: "products", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "lot_number", Label: "批次", Type: "string", Width: 100},
			{Name: "sample_size", Label: "抽样数", Type: "integer", Width: 70},
			{Name: "total_inspected", Label: "检验总数", Type: "integer", Width: 80},
			{Name: "pass_count", Label: "合格数", Type: "integer", Width: 70},
			{Name: "fail_count", Label: "不合格数", Type: "integer", Width: 80},
			{
				Name: "result", Label: "检验结果", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "pass", Label: "合格", Color: "green"},
					{Value: "fail", Label: "不合格", Color: "red"},
					{Value: "conditional", Label: "有条件放行", Color: "orange"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "inspector_id", Label: "检验员", Type: "string", Width: 90},
			{Name: "inspected_at", Label: "检验时间", Type: "datetime", Width: 160},
			{Name: "notes", Label: "备注", Type: "text", HideInList: true},
			{Name: "defect_details", Label: "不良明细", Type: "json", HideInList: true},
			{
				Name: "status", Label: "状态", Type: "select", Default: "pending",
				Options: []sdk.FieldOption{
					{Value: "pending", Label: "待检", Color: "default"},
					{Value: "completed", Label: "已完成", Color: "success"},
				},
				Render: &sdk.FieldRender{Type: "badge"},
				Width:  80,
			},
		},
		ListColumns:  []string{"code", "shipment_id", "product_id", "lot_number", "sample_size", "pass_count", "fail_count", "result", "inspector_id", "status"},
		DefaultSort:  "inspected_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "lot_number"},
		Filters:      []string{"result", "status"},
	}
}

// ---------------------------------------------------------------------------
// ncrReportEntity — erp_ncr_reports
// ---------------------------------------------------------------------------

func ncrReportEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "ncr_reports",
		Label:      "不合格品报告",
		Table:      "erp_ncr_reports",
		PrimaryKey: "id",
		Icon:       "ExclamationCircleOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "NCR编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{
				Name: "source", Label: "来源", Type: "select", Required: true,
				Options: []sdk.FieldOption{
					{Value: "iqc", Label: "来料检验", Color: "blue"},
					{Value: "ipqc", Label: "过程检验", Color: "cyan"},
					{Value: "oqc", Label: "出货检验", Color: "green"},
					{Value: "customer_return", Label: "客户退货", Color: "orange"},
					{Value: "internal", Label: "内部发现", Color: "purple"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "source_id", Label: "来源单据", Type: "string", HideInList: true},
			{Name: "product_id", Label: "产品", Type: "relation", RefEntity: "products", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "material_id", Label: "物料", Type: "relation", RefEntity: "materials", RefModule: "plm", RefDisplay: "name", Width: 120},
			{Name: "lot_number", Label: "批次号", Type: "string", Width: 100},
			{Name: "defect_qty", Label: "不良数量", Type: "number", Precision: intPtr(4), Width: 90},
			{Name: "defect_type", Label: "不良类型", Type: "string", Width: 100},
			{Name: "description", Label: "不良描述", Type: "text", HideInList: true},
			{
				Name: "disposition", Label: "处置方式", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "use_as_is", Label: "让步接收", Color: "blue"},
					{Value: "rework", Label: "返工", Color: "orange"},
					{Value: "scrap", Label: "报废", Color: "red"},
					{Value: "return_to_supplier", Label: "退供应商", Color: "purple"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{
				Name: "severity", Label: "严重程度", Type: "select",
				Options: []sdk.FieldOption{
					{Value: "minor", Label: "轻微", Color: "default"},
					{Value: "major", Label: "严重", Color: "orange"},
					{Value: "critical", Label: "致命", Color: "red"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"minor":    "#d9d9d9",
						"major":    "#fa8c16",
						"critical": "#f5222d",
					},
				},
				Width: 80,
			},
			{
				Name: "status", Label: "状态", Type: "select", Default: "open",
				Options: []sdk.FieldOption{
					{Value: "open", Label: "待处理", Color: "default"},
					{Value: "reviewing", Label: "评审中", Color: "processing"},
					{Value: "dispositioned", Label: "已处置", Color: "cyan"},
					{Value: "closed", Label: "已关闭", Color: "success"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"open":          "#d9d9d9",
						"reviewing":     "#1677ff",
						"dispositioned": "#13c2c2",
						"closed":        "#52c41a",
					},
				},
				Width: 80,
			},
			{Name: "owner_id", Label: "责任人", Type: "string", Width: 90},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
			{Name: "closed_at", Label: "关闭时间", Type: "datetime", ReadOnly: true, HideInList: true},
		},
		ListColumns:  []string{"code", "source", "product_id", "material_id", "defect_type", "defect_qty", "severity", "disposition", "status", "owner_id"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "defect_type", "lot_number"},
		Filters:      []string{"source", "severity", "disposition", "status"},
		Relations: []sdk.RelationDef{
			{Name: "capas", Label: "CAPA", Type: "has_many", Target: "capa", ForeignKey: "ncr_id", Display: "table"},
		},
	}
}

// ---------------------------------------------------------------------------
// capaEntity — erp_capa
// ---------------------------------------------------------------------------

func capaEntity() sdk.EntityDef {
	return sdk.EntityDef{
		Name:       "capa",
		Label:      "CAPA",
		Table:      "erp_capa",
		PrimaryKey: "id",
		Icon:       "MedicineBoxOutlined",
		Fields: []sdk.FieldDef{
			{Name: "id", Label: "ID", Type: "string", ReadOnly: true, HideInList: true, HideInForm: true},
			{Name: "code", Label: "CAPA编号", Type: "string", ReadOnly: true, Unique: true, Width: 140},
			{
				Name: "type", Label: "类型", Type: "select", Required: true,
				Options: []sdk.FieldOption{
					{Value: "corrective", Label: "纠正措施", Color: "orange"},
					{Value: "preventive", Label: "预防措施", Color: "blue"},
				},
				Render: &sdk.FieldRender{Type: "tag"},
				Width:  90,
			},
			{Name: "ncr_id", Label: "关联NCR", Type: "relation", RefEntity: "ncr_reports", RefModule: "erp", RefDisplay: "code", Width: 140},
			{Name: "title", Label: "标题", Type: "string", Required: true},
			{Name: "root_cause", Label: "根因分析", Type: "text", HideInList: true},
			{Name: "action_plan", Label: "措施计划", Type: "text", HideInList: true},
			{Name: "owner_id", Label: "责任人", Type: "string", Width: 90},
			{Name: "due_date", Label: "截止日期", Type: "date", Width: 110},
			{Name: "verification", Label: "有效性验证", Type: "text", HideInList: true},
			{
				Name: "status", Label: "状态", Type: "select", Default: "open",
				Options: []sdk.FieldOption{
					{Value: "open", Label: "待处理", Color: "default"},
					{Value: "in_progress", Label: "进行中", Color: "processing"},
					{Value: "pending_verification", Label: "待验证", Color: "cyan"},
					{Value: "closed", Label: "已关闭", Color: "success"},
				},
				Render: &sdk.FieldRender{
					Type: "tag",
					ColorMap: map[string]string{
						"open":                 "#d9d9d9",
						"in_progress":          "#1677ff",
						"pending_verification": "#13c2c2",
						"closed":               "#52c41a",
					},
				},
				Width: 80,
			},
			{Name: "created_at", Label: "创建时间", Type: "datetime", ReadOnly: true, Width: 160},
			{Name: "closed_at", Label: "关闭时间", Type: "datetime", ReadOnly: true, HideInList: true},
		},
		ListColumns:  []string{"code", "type", "title", "ncr_id", "owner_id", "due_date", "status", "created_at"},
		DefaultSort:  "created_at",
		DefaultOrder: "desc",
		Searchable:   []string{"code", "title"},
		Filters:      []string{"type", "status"},
	}
}
