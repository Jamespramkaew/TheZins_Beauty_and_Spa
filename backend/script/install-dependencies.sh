#!/bin/bash

echo "Installing Go dependencies..."

# Environment variables
go get github.com/joho/godotenv

# Fiber v3
go get github.com/gofiber/fiber/v3
go get github.com/gofiber/fiber/v3/middleware/cors
go get github.com/gofiber/fiber/v3/middleware/logger
go get github.com/gofiber/fiber/v3/middleware/recover

# Validation
go get github.com/go-playground/validator/v10

# Auth Service
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt

# Security
go get github.com/gofiber/fiber/v3/middleware/helmet
go get github.com/gofiber/fiber/v3/middleware/limiter

# GORM
go get gorm.io/gorm
go get gorm.io/driver/postgres

# Testing
go get github.com/stretchr/testify/assert

# Clean up and download
go mod tidy
go mod download

echo "✅ All dependencies installed successfully!"
