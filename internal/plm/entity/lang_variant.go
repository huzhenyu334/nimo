package entity

import "time"

// BOMItemLangVariant 包装件语言变体（多语言说明书/保修卡/彩盒等）
type BOMItemLangVariant struct {
	ID             string    `json:"id" gorm:"primaryKey;size:32"`
	BOMItemID      string    `json:"bom_item_id" gorm:"size:32;not null;index"`
	VariantIndex   int       `json:"variant_index" gorm:"not null;default:1"`
	MaterialCode   string    `json:"material_code,omitempty" gorm:"size:50"`
	LanguageCode   string    `json:"language_code" gorm:"size:10;not null"`
	LanguageName   string    `json:"language_name" gorm:"size:50;not null"`
	DesignFileID   string    `json:"design_file_id,omitempty" gorm:"size:100"`
	DesignFileName string    `json:"design_file_name,omitempty" gorm:"size:200"`
	DesignFileURL  string    `json:"design_file_url,omitempty" gorm:"size:500"`
	Notes          string    `json:"notes,omitempty" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	BOMItem *ProjectBOMItem `json:"bom_item,omitempty" gorm:"foreignKey:BOMItemID"`
}

func (BOMItemLangVariant) TableName() string {
	return "bom_item_lang_variants"
}
