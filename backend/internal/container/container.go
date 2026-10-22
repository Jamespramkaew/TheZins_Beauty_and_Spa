package container

// For Dependency Injection

import (
	"backend/config"
	"backend/internal/authentication"
	"backend/internal/branch"
	"backend/internal/customer"
	"backend/internal/employee"
	"backend/internal/order"
	"backend/internal/stock"
	stocklog "backend/internal/stock-log"

	"gorm.io/gorm"
)

type Container struct {
	DB *gorm.DB

	// Handlers
	AuthHandler     *authentication.Handler
	CustomerHandler *customer.Handler
	OrderHandler    *order.Handler
	StockHandler    *stock.Handler
	StockLogHandler *stocklog.Handler
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

	// Initialize Stock Module
	stockRepository := stock.NewRepository(db)
	stockService := stock.NewService(stockRepository)
	stockHandler := stock.NewHandler(stockService)

	// Initialize Branch / Employee (repository อย่างเดียว ยังไม่มี service/handler)
	branchRepository := branch.NewRepository(db)
	employeeRepository := employee.NewRepository(db)
	// TODO: ส่งเข้า stocklog.NewService เมื่อเริ่มใช้เช็ค branchId / employeeId
	_, _ = branchRepository, employeeRepository

	// Initialize Stock Log Module
	stockLogRepository := stocklog.NewRepository(db)
	stockLogService := stocklog.NewService(stockLogRepository, stockRepository)
	stockLogHandler := stocklog.NewHandler(stockLogService)

	return &Container{
		DB:              db,
		AuthHandler:     authHandler,
		CustomerHandler: customerHandler,
		OrderHandler:    orderHandler,
		StockHandler:    stockHandler,
		StockLogHandler: stockLogHandler,
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
