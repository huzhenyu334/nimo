package entity

import "time"

// InventoryRecord 库存记录
type InventoryRecord struct {
	ID           string   `json:"id" gorm:"primaryKey;size:32"`
	MaterialName string   `json:"material_name" gorm:"size:200;not null"`
	MaterialCode string   `json:"material_code" gorm:"size:50"`
	MPN          string   `json:"mpn" gorm:"size:100"`
	SupplierID   *string  `json:"supplier_id" gorm:"size:32"`
	Quantity     float64  `json:"quantity" gorm:"type:decimal(10,2);default:0"`
	Unit         string   `json:"unit" gorm:"size:20;default:pcs"`
	Warehouse    string   `json:"warehouse" gorm:"size:100"`
	LastInDate   *time.Time `json:"last_in_date"`
	SafetyStock  float64  `json:"safety_stock" gorm:"type:decimal(10,2);default:0"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联
	Supplier     *Supplier  `json:"supplier,omitempty" gorm:"foreignKey:SupplierID"`
}

func (InventoryRecord) TableName() string {
	return "srm_inventory_records"
}

// InventoryTransaction 库存流水
type InventoryTransaction struct {
	ID            string    `json:"id" gorm:"primaryKey;size:32"`
	InventoryID   string    `json:"inventory_id" gorm:"size:32;not null;index"`
	Type          string    `json:"type" gorm:"size:10;not null"` // in/out/adjust
	Quantity      float64   `json:"quantity" gorm:"type:decimal(10,2);not null"`
	ReferenceType string    `json:"reference_type" gorm:"size:20"` // inspection/manual/adjust
	ReferenceID   string    `json:"reference_id" gorm:"size:32"`
	Operator      string    `json:"operator" gorm:"size:100"`
	Notes         string    `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
}

func (InventoryTransaction) TableName() string {
	return "srm_inventory_transactions"
}

// 库存流水类型
const (
	InventoryTxTypeIn     = "in"
	InventoryTxTypeOut    = "out"
	InventoryTxTypeAdjust = "adjust"
)

// 来源类型
const (
	InventoryRefInspection = "inspection"
	InventoryRefManual     = "manual"
	InventoryRefAdjust     = "adjust"
)
