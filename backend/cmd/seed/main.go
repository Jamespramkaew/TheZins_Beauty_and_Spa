package main

import (
	"flag"
	"log"

	"backend/config"
	"backend/database"

	"gorm.io/gorm"
)

func main() {
	// Flags
	reset := flag.Bool("reset", false, "Reset database before seeding")
	flag.Parse()

	log.SetPrefix("[seed] ")
	log.SetFlags(log.Ldate | log.Ltime)

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database")

	// Reset database if flag is set
	if *reset {
		log.Println("Resetting database...")
		if err := database.ResetDatabase(db); err != nil {
			log.Fatalf("Failed to reset database: %v", err)
		}
		log.Println("Database reset completed")
	} else {
		// Run migrations
		if err := database.RunMigration(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed")
	}

	// Seed data
	log.Println("Starting to seed data...")

	if err := database.SeedBranches(db); err != nil {
		log.Fatalf("Failed to seed branches: %v", err)
	}
	log.Println("✅ Seeded branches")

	if err := database.SeedUsers(db); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}
	log.Println("✅ Seeded users")

	if err := database.SeedCourses(db); err != nil {
		log.Fatalf("Failed to seed courses: %v", err)
	}
	log.Println("✅ Seeded courses")

	if err := database.SeedStocks(db); err != nil {
		log.Fatalf("Failed to seed stocks: %v", err)
	}
	log.Println("✅ Seeded stocks")

	if err := seedExamples(db); err != nil {
		log.Fatalf("Failed to seed examples: %v", err)
	}

	log.Println("✅ Seeding completed successfully!")
}

func seedExamples(db *gorm.DB) error {
	examples := []database.Example{
		{Message: "Welcome to The ZINs Beauty and Spa"},
		{Message: "Book your appointment today"},
		{Message: "We offer premium spa services"},
		{Message: "Relax and rejuvenate with us"},
		{Message: "Professional beauty treatments"},
		{Message: "Your wellness is our priority"},
		{Message: "Experience luxury spa services"},
		{Message: "Expert beauticians at your service"},
		{Message: "Transform your beauty routine"},
		{Message: "Indulge in our signature treatments"},
	}

	log.Println("Seeding examples...")

	for _, example := range examples {
		if err := db.Create(&example).Error; err != nil {
			log.Printf("Warning: Failed to create example '%s': %v", example.Message, err)
			continue
		}
	}

	log.Printf("✅ Seeded %d examples", len(examples))
	return nil
}
