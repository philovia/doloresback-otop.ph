package models

import (
	"time"

	"gorm.io/gorm"
)

type SoldItems struct {
	gorm.Model
	ID           uint        `gorm:"primaryKey;autoIncrement"`
	ProductID    uint        `gorm:"not null"`
	Product      OtopProducts `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"product"`
	QuantitySold int64       `json:"quantity_sold"`
	TotalAmount  float64     `json:"total_amount"`
	SoldDate     time.Time   `json:"sold_date"`
	SupplierID   uint        `gorm:"not null"` // SupplierID will be populated based on the OtopProduct
}

func (s *SoldItems) BeforeCreate(tx *gorm.DB) (err error) {
	// Ensure the SoldDate is set to the current time if not provided
	if s.SoldDate.IsZero() {
		s.SoldDate = time.Now()
	}

	// Fetch the associated OtopProduct to get the SupplierID
	var product OtopProducts
	if err := tx.First(&product, s.ProductID).Error; err != nil {
		return err // Return error if product is not found
	}

	// Set the SupplierID from the OtopProduct
	s.SupplierID = product.SupplierID

	return nil
}
