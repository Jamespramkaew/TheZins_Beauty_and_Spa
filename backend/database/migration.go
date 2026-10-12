package database

import (
	"log"
	"gorm.io/gorm"
)

func RunMigration(db *gorm.DB) error {
	log.Println("Running database migration..")

	err := db.AutoMigrate(
		&Example{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func DropAllTables(db *gorm.DB) error {
	log.Println("Dropping all tables...")

	err := db.Migrator().DropTable(
		&Example{},
	)

	if err != nil {
		return err
	}

	log.Println("All tables dropped successfully")
	return nil
}

func ResetDatabase(db *gorm.DB) error {
	if err := DropAllTables(db); err != nil {
		return err
	}
	return RunMigration(db)
}