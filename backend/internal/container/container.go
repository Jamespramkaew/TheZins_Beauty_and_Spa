package container

// For Dependency Injection

import (
	"backend/config"
	"backend/internal/example"

	"gorm.io/gorm"
)

type Container struct {
	DB *gorm.DB

	// Handlers
	ExampleHandler *example.Handler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Initialize Example Module
	exampleRepository := example.NewRepository(db)
	exampleService := example.NewService(exampleRepository)
	exampleHandler := example.NewHandler(exampleService)

	return &Container{
		DB:             db,
		ExampleHandler: exampleHandler,
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