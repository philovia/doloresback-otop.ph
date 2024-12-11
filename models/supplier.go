package models

import "gorm.io/gorm"

type Supplier struct {
	gorm.Model
	ID           uint           `json:"id" gorm:"primary_key"`
	StoreName    string         `json:"store_name" gorm:"unique"`
	Email        string         `json:"email"`
	PhoneNumber  string         `json:"phone_number"`
	Address      string         `json:"address"`
	Password     string         `json:"password"`
	Role         string         `gorm:"default:'supplier'"`
	Products     []Product      `gorm:"foreignKey:SupplierID"`
	Purchased    int            `gorm:"default:0"`
	OtopProducts []OtopProducts `gorm:"foreignKey:SupplierID"`
}

// TableName overrides the default table name used by GORM
func (Supplier) TableName() string {
	return "suppliers"
}
