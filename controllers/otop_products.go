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

func GetTopSoldProducts(c *fiber.Ctx) error {
	type ProductSales struct {
		ProductID    uint    `json:"product_id"`
		Name         string  `json:"name"`
		QuantitySold int     `json:"quantity_sold"`
		Price        float64 `json:"price"`
		TotalAmount  float64 `json:"total_amount"`
	}

	var topProducts []ProductSales

	// Join SoldItems and Products, group by ProductID, and order by total quantity sold
	if err := database.DB.Table("sold_items").
		Select("sold_items.product_id, products.name, products.price, SUM(sold_items.quantity_sold) as quantity_sold, SUM(sold_items.quantity_sold * products.price) as total_amount").
		Joins("JOIN products ON products.id = sold_items.product_id").
		Group("sold_items.product_id, products.name, products.price").
		Order("quantity_sold DESC").
		Limit(3).
		Scan(&topProducts).Error; err != nil {
		log.Println("Error fetching top sold products:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch top sold products",
		})
	}

	return c.JSON(fiber.Map{
		"top_sold_products": topProducts,
	})
}

// This is the function that call the every handler it should be store on the controller folder
func CallUpdateOtopProduct(p models.OtopProducts) error {
	// Use GORM to update the product
	result := database.DB.Model(&models.OtopProducts{}).Where("supplier_id = ? AND store_name = ?", p.SupplierID, p.StoreName).
		Updates(models.OtopProducts{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Category:    p.Category,
			Quantity:    p.Quantity,
		})

	if result.Error != nil {
		log.Println("Error executing update:", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows were updated, please check the supplier_id and store_name")
	}

	return nil
}

// Fiber handler that call all the fucntion from the contorller this will be put inside the handler folder
func UpdateOtopProductHandler(c *fiber.Ctx) error {
	var updateReq models.OtopProducts

	// Parse the request body into the OtopProducts model
	if err := c.BodyParser(&updateReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Call the update function
	if err := CallUpdateOtopProduct(updateReq); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update product",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Product updated successfully",
	})
}

func CreateOtopProduct(c *fiber.Ctx) error {
	var otopProduct models.OtopProducts

	// Parse the request body
	if err := c.BodyParser(&otopProduct); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product data"})
	}

	// Validate description
	if otopProduct.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Description is required"})
	}

	// Check for duplicate description
	var existingProduct models.OtopProducts
	if err := database.DB.Where("description = ?", otopProduct.Description).First(&existingProduct).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Description must be unique"})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
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

