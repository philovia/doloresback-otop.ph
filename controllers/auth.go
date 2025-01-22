package controllers

import (
	// "fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/m/database"
	"github.com/m/models"
	"github.com/m/utils"
	"gopkg.in/gomail.v2"
	// "golang.org/x/crypto/bcrypt"
	// "gopkg.in/gomail.v2"
)

func Register(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user data"})
	}

	// Check if the email is already registered
	var existingUser models.User
	if err := database.DB.Where("username = ?", user.UserName).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email is already registered"})
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

	// Save the user to the database
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error saving user"})
	}

	// Send email notification
	m := gomail.NewMessage()
	m.SetHeader("From", "giemacazar@gmail.com")
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "Registration Confirmation")
	m.SetBody("text/plain", "Hello "+user.UserName+",\n\nYour registration was successful. Welcome to our platform!")

	d := gomail.NewDialer("smtp.gmail.com", 587, "giemacazar@gmail.com", "wtga mbuz ooxc ymzs")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error sending registration email"})
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

		// Return the token, role, and supplier details for the supplier
		return c.JSON(fiber.Map{
			"token":        token,
			"role":         "supplier",
			"supplier_id":  storedSupplier.ID,
			"store_name":   storedSupplier.StoreName,
			"phone_number": storedSupplier.PhoneNumber,
			"addres":       storedSupplier.Address,
			"id":           storedSupplier.ID,
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
	token, err := utils.GenerateToken(storedUser.UserName, storedUser.Role, storedUser.ID, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	// Return the token and role for the user
	return c.JSON(fiber.Map{
		"token": token,
		"role":  storedUser.Role,
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
