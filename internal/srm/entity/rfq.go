package entity

import "time"

// RFQ 询价单
type RFQ struct {
	ID        string    `json:"id" gorm:"primaryKey;size:32"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RFQQuote 询价报价
type RFQQuote struct {
	ID        string    `json:"id" gorm:"primaryKey;size:32"`
	RFQID     string    `json:"rfq_id" gorm:"size:32"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