// the correct one to handle POS
func RecordSoldItem(c *fiber.Ctx) error {
	var soldItems []models.SoldItems
	var responses []map[string]interface{}

	// Parse the request body (expecting a JSON array)
	if err := c.BodyParser(&soldItems); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid sold items data",
		})
	}

	for _, item := range soldItems {
		var otopProduct models.OtopProducts

		// Fetch product using ProductID
		if err := database.DB.First(&otopProduct, item.ProductID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Product with ID %d not found", item.ProductID),
			})
		}

		// Check stock
		if otopProduct.Quantity < item.QuantitySold {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Insufficient quantity for product ID %d", item.ProductID),
			})
		}

		// Set total amount and sold date
		item.TotalAmount = float64(item.QuantitySold) * otopProduct.Price
		item.SoldDate = time.Now()

		// Set SupplierID from product
		item.SupplierID = otopProduct.SupplierID

		// Reduce product stock
		otopProduct.Quantity -= item.QuantitySold
		if err := database.DB.Save(&otopProduct).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update product quantity",
			})
		}

		// Save the sold item
		if err := database.DB.Create(&item).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to record sold item",
			})
		}

		// Reload the sold item with preloaded product and supplier
		var fullSoldItem models.SoldItems
		if err := database.DB.
			Preload("Product").
			Preload("Product.Supplier").
			First(&fullSoldItem, item.ID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch full sold item",
			})
		}

		// Append the full response
		responses = append(responses, map[string]interface{}{
			"soldItem": fullSoldItem,
			"supplier": fullSoldItem.Product.Supplier,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(responses)
}

// // function for POS
// func RecordSoldItem(c *fiber.Ctx) error {
// 	var soldItem models.SoldItems
// 	var otopProduct models.OtopProducts
// 	var supplier models.Supplier

// 	// Parse the request body
// 	if err := c.BodyParser(&soldItem); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid sold item data"})
// 	}

// 	// Fetch the product being sold
// 	if err := database.DB.First(&otopProduct, soldItem.ProductID).Error; err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
// 	}

// 	// Fetch the supplier information (assuming the product has a SupplierID)
// 	if err := database.DB.First(&supplier, otopProduct.SupplierID).Error; err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
// 	}

// 	// Check if sufficient quantity is available
// 	if otopProduct.Quantity < soldItem.QuantitySold {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient product quantity"})
// 	}

// 	// Calculate the total amount for the sold item
// 	soldItem.TotalAmount = float64(soldItem.QuantitySold) * otopProduct.Price

// 	// Update the product quantity
// 	otopProduct.Quantity -= soldItem.QuantitySold
// 	if err := database.DB.Save(&otopProduct).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product quantity"})
// 	}

// 	// Create the sold item record
// 	if err := database.DB.Create(&soldItem).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to record sold item"})
// 	}

// 	// Include the supplier information in the response
// 	response := map[string]interface{}{
// 		"soldItem": soldItem,
// 		"supplier": supplier,
// 	}

// 	// Return the sold item data along with supplier details
// 	return c.Status(fiber.StatusCreated).JSON(response)
// }

// func GetAllSoldItems(c *fiber.Ctx) error {
// 	var soldItems []models.SoldItems
// 	if err := database.DB.Find(&soldItems).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch sold items"})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(soldItems)
// }

// func GetAllSoldItems(c *fiber.Ctx) error {
// 	var soldItems []models.SoldItems

// 	// Fetch all sold items and preload related OtopProduct data
// 	if err := database.DB.Preload("Product").Find(&soldItems).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch sold items",
// 		})
// 	}

// 	return c.JSON(soldItems)
// }

func GetAllSoldItems(c *fiber.Ctx) error {
	var soldItems []models.SoldItems

	// Fetch all sold items and preload product data
	if err := database.DB.Preload("Product").Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items",
		})
	}

	// Calculate overall amount sold
	var overallAmountSold float64
	for _, item := range soldItems {
		overallAmountSold += float64(item.QuantitySold) * item.Product.Price
	}

	return c.JSON(fiber.Map{
		"sold_items":          soldItems,
		"overall_amount_sold": overallAmountSold,
	})
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

func AddToCartHandler(c *fiber.Ctx) error {
	// Define the request body structure
	type AddToCartRequest struct {
		ProductID  int `json:"productID"`
		SupplierID int `json:"supplierID"`
		Quantity   int `json:"quantity"`
	}

	// Parse the request body into AddToCartRequest struct
	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request data"})
	}

	// Check if product ID and supplier ID are valid
	if req.ProductID == 0 || req.SupplierID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid product or supplier ID: Product ID: %d, Supplier ID: %d", req.ProductID, req.SupplierID),
		})
	}

	// Retrieve product by ID
	var otopProduct models.OtopProducts
	productResult := database.DB.First(&otopProduct, req.ProductID)
	if productResult.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	// Retrieve supplier by ID
	var supplier models.Supplier
	supplierResult := database.DB.First(&supplier, req.SupplierID)
	if supplierResult.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}

	// Further validation (optional)
	if otopProduct.SupplierID != supplier.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Product and supplier mismatch: Product SupplierID: %d, Supplier ID: %d", otopProduct.SupplierID, supplier.ID),
		})
	}

	// Logic to add to cart (this part can be customized as needed)
	// For example: Add the product to the user's cart in the database

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Item added to cart successfully",
	})
}

func POSController(c *fiber.Ctx) error {
	type CartItem struct {
		ProductID uint    `json:"product_id"`
		Name      string  `json:"name"`
		Quantity  int64   `json:"quantity"`
		Price     float64 `json:"price"`
		Total     float64 `json:"total"`
	}

	type CheckoutRequest struct {
		Items    []CartItem `json:"items"`
		Received float64    `json:"received"`
		Total    float64    `json:"total"`
		Change   float64    `json:"change"`
	}

	var products []models.OtopProducts
	search := c.Query("search", "")

	// Fetch products along with supplier information
	if search != "" {
		database.DB.Preload("Supplier").Where("name LIKE ?", "%"+search+"%").Find(&products)
	} else {
		database.DB.Preload("Supplier").Find(&products)
	}

	// Return available products (GET request)
	if c.Method() == fiber.MethodGet {
		return c.JSON(products)
	}

	// Handle checkout request
	var request CheckoutRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid checkout data"})
	}

	// Validate received amount
	if request.Received < request.Total {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Received amount is less than the total"})
	}

	// Process the transaction
	var supplierIDs = make(map[uint]bool) // To store unique supplier IDs for the items
	var transactionItems []models.TransactionItem

	// Create the transaction first (without any items yet)
	transaction := models.Transaction{
		Total:     request.Total,
		Received:  request.Received,
		Change:    request.Change,
		CreatedAt: time.Now(),
	}

	// Iterate over items to validate suppliers
	for _, item := range request.Items {
		var product models.OtopProducts
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("Product ID %d not found", item.ProductID)})
		}

		// Check stock
		if product.Quantity < item.Quantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Insufficient stock for product %s", product.Name)})
		}

		// Deduct stock
		product.Quantity -= item.Quantity
		if err := database.DB.Save(&product).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product stock"})
		}

		// Add supplierID to the map if it exists (to ensure uniqueness)
		if product.SupplierID > 0 {
			supplierIDs[product.SupplierID] = true
		}

		// Create the transaction item
		transactionItem := models.TransactionItem{
			ProductID:     item.ProductID,
			TransactionID: transaction.ID, // Will be updated later after the transaction is saved
			Quantity:      item.Quantity,
			Price:         item.Price,
			Total:         item.Total,
		}
		transactionItems = append(transactionItems, transactionItem)
	}

	// Check if there are no suppliers for the transaction (if supplierIDs is empty)
	if len(supplierIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No valid suppliers found for the transaction"})
	}

	// Use the first supplier (or you can add your own logic to handle multiple suppliers)
	var supplierID uint
	for supplierID = range supplierIDs {
		break
	}

	// Assign a valid supplierID to the transaction
	transaction.SupplierID = supplierID

	// Insert the transaction into the database
	if err := database.DB.Create(&transaction).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	// Now that the transaction is created, we can assign the transaction ID to the items
	for _, item := range transactionItems {
		item.TransactionID = transaction.ID
		if err := database.DB.Create(&item).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save transaction items"})
		}
	}

	// Insert the unique supplier IDs linked to the transaction (using TransactionSupplier)
	for supplierID := range supplierIDs {
		transactionSupplier := models.TransactionSupplier{
			TransactionID: transaction.ID,
			SupplierID:    supplierID,
		}
		if err := database.DB.Create(&transactionSupplier).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to link suppliers to transaction"})
		}
	}

	// Prepare the receipt response with transaction details
	receipt := fiber.Map{
		"transaction_id": transaction.ID,
		"date":           transaction.CreatedAt.Format("2006-01-02 15:04:05"),
		"items":          transactionItems,
		"total":          request.Total,
		"received":       request.Received,
		"change":         request.Change,
	}

	// Return the receipt response as JSON
	return c.JSON(receipt)
}

type SummaryRequest struct {
	IntervalType string `json:"interval" validate:"required,oneof=daily weekly monthly yearly"`
}

func GetSalesSummary(c *fiber.Ctx) error {
	var req SummaryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	switch req.IntervalType {
	case "daily":
		return getDailySummary(c)
	case "weekly":
		return getWeeklySummary(c)
	case "monthly":
		return getMonthlySummary(c)
	case "yearly":
		return getYearlySummary(c)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Interval must be daily, weekly, monthly, or yearly",
		})
	}
}

// get data for daily
func getDailySummary(c *fiber.Ctx) error {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		// Sunday, move to last Monday
		now = now.AddDate(0, 0, -6)
	}

	// Start on Monday
	startOfWeek := now.AddDate(0, 0, -(int(now.Weekday()) - 1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	endOfWeek := startOfWeek.AddDate(0, 0, 5).Add(24*time.Hour - time.Nanosecond) // Saturday night

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("created_at BETWEEN ? AND ?", startOfWeek, endOfWeek).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items",
		})
	}

	// Group by Day
	dailySales := map[string]float64{
		"Monday":    0,
		"Tuesday":   0,
		"Wednesday": 0,
		"Thursday":  0,
		"Friday":    0,
		"Saturday":  0,
	}

	for _, item := range soldItems {
		day := item.CreatedAt.Weekday().String()
		if day != "Sunday" {
			dayName := string(day)
			dailySales[dayName] += float64(item.QuantitySold) * item.Product.Price
		}
	}

	return c.JSON(dailySales)
}

// get data for the week

func getWeeklySummary(c *fiber.Ctx) error {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("created_at >= ?", startOfMonth).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items",
		})
	}

	weeklySales := map[string]float64{
		"Week 1": 0,
		"Week 2": 0,
		"Week 3": 0,
		"Week 4": 0,
		"Week 5": 0,
	}

	for _, item := range soldItems {
		day := item.CreatedAt.Day()
		var week string
		switch {
		case day >= 1 && day <= 7:
			week = "Week 1"
		case day >= 8 && day <= 14:
			week = "Week 2"
		case day >= 15 && day <= 21:
			week = "Week 3"
		case day >= 22 && day <= 28:
			week = "Week 4"
		default:
			week = "Week 5"
		}
		weeklySales[week] += float64(item.QuantitySold) * item.Product.Price
	}

	return c.JSON(weeklySales)
}

// get data for the month
func getMonthlySummary(c *fiber.Ctx) error {
	now := time.Now()
	currentYear := now.Year()

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("EXTRACT(YEAR FROM created_at) = ?", currentYear).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items",
		})
	}

	monthlySales := map[string]float64{
		"January":   0,
		"February":  0,
		"March":     0,
		"April":     0,
		"May":       0,
		"June":      0,
		"July":      0,
		"August":    0,
		"September": 0,
		"October":   0,
		"November":  0,
		"December":  0,
	}

	for _, item := range soldItems {
		month := item.CreatedAt.Month().String()
		monthlySales[month] += float64(item.QuantitySold) * item.Product.Price
	}

	return c.JSON(monthlySales)
}

// yearly function for sold items chat
func getYearlySummary(c *fiber.Ctx) error {
	now := time.Now()
	startYear := now.Year() - 4 // last 4 years + current

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("EXTRACT(YEAR FROM created_at) >= ?", startYear).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items",
		})
	}

	yearlySales := make(map[int]float64)

	for i := startYear; i <= now.Year(); i++ {
		yearlySales[i] = 0
	}

	for _, item := range soldItems {
		year := item.CreatedAt.Year()
		if _, ok := yearlySales[year]; ok {
			yearlySales[year] += float64(item.QuantitySold) * item.Product.Price
		}
	}

	return c.JSON(yearlySales)
}

// supplier get summary

type SupplierSalesRequest struct {
	IntervalType string `json:"interval" validate:"required,oneof=daily weekly monthly yearly"`
	SupplierID   uint   `json:"supplier_id" validate:"required"`
}

func GetSupplierSalesSummary(c *fiber.Ctx) error {
	var req SupplierSalesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	switch req.IntervalType {
	case "daily":
		return getSupplierDailySummary(c, req.SupplierID)
	case "weekly":
		return getSupplierWeeklySummary(c, req.SupplierID)
	case "monthly":
		return getSupplierMonthlySummary(c, req.SupplierID)
	case "yearly":
		return getSupplierYearlySummary(c, req.SupplierID)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Interval must be daily, weekly, monthly, or yearly",
		})
	}
}

// Get daily sales for a specific supplier
func getSupplierDailySummary(c *fiber.Ctx, supplierID uint) error {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		// Sunday, move to last Monday
		now = now.AddDate(0, 0, -6)
	}

	// Start on Monday
	startOfWeek := now.AddDate(0, 0, -(int(now.Weekday()) - 1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	endOfWeek := startOfWeek.AddDate(0, 0, 5).Add(24*time.Hour - time.Nanosecond) // Saturday night

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("sold_date BETWEEN ? AND ? AND supplier_id = ?", startOfWeek, endOfWeek, supplierID).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items for the supplier",
		})
	}

	// Group by Day
	dailySales := map[string]float64{
		"Monday":    0,
		"Tuesday":   0,
		"Wednesday": 0,
		"Thursday":  0,
		"Friday":    0,
		"Saturday":  0,
	}

	for _, item := range soldItems {
		day := item.SoldDate.Weekday().String()
		if day != "Sunday" {
			dayName := string(day)
			dailySales[dayName] += float64(item.QuantitySold) * item.TotalAmount
		}
	}

	return c.JSON(dailySales)
}

