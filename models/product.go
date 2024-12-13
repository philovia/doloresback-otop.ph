package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID               uint    `gorm:"primaryKey" json:"id"`
	SequentialNumber string  `json:"sequential_number"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Price            float64 `json:"price"`
	Quantity         int64   `json:"quantity"`
	SupplierID       uint    `json:"supplier_id"`
	Category         string  `json:"category"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	// Validate the category field
	if p.Category != "Food" && p.Category != "Non-Food" {
		return errors.New("category must be either 'Food' or 'Non-Food'")
	}

	// Generate custom ID starting at 10000
	var maxID *uint
	result := tx.Table("products").Select("MAX(id)").Scan(&maxID)
	if result.Error != nil {
		return result.Error
	}

	if maxID == nil || *maxID < 10000 {
		p.ID = 10000
	} else {
		p.ID = *maxID + 1
	}

	// Generate Sequential Number (e.g., "P-10001", "P-10002", etc.)
	var maxSequentialNumber *int64
	result = tx.Table("products").Select("MAX(CAST(SUBSTRING(sequential_number, 3) AS BIGINT))").Scan(&maxSequentialNumber)
	if result.Error != nil {
		return result.Error
	}

	// Generate sequential number based on the last max sequential number
	if maxSequentialNumber == nil || *maxSequentialNumber == 0 {
		p.SequentialNumber = fmt.Sprintf("P-%d", 10001) // Start from P-10001
	} else {
		p.SequentialNumber = fmt.Sprintf("P-%d", *maxSequentialNumber+1)
	}

	return nil
}
