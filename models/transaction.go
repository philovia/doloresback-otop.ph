package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	Total            float64           `json:"total"`                                             // Total cost of the transaction
	Received         float64           `json:"received"`                                          // Amount received from the customer
	Change           float64           `json:"change"`                                            // Change returned to the customer
	SupplierID       uint              `json:"supplier_id"`                                       // Foreign key linking to the supplier
	Supplier         Supplier          `json:"supplier"`                                          // Relation to Supplier
	TransactionItems []TransactionItem `json:"transaction_items" gorm:"foreignKey:TransactionID"` // Relation to transaction items
	CreatedAt        time.Time         `json:"created_at"`                                        // Transaction date
}

type TransactionItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID uint      `json:"transaction_id"`
	ProductID     uint      `json:"product_id"`
	Quantity      int64     `json:"quantity"`
	Price         float64   `json:"price"`
	Total         float64   `json:"total"`
	SupplierID    uint      `json:"supplier_id"` // Add SupplierID to associate with each product
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TransactionSupplier struct {
	ID            uint `gorm:"primaryKey"`
	TransactionID uint
	SupplierID    uint
	Transaction   Transaction `gorm:"foreignKey:TransactionID"`
	Supplier      Supplier    `gorm:"foreignKey:SupplierID"`
}
