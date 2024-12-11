package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey" json:"id"`
	AdminID     uint      `json:"admin_id"`
	SupplierID  uint      `json:"supplier_id"`
	ProductID   uint      `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int64     `json:"quantity"`
	Price       float64   `json:"price"`
	OrderDate   time.Time `json:"order_date"`
	Status      string    `json:"status"`
	Descriptiom string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Supplier    Supplier  `gorm:"foreignKey:SupplierID;references:ID"`
}

// Custom JSON Marshal for OrderDate
func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order // Create an alias to avoid recursion
	return json.Marshal(&struct {
		OrderDate string `json:"order_date"`
		Alias
	}{
		OrderDate: o.OrderDate.Format("2006-01-02 03:04 PM"),
		Alias:     (Alias)(o),
	})
}