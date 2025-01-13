package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/m/controllers"
	middleware "github.com/m/middleware"
)

func UserRoutes(app *fiber.App) {

	// public routes (DONE)
	api := app.Group("/api")
	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.UnifiedLogin)
	api.Post("/logout", controllers.Logout)

	// Admin-only routes(DONE)
	supplier := app.Group("/supplier")
	supplier.Use(middleware.JWTProtected)
	supplier.Use(middleware.IsAdmin)
	supplier.Post("/", controllers.CreateSupplier)
	supplier.Get("/", controllers.GetSuppliers)
	supplier.Get("/:storeName", controllers.GetSupplierByStoreName)
	supplier.Put("/:storeName", controllers.UpdateSupplier)
	supplier.Delete("/:storeName", controllers.DeleteSupplier)

	// Add products by supplier (DONE)
	supplierRoutes := app.Group("/products", middleware.IsSupplier, middleware.JWTProtected)
	supplierRoutes.Post("/", controllers.AddProduct)
	supplierRoutes.Get("/", controllers.GetProducts)
	supplierRoutes.Get("/:supplier_id", controllers.GetProductByName) // search by name
	supplierRoutes.Put("/:id", controllers.UpdateProduct)
	supplierRoutes.Delete("/:id", controllers.DeleteProduct)
	supplierRoutes.Get("/", controllers.GetSupplierProducts)
	supplierRoutes.Get("/", controllers.GetMyProducts)
	supplierRoutes.Post("/confirm/:id", controllers.ConfirmOrders)            //supplier will confirm the order from admin by order id
	supplierRoutes.Get("/orders/:supplier_id", controllers.GetSupplierOrders) // with toke of the suppliers unique

	// the supplier will confirmed the order from admin(NOT YET)
	app.Put("/orders/:id/confirm", controllers.ConfirmOrder)
	app.Get("/orders/:supplier_id", controllers.GetSupplierOrders)
	app.Get("/suppliers/all_purchases", controllers.GetSupplierPurchaseCounts)             // for addmin only with limit of 6 suppliers
	app.Get("/suppliers/all_purchase", controllers.GetSupplierPurchaseCount)               // all suppliers will get puschased
	app.Get("/suppliers/purchases/:id", controllers.GetSupplierPurchasesByID)              // for admin only
	app.Get("/supplier/purchases", middleware.IsSupplier, controllers.GetMyTotalPurchases) // for suppliers

	//Can Get Total Otop Products Stocks & Name(DONE)
	app.Post("/api/otop/add_products", controllers.CreateOtopProduct) // create new otop products (USED)
	app.Get("/api/otop/products", controllers.GetOtopProducts)        // get all otop products (USED)
	app.Delete("/api/otop/:id", controllers.DeleteOtopProduct)        // delete product using drop down (USED)
	app.Put("/api/otop/:id", controllers.UpdateOtopProduct)
	app.Get("/api/otop/total_quantity", controllers.GetOtopTotalQuantity)                                            // total of otop products quantity of all products
	app.Get("/api/otop/total_quantity_name", controllers.GetOtopTotalQuantityName)                                   // diffrerent store name and total  quantity of products
	app.Get("/api/otop/total_products", controllers.GetOtopTotalProducts)                                            // total number of products(USED)
	app.Get("/api/otop/total_categories", controllers.GetTotalProductsByCategory)                                    // total products on food and non-food(USED)
	app.Get("/api/otop/total_suppliers", controllers.GetTotalSuppliers)                                              // count all suppliers(USED)
	app.Get("/api/otop/total_suppliers_product", controllers.GetSupplierProductCounts)                               // supplier and number of products
	app.Get("/api/otop/total_amount_suppliers/:supplier_id/total_amount", controllers.GetTotalPurchasedBySupplierID) // total of every suppliers amount of purchased

	app.Post("/api/otop/sold_items", controllers.RecordSoldItem)                           // makinng solds POS
	app.Get("/api/otop/solds_products", controllers.GetAllSoldItems)                       // the total solds
	app.Get("/api/otop/solds_products/:supplier_id", controllers.GetSoldItemsBySupplierID) // by supplier solds
	app.Post("/api/otop/add_cart", controllers.AddToCartHandler)
	// Order Management for the admin with supplier (DONE)
	admin := app.Group("/order")
	admin.Post("/", controllers.CreateOrder)
	admin.Get("/", controllers.GetOrders)
	admin.Get("/:id", controllers.GetOrder)
	admin.Put("/:id", controllers.UpdateOrder)
	admin.Delete("/:id", controllers.DeleteOrder)

	//for admin and cashier (DONE)
	app.Get("/api/products/total_quantity", controllers.GetTotalQuantity)
	app.Get("/api/products", controllers.GetProducts)
	app.Get("/api/products/supplier/:supplier_id", controllers.GetProductsByStore)

	app.Post("/api/otop/POS", controllers.POSController)
}
