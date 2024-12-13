package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type OtopProducts struct {
	gorm.Model
	ID               uint     `gorm:"primaryKey;autoIncrement"` // Remove 'unsigned' here
	Name             string   `json:"name"`
	Description      string   `json:"description" gorm:"not null;unique"`
	Price            float64  `json:"price"`
	Quantity         int64    `json:"quantity"`
	Category         string   `json:"category"`
	SupplierID       uint     `gorm:"not null"`
	Supplier         Supplier `gorm:"foreignKey:SupplierID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"supplier"`
	StoreName        string   `json:"store_name"`
	SequentialNumber string   `json:"sequential_number"`
}

func (p *OtopProducts) BeforeCreate(tx *gorm.DB) (err error) {
	// Validate the category field
	if p.Category != "Food" && p.Category != "Non-Food" {
		return errors.New("category must be either 'Food' or 'Non-Food'")
	}

	// Generate custom ID starting at 10000
	var maxID *uint // Changed to pointer to handle NULL value
	result := tx.Table("otop_products").Select("MAX(id)").Scan(&maxID)
	if result.Error != nil {
		return result.Error
	}

	if maxID == nil || *maxID < 10000 {
		p.ID = 10000
	} else {
		p.ID = *maxID + 1
	}

	// Generate Sequential Number (e.g., "P-10001", "P-10002", etc.)
	var maxSequentialNumber *int64 // Changed to pointer to handle NULL value
	result = tx.Table("otop_products").Select("MAX(CAST(SUBSTRING(sequential_number, 3) AS BIGINT))").Scan(&maxSequentialNumber)
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

func (p *OtopProducts) BeforeSave(tx *gorm.DB) (err error) {
	if p.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	return nil
}
