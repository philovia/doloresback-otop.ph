package controllers

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	// "github.com/golang-jwt/jwt/v4"
	"github.com/m/database"
	"github.com/m/models"
	// "gorm.io/gorm"
)

func CreateOtopProduct(c *fiber.Ctx) error {
	var otopProduct models.OtopProducts

	// Parse the request body
	if err := c.BodyParser(&otopProduct); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product data"})
	}

	// Fetch the supplier by store_name
	var supplier models.Supplier
	if err := database.DB.Where("store_name = ?", otopProduct.StoreName).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Store not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Increment the Purchased count
	supplier.Purchased += 1
	if err := database.DB.Save(&supplier).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update purchased count"})
	}

	var lastProduct models.OtopProducts
	err := database.DB.Raw("SELECT * FROM otop_products ORDER BY created_at DESC LIMIT 1").Scan(&lastProduct).Error
	if err != nil {
		log.Println("Failed to fetch last product:", err)
	}

	// Generate new sequential number
	seqNumber := 1
	if lastProduct.ID != 0 {
		// Extract the sequential number from the last product's sequential_number
		// Assuming format is like 'SEQ-0001'
		fmt.Sscanf(lastProduct.SequentialNumber, "SP-%04d", &seqNumber)
		seqNumber++
	}
	otopProduct.SequentialNumber = fmt.Sprintf("SP-%04d", seqNumber)

	// Assign the SupplierID from the fetched supplier
	otopProduct.SupplierID = supplier.ID

	// Save the product
	otopProduct.CreatedAt = time.Now()
	if err := database.DB.Create(&otopProduct).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create product"})
	}

	// Preload the related supplier data and return the product
	if err := database.DB.Preload("Supplier").First(&otopProduct, otopProduct.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch related supplier"})
	}

	return c.Status(fiber.StatusCreated).JSON(otopProduct)
}

func GetOtopProducts(c *fiber.Ctx) error {
	var otopProducts []models.OtopProducts

	// Use Preload to fetch related Supplier data
	if err := database.DB.Preload("Supplier").Find(&otopProducts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch products"})
	}

	// Return the products with related supplier data as JSON
	return c.JSON(otopProducts)
}

func GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var otopProduct models.OtopProducts

	// Fetch the product by ID
	database.DB.First(&otopProduct, id)

	if otopProduct.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found",
		})
	}

	// Return the product as JSON
	return c.JSON(otopProduct)
}

func UpdateOtopProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var otopProduct models.OtopProducts

	// Find the existing product
	database.DB.First(&otopProduct, id)

	if otopProduct.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found",
		})
	}

	// Parse the new product data
	if err := c.BodyParser(&otopProduct); err != nil {
		return err
	}

	// Save the updated product
	if err := database.DB.Save(&otopProduct).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product"})
	}

	// Return the updated product
	return c.JSON(otopProduct)
}

func DeleteOtopProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var otopProduct models.OtopProducts

	// Find the product by ID
	database.DB.First(&otopProduct, id)

	if otopProduct.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found",
		})
	}

	// Delete the product
	database.DB.Delete(&otopProduct)

	// Return a success message
	return c.JSON(fiber.Map{
		"message": "Product deleted",
	})
}

func GetOtopTotalQuantity(c *fiber.Ctx) error {
	log.Println("Received request to calculate total quantity of products")

	var totalQuantity int64
	err := database.DB.Model(&models.OtopProducts{}).Select("SUM(quantity)").Scan(&totalQuantity).Error

	if err != nil {
		log.Println("Error calculating total quantity:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to calculate total quantity"})
	}

	log.Println("Total quantity calculated:", totalQuantity)
	return c.JSON(fiber.Map{"total_quantity": totalQuantity})
}

func GetOtopProductByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var product models.OtopProducts

	// Fetch the product by ID
	if err := database.DB.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Product not found",
		})
	}

	// Return the product as JSON
	return c.JSON(product)
}

func GetOtopTotalQuantityName(c *fiber.Ctx) error {
	log.Println("Received request to calculate total quantity of products by name")

	var result []struct {
		ProductName   string `json:"product_name"`
		TotalQuantity int64  `json:"total_quantity"`
	}

	err := database.DB.Model(&models.OtopProducts{}).
		Select("name as product_name, SUM(quantity) as total_quantity").
		Group("name").
		Scan(&result).Error

	if err != nil {
		log.Println("Error calculating total quantity by product name:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to calculate total quantity"})
	}

	log.Println("Total quantity calculated by product name:", result)
	return c.JSON(result)
}

func GetOtopTotalProducts(c *fiber.Ctx) error {
	log.Println("Received request to calculate total number of unique products")

	var result struct {
		TotalProducts int64 `json:"total_products"`
	}

	err := database.DB.Model(&models.OtopProducts{}).
		Select("COUNT(DISTINCT name) as total_products").
		Scan(&result).Error

	if err != nil {
		log.Println("Error calculating total number of products:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to calculate total products"})
	}

	log.Println("Total number of products calculated:", result.TotalProducts)
	return c.JSON(result)
}

func GetTotalProductsByCategory(c *fiber.Ctx) error {
	var foodCount, nonFoodCount int64

	if err := database.DB.Model(&models.OtopProducts{}).Where("category = ?", "Food").Count(&foodCount).Error; err != nil {
		log.Println("Error counting food products:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count food products"})
	}

	if err := database.DB.Model(&models.OtopProducts{}).Where("category = ?", "Non-Food").Count(&nonFoodCount).Error; err != nil {
		log.Println("Error counting non-food products:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count non-food products"})
	}

	return c.JSON(fiber.Map{
		"Food":     foodCount,
		"Non-Food": nonFoodCount,
	})

}

func GetTotalPurchasedBySupplierID(c *fiber.Ctx) error {
	// Get Supplier ID from the request params
	supplierID := c.Params("id")
	if supplierID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Supplier ID is required"})
	}

	var totalPurchased float64

	// Query to calculate total purchased for the given supplier ID
	err := database.DB.Model(&models.OtopProducts{}).
		Where("supplier_id = ?", supplierID).
		Select("SUM(price * quantity * purchased)"). // Include the purchased field in the calculation
		Scan(&totalPurchased).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"supplier_id":     supplierID,
		"total_purchased": totalPurchased,
	})
}

// function for POS

func RecordSoldItem(c *fiber.Ctx) error {
	var soldItem models.SoldItems
	var otopProduct models.OtopProducts

	// Parse the request body
	if err := c.BodyParser(&soldItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid sold item data"})
	}

	// Fetch the product being sold
	if err := database.DB.First(&otopProduct, soldItem.ProductID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	// Check if sufficient quantity is available
	if otopProduct.Quantity < soldItem.QuantitySold {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient product quantity"})
	}

	// Calculate the total amount for the sold item
soldItem.TotalAmount = float64(soldItem.QuantitySold) * otopProduct.Price


	// Update the product quantity
	otopProduct.Quantity -= soldItem.QuantitySold
	if err := database.DB.Save(&otopProduct).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product quantity"})
	}

	// Create the sold item record
	if err := database.DB.Create(&soldItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to record sold item"})
	}

	// Return the sold item data
	return c.Status(fiber.StatusCreated).JSON(soldItem)
}

func GetAllSoldItems(c *fiber.Ctx) error {
	var soldItems []models.SoldItems
	if err := database.DB.Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch sold items"})
	}
	return c.Status(fiber.StatusOK).JSON(soldItems)
}

// GetSoldItemsBySupplierID retrieves sold items filtered by SupplierID
func GetSoldItemsBySupplierID(c *fiber.Ctx) error {
	supplierID := c.Params("supplier_id") // Get SupplierID from URL params

	var soldItems []models.SoldItems
	// Fetch sold items by SupplierID
	if err := database.DB.Where("supplier_id = ?", supplierID).Preload("Product").Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch sold items for the supplier"})
	}

	return c.Status(fiber.StatusOK).JSON(soldItems)
}