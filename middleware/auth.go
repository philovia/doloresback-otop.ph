package middleware

import (
	// "os"÷\\\
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/m/utils"
)

func JWTProtected(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")
	token, err := utils.ParseToken(tokenString)
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized Pleased Input Token"})
	}
	c.Locals("user", token)
	return c.Next()
}
func IsAdmin(c *fiber.Ctx) error {
	// Check if token is already set in Locals
	userToken := c.Locals("user")
	if userToken == nil {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// Remove "Bearer " prefix and parse the token
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		parsedToken, err := utils.ParseToken(tokenStr)
		if err != nil || !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// Store the parsed token in Locals
		c.Locals("user", parsedToken)
		userToken = parsedToken
	}

	// Type assertion to retrieve claims
	claims, ok := userToken.(*jwt.Token).Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check for the "admin" role in claims
	if role, exists := claims["role"].(string); !exists || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	return c.Next()
}
func IsSupplier(c *fiber.Ctx) error {
	// Get the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	// Remove "Bearer " prefix and parse the token
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	userToken, err := utils.ParseToken(tokenStr)
	if err != nil || !userToken.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Retrieve the claims from the token
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check if the role claim exists and is "supplier"
	if role, exists := claims["role"].(string); !exists || role != "supplier" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if supplierID, exists := claims["supplier_id"].(float64); exists {
		// Store the supplier_id and token in Locals
		c.Locals("supplier_id", uint(supplierID)) // Store supplier_id in Locals
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Supplier ID not found in token"})
	}

	// Store token in Locals for further use
	c.Locals("supplier", userToken)
	return c.Next()
}
func IsCashier(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	userToken, err := utils.ParseToken(tokenStr)
	if err != nil || !userToken.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if role, exists := claims["role"].(string); !exists || role != "cashier" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	c.Locals("cashier", userToken)

	return c.Next()
}
