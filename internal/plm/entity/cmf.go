package entity

import "time"

// CMFDesign 一个外观零件的一种CMF方案
type CMFDesign struct {
	ID                  string    `json:"id" gorm:"primaryKey;size:32"`
	ProjectID           string    `json:"project_id" gorm:"size:32;not null;index"`
	TaskID              string    `json:"task_id" gorm:"size:32;not null;index"`
	BOMItemID           string    `json:"bom_item_id" gorm:"size:32;not null;index"`
	SchemeName          string    `json:"scheme_name" gorm:"size:64"`
	Color               string    `json:"color" gorm:"size:64"`
	ColorCode           string    `json:"color_code" gorm:"size:64"`
	GlossLevel          string    `json:"gloss_level" gorm:"size:32"`  // 高光/半哑/哑光/丝光/镜面
	SurfaceTreatment    string    `json:"surface_treatment" gorm:"size:128"`
	TexturePattern      string    `json:"texture_pattern" gorm:"size:64"` // 皮纹/磨砂/拉丝
	CoatingType         string    `json:"coating_type" gorm:"size:64"`    // UV漆/PU漆/粉末涂装
	RenderImageFileID   *string   `json:"render_image_file_id" gorm:"size:32"`
	RenderImageFileName string    `json:"render_image_file_name" gorm:"size:256"`
	Notes               string    `json:"notes" gorm:"type:text"`
	SortOrder           int       `json:"sort_order" gorm:"default:0"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	BOMItem  *ProjectBOMItem `json:"bom_item,omitempty" gorm:"foreignKey:BOMItemID"`
	Drawings []CMFDrawing    `json:"drawings,omitempty" gorm:"foreignKey:CMFDesignID"`
}

// CMFDrawing CMF方案的工艺图纸
type CMFDrawing struct {
	ID          string    `json:"id" gorm:"primaryKey;size:32"`
	CMFDesignID string    `json:"cmf_design_id" gorm:"size:32;not null;index"`
	DrawingType string    `json:"drawing_type" gorm:"size:32"` // 丝印/激光雷雕/UV转印/移印/烫金
	FileID      string    `json:"file_id" gorm:"size:32;not null"`
	FileName    string    `json:"file_name" gorm:"size:256;not null"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
