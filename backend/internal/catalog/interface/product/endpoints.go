package productHandler

const (
	// Base paths
	ProductsBasePath      = "/products"
	AdminProductsBasePath = "/admin/products"

	// === Product Resource Paths ===
	IDPath = "/{id}"

	// === Public Product Endpoints ===

	// Get all published products (public access)
	GetAllPublishedProductsEndpoint = ProductsBasePath

	// Get product by ID (public access)
	GetProductByIDEndpoint = ProductsBasePath + IDPath

	// === Admin-Only Endpoints ===

	// Get all products including drafts (admin only)
	GetAdminAllProductsEndpoint = AdminProductsBasePath

	// Retrieves a product by ID, whatever its status (admin only)
	GetAdminProductByIDEndpoint = AdminProductsBasePath + IDPath

	// Create product with price (admin only)
	CreateProductWithPriceEndpoint = AdminProductsBasePath

	// Modify product (admin only)
	ModifyProductEndpoint = AdminProductsBasePath + IDPath

	// Remove product (admin only)
	RemoveProductEndpoint = AdminProductsBasePath + IDPath
)
