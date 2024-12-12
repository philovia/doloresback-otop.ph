package controllers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/m/database"
	"github.com/m/models"
	"github.com/m/utils"
)

func Register(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user data"})
	}

	// Check if the role is either "admin" or "cashier"
	if user.Role != "admin" && user.Role != "cashier" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Only 'admin' and 'cashier' roles are allowed."})
	}

	// Validate the username based on the role
	if user.Role == "admin" && !strings.HasSuffix(user.UserName, "_admin") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Admin username must end with '_admin'"})
	}
	if user.Role == "cashier" && !strings.HasSuffix(user.UserName, "_cashier") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cashier username must end with '_cashier'"})
	}

	// Save user to the database
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error saving user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

func UnifiedLogin(c *fiber.Ctx) error {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body to get credentials
	if err := c.BodyParser(&creds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid login data"})
	}

	// Check if the email exists in the supplier table first
	var storedSupplier models.Supplier
	if err := database.DB.Where("email = ?", creds.Email).First(&storedSupplier).Error; err == nil {
		// If the supplier exists, check if the password matches
		if storedSupplier.Password != creds.Password {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		// Generate token for the supplier
		token, err := utils.GenerateToken(storedSupplier.StoreName, "supplier", storedSupplier.ID, storedSupplier.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
		}

		// Return the token, role, and supplier_id for the supplier
		return c.JSON(fiber.Map{
			"token":       token,
			"role":        "supplier",
			"supplier_id": storedSupplier.ID, // Add supplier ID to the response
		})
	}

	// If not found in the supplier table, check in the user table
	var storedUser models.User
	if err := database.DB.Where("email = ?", creds.Email).First(&storedUser).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	// If the password does not match for the user
	if storedUser.Password != creds.Password {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate token for the user
	token, err := utils.GenerateToken(storedUser.UserName, storedUser.Role, storedUser.ID, storedSupplier.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	// Return the token, role, and supplier_id (optional for non-supplier users)
	return c.JSON(fiber.Map{
		"token":       token,
		"role":        storedUser.Role,
		"supplier_id": storedSupplier.ID, // Return the supplier_id in case it's available
	})
}

func Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Logout successful, please clear your token",
	})
}

func SupplierLogin(c *fiber.Ctx) error {
	var creds models.Supplier
	if err := c.BodyParser(&creds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid login data"})
	}

	var storedSupplier models.Supplier
	if err := database.DB.Where("email = ?", creds.Email).First(&storedSupplier).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Supplier not found"})
	}

	if storedSupplier.Password != creds.Password {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := utils.GenerateToken(storedSupplier.StoreName, "supplier", storedSupplier.ID, storedSupplier.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	return c.JSON(fiber.Map{"token": token, "role": "supplier"})
}

func SupplierLogout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Supplier logout successful, please clear your token",
	})
}
