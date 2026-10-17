package container

// For Dependency Injection

import (
	"backend/config"
	"backend/internal/authentication"
	"backend/internal/customer"
	"backend/internal/order"

	"gorm.io/gorm"
)

type Container struct {
	DB *gorm.DB

	// Handlers
	AuthHandler     *authentication.Handler
	CustomerHandler *customer.Handler
	OrderHandler    *order.Handler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Initialize Authentication Module
	authRepository := authentication.NewRepository(db)
	authService := authentication.NewService(authRepository)
	authHandler := authentication.NewHandler(authService)

	// Initialize Customer Module
	customerRepository := customer.NewRepository(db)
	customerService := customer.NewService(customerRepository)
	customerHandler := customer.NewHandler(customerService)

	// Initialize Order Module
	orderRepository := order.NewRepository(db)
	orderService := order.NewService(orderRepository)
	orderHandler := order.NewHandler(orderService)

	return &Container{
		DB:              db,
		AuthHandler:     authHandler,
		CustomerHandler: customerHandler,
		OrderHandler:    orderHandler,
	}
}

func (c *Container) Close() error {
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
