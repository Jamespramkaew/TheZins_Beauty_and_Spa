package container

// For Dependency Injection

import (
	"backend/config"
	"backend/internal/customer"
	"backend/internal/order"

	"gorm.io/gorm"
)

type Container struct {
	DB *gorm.DB

	// Handlers
	CustomerHandler *customer.Handler
	OrderHandler    *order.Handler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
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
