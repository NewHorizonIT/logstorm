package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/NewHorizonIT/logstorm/internal/config"
)

func Connect(cnf *config.DatabaseConfig) *gorm.DB {
	// Build the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cnf.Host, cnf.Port, cnf.User, cnf.Password, cnf.Name)

	// Open connect to the database using gorm
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Ping database to ensure connection is established
	if err := PingDB(db); err != nil {
		log.Fatalf("Database connection is not healthy: %v", err)
	}

	log.Println("Successfully connected to the database")
	return db
}

// CloseDB closes the database connection
func CloseDB(db *gorm.DB) error {

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return err
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
		return err
	}
	log.Println("Database connection closed successfully")
	return nil
}

// PingDB checks the database connection
func PingDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Printf("Failed to ping database: %v", err)
		return err
	}
	log.Println("Database connection is healthy")
	return nil
}

// Set Connection Pooling parameters
func SetConnectionPool(db *sql.DB, maxOpenConns, maxIdleConns int) {
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	log.Printf("Set connection pool: MaxOpenConns=%d, MaxIdleConns=%d", maxOpenConns, maxIdleConns)
}
