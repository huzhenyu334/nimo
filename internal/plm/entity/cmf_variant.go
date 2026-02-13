package entity

import "time"

// BOMItemCMFVariant CMF变体（外观件的颜色/材质/表面处理方案）
type BOMItemCMFVariant struct {
	ID                  string    `json:"id" gorm:"primaryKey;size:32"`
	BOMItemID           string    `json:"bom_item_id" gorm:"size:32;not null;index"`
	VariantIndex        int       `json:"variant_index" gorm:"not null;default:1"`
	MaterialCode        string    `json:"material_code,omitempty" gorm:"size:50"`
	ColorName           string    `json:"color_name,omitempty" gorm:"size:100"`
	ColorHex            string    `json:"color_hex,omitempty" gorm:"size:7"`
	Material            string    `json:"material,omitempty" gorm:"size:200"`
	Finish              string    `json:"finish,omitempty" gorm:"size:200"`
	Texture             string    `json:"texture,omitempty" gorm:"size:200"`
	Coating             string    `json:"coating,omitempty" gorm:"size:200"`
	PantoneCode         string    `json:"pantone_code,omitempty" gorm:"size:50"`
	GlossLevel          string    `json:"gloss_level,omitempty" gorm:"size:32"`
	ReferenceImageFileID string   `json:"reference_image_file_id,omitempty" gorm:"size:100"`
	ReferenceImageURL   string    `json:"reference_image_url,omitempty" gorm:"size:500"`
	ProcessDrawingType  string    `json:"process_drawing_type,omitempty" gorm:"size:50"`
	ProcessDrawings     string    `json:"process_drawings,omitempty" gorm:"type:jsonb;default:'[]'"`
	Notes               string    `json:"notes,omitempty" gorm:"type:text"`
	Status              string    `json:"status" gorm:"size:20;not null;default:draft"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// Relations
	BOMItem *ProjectBOMItem `json:"bom_item,omitempty" gorm:"foreignKey:BOMItemID"`
}

func (BOMItemCMFVariant) TableName() string {
	return "bom_item_cmf_variants"
}
