package container

// For Dependency Injection

import (
	"backend/config"

	"gorm.io/gorm"
)

type Container struct {
	DB *gorm.DB

	// Handlers
	// TODO: Add handlers here
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	return &Container{
		DB: db,
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