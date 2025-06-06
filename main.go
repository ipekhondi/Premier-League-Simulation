package main

import (
	"log"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
) // Import neccessary packages for GORM and SQLite

var db *gorm.DB

// db is a global variable that hold the database connection

// InitDB initializes the database connection and migrates the schema
// It creates the database file if it does not exist and sets up the tables for Team and Match models.
// This function should be called at the start of the application to ensure the database is ready for use.
func InitDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("data/league.db"), &gorm.Config{}) // SQLite database connection
	if err != nil {
		log.Fatal("Failed to connect database:", err) // Fatal log if connection fails
	}

	err = db.AutoMigrate(&Team{}, &Match{}) // Automatically migrate the schema, creating tables if they do not exist
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
		// Controls the database schema migration again if it fails
	}
}

func main() {
	// Initialize database
	InitDB()

	// Set up HTTP routes
	http.HandleFunc("/api/league/table", handleGetLeagueTable)
	http.HandleFunc("/api/league/play-next", handlePlayNextWeek)
	http.HandleFunc("/api/league/initialize", handleInitializeLeague)

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
