package entity

import "time"

// ProductSKU 产品SKU/配色方案
type ProductSKU struct {
	ID          string    `json:"id" gorm:"primaryKey;size:32"`
	ProjectID   string    `json:"project_id" gorm:"size:32;not null;index"`
	Name        string    `json:"name" gorm:"size:128;not null"`
	Code        string    `json:"code" gorm:"size:32"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status" gorm:"size:16;not null;default:active"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedBy   string    `json:"created_by" gorm:"size:32;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Project    *Project       `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	CMFConfigs []SKUCMFConfig `json:"cmf_configs,omitempty" gorm:"foreignKey:SKUID"`
	BOMItems   []SKUBOMItem   `json:"bom_items,omitempty" gorm:"foreignKey:SKUID"`
}

func (ProductSKU) TableName() string {
	return "product_skus"
}

// SKUCMFConfig SKU × SBOM Item → CMF配置（颜色/表面处理）
type SKUCMFConfig struct {
	ID               string    `json:"id" gorm:"primaryKey;size:32"`
	SKUID            string    `json:"sku_id" gorm:"column:sku_id;size:32;not null;index"`
	BOMItemID        string    `json:"bom_item_id" gorm:"column:bom_item_id;size:32;not null"`
	Color            string    `json:"color" gorm:"size:64"`
	ColorCode        string    `json:"color_code" gorm:"size:32"`
	SurfaceTreatment string    `json:"surface_treatment" gorm:"size:128"`
	ProcessParams    string    `json:"process_params,omitempty" gorm:"type:jsonb"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	SKU     *ProductSKU     `json:"sku,omitempty" gorm:"foreignKey:SKUID"`
	BOMItem *ProjectBOMItem `json:"bom_item,omitempty" gorm:"foreignKey:BOMItemID"`
}

func (SKUCMFConfig) TableName() string {
	return "sku_cmf_configs"
}

// SKUBOMItem SKU与SBOM零件的关联（从SBOM全量中勾选该SKU使用的零件）
type SKUBOMItem struct {
	ID           string  `json:"id" gorm:"primaryKey;size:32"`
	SKUID        string  `json:"sku_id" gorm:"column:sku_id;size:32;not null;index"`
	BOMItemID    string  `json:"bom_item_id" gorm:"column:bom_item_id;size:32;not null;index"`
	CMFVariantID *string `json:"cmf_variant_id,omitempty" gorm:"size:32"`
	Quantity     float64 `json:"quantity" gorm:"type:numeric(15,4);default:0"` // 0表示使用SBOM默认数量
	Notes        string  `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	SKU        *ProductSKU        `json:"sku,omitempty" gorm:"foreignKey:SKUID"`
	BOMItem    *ProjectBOMItem    `json:"bom_item,omitempty" gorm:"foreignKey:BOMItemID"`
	CMFVariant *BOMItemCMFVariant `json:"cmf_variant,omitempty" gorm:"foreignKey:CMFVariantID"`
}

func (SKUBOMItem) TableName() string {
	return "sku_bom_items"
}