// Get weekly sales for a specific supplier
func getSupplierWeeklySummary(c *fiber.Ctx, supplierID uint) error {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("sold_date >= ? AND supplier_id = ?", startOfMonth, supplierID).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items for the supplier",
		})
	}

	weeklySales := map[string]float64{
		"Week 1": 0,
		"Week 2": 0,
		"Week 3": 0,
		"Week 4": 0,
		"Week 5": 0,
	}

	for _, item := range soldItems {
		day := item.SoldDate.Day()
		var week string
		switch {
		case day >= 1 && day <= 7:
			week = "Week 1"
		case day >= 8 && day <= 14:
			week = "Week 2"
		case day >= 15 && day <= 21:
			week = "Week 3"
		case day >= 22 && day <= 28:
			week = "Week 4"
		default:
			week = "Week 5"
		}
		weeklySales[week] += float64(item.QuantitySold) * item.TotalAmount
	}

	return c.JSON(weeklySales)
}

// Get monthly sales for a specific supplier
func getSupplierMonthlySummary(c *fiber.Ctx, supplierID uint) error {
	now := time.Now()
	currentYear := now.Year()

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("EXTRACT(YEAR FROM sold_date) = ? AND supplier_id = ?", currentYear, supplierID).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items for the supplier",
		})
	}

	monthlySales := map[string]float64{
		"January":   0,
		"February":  0,
		"March":     0,
		"April":     0,
		"May":       0,
		"June":      0,
		"July":      0,
		"August":    0,
		"September": 0,
		"October":   0,
		"November":  0,
		"December":  0,
	}

	for _, item := range soldItems {
		month := item.SoldDate.Month().String()
		monthlySales[month] += float64(item.QuantitySold) * item.TotalAmount
	}

	return c.JSON(monthlySales)
}

// Get yearly sales for a specific supplier
func getSupplierYearlySummary(c *fiber.Ctx, supplierID uint) error {
	now := time.Now()
	startYear := now.Year() - 4 // last 4 years + current

	var soldItems []models.SoldItems
	if err := database.DB.Preload("Product").Where("EXTRACT(YEAR FROM sold_date) >= ? AND supplier_id = ?", startYear, supplierID).Find(&soldItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sold items for the supplier",
		})
	}

	yearlySales := make(map[int]float64)

	for i := startYear; i <= now.Year(); i++ {
		yearlySales[i] = 0
	}

	for _, item := range soldItems {
		year := item.SoldDate.Year()
		if _, ok := yearlySales[year]; ok {
			yearlySales[year] += float64(item.QuantitySold) * item.TotalAmount
		}
	}

	return c.JSON(yearlySales)
}

// fetch data using date

type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func GetSoldItemsByDateRangePost(c *fiber.Ctx) error {
	var dateRange DateRange
	var soldItems []models.SoldItems

	// Parse JSON body
	if err := c.BodyParser(&dateRange); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db := database.DB

	// If both start and end date are provided
	if dateRange.StartDate != "" && dateRange.EndDate != "" {
		startDate, err1 := time.Parse("2006-01-02", dateRange.StartDate)
		endDate, err2 := time.Parse("2006-01-02", dateRange.EndDate)
		if err1 != nil || err2 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid date format. Please use YYYY-MM-DD.",
			})
		}

		endDate = endDate.Add(24 * time.Hour)

		if err := db.Preload("Product").
			Where("created_at BETWEEN ? AND ?", startDate, endDate).
			Find(&soldItems).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch sold items",
			})
		}
	} else {
		// No date provided, fetch all
		if err := db.Preload("Product").Find(&soldItems).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch sold items",
			})
		}
	}

	// Calculate overall amount sold
	var overallAmountSold float64
	for _, item := range soldItems {
		overallAmountSold += float64(item.QuantitySold) * item.Product.Price
	}

	return c.JSON(fiber.Map{
		"sold_items":          soldItems,
		"overall_amount_sold": overallAmountSold,
	})
}
