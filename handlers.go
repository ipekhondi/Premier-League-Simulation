package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type LeagueResponse struct {
	Teams   []Team  `json:"teams"`
	Matches []Match `json:"matches"`
	Week    int     `json:"week"`
}

// handleGetLeagueTable returns the current league standings
func handleGetLeagueTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var teams []Team
	result := db.Order("points desc, goals_for - goals_against desc, goals_for desc").Find(&teams)
	if result.Error != nil {
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// handlePlayNextWeek simulates the next week's matches
func handlePlayNextWeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current week from the latest played match (where we have scores)
	var currentWeek int
	if err := db.Model(&Match{}).
		Where("home_goals IS NOT NULL").
		Select("COALESCE(MAX(week), 0)").
		Scan(&currentWeek).Error; err != nil {
		http.Error(w, "Error fetching current week", http.StatusInternalServerError)
		return
	}
	log.Printf("Initial currentWeek (last played week or 0): %d", currentWeek)

	// Get the maximum possible week number
	var maxWeek int
	if err := db.Model(&Match{}).Select("MAX(week)").Scan(&maxWeek).Error; err != nil {
		http.Error(w, "Error fetching max week", http.StatusInternalServerError)
		return
	}
	log.Printf("Max week fetched from DB: %d", maxWeek)

	currentWeek++ // Move to next week

	// If we've reached beyond the end of the season, return an error
	if currentWeek > maxWeek {
		http.Error(w, "Season is complete", http.StatusBadRequest)
		return
	}

	log.Printf("Processing week %d of %d", currentWeek, maxWeek)

	// Get current week's matches
	var weekMatches []Match
	// Preload is still useful if we want to access team names directly from weekMatches for some reason,
	// but for updates, we will fetch teams explicitly.
	if err := db.Preload("HomeTeam").Preload("AwayTeam").Where("week = ?", currentWeek).Find(&weekMatches).Error; err != nil {
		http.Error(w, "Error fetching matches", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d matches for week %d", len(weekMatches), currentWeek)

	// Simulate matches
	for i := range weekMatches {
		// It's safer to work with a pointer to the match from the slice
		// when we intend to modify and save it.
		currentMatch := &weekMatches[i]

		// Explicitly fetch home and away teams using their IDs from the currentMatch
		var homeTeamToUpdate Team
		if err := db.First(&homeTeamToUpdate, currentMatch.HomeTeamID).Error; err != nil {
			log.Printf("Error fetching home team ID %d for match %d: %v", currentMatch.HomeTeamID, currentMatch.ID, err)
			http.Error(w, "Error fetching home team for match simulation", http.StatusInternalServerError)
			return
		}

		var awayTeamToUpdate Team
		if err := db.First(&awayTeamToUpdate, currentMatch.AwayTeamID).Error; err != nil {
			log.Printf("Error fetching away team ID %d for match %d: %v", currentMatch.AwayTeamID, currentMatch.ID, err)
			http.Error(w, "Error fetching away team for match simulation", http.StatusInternalServerError)
			return
		}

		// Base goals on team strengths with some randomness
		rand.Seed(time.Now().UnixNano())                                                      // Consider seeding rand once globally if performance is an issue
		homeStrengthFactor := float64(homeTeamToUpdate.Strength) * (1.2 + rand.Float64()*0.3) // Home advantage
		awayStrengthFactor := float64(awayTeamToUpdate.Strength) * (1.0 + rand.Float64()*0.3)

		// Calculate scores as int first
		homeScore := int(homeStrengthFactor / 20.0)
		awayScore := int(awayStrengthFactor / 20.0)

		// Assign addresses of scores to the currentMatch fields
		currentMatch.HomeGoals = &homeScore
		currentMatch.AwayGoals = &awayScore

		// Log using the names from the explicitly fetched teams
		log.Printf("Match: %s vs %s, Score: %d-%d", homeTeamToUpdate.Name, awayTeamToUpdate.Name, homeScore, awayScore)

		// Update team statistics on the explicitly fetched team instances
		homeTeamToUpdate.UpdateStats(homeScore, awayScore)
		awayTeamToUpdate.UpdateStats(awayScore, homeScore)

		// Save match results (scores) to the currentMatch
		if err := db.Save(currentMatch).Error; err != nil {
			log.Printf("Error saving match ID %d scores: %v", currentMatch.ID, err)
			http.Error(w, "Error saving match scores", http.StatusInternalServerError)
			return
		}

		// Save updated team statistics for the explicitly fetched teams
		// GORM will use the ID on homeTeamToUpdate/awayTeamToUpdate to perform an UPDATE.
		if err := db.Save(&homeTeamToUpdate).Error; err != nil {
			log.Printf("Error saving updated home team ID %d stats: %v", homeTeamToUpdate.ID, err)
			http.Error(w, "Error saving home team stats", http.StatusInternalServerError)
			return
		}
		if err := db.Save(&awayTeamToUpdate).Error; err != nil {
			log.Printf("Error saving updated away team ID %d stats: %v", awayTeamToUpdate.ID, err)
			http.Error(w, "Error saving away team stats", http.StatusInternalServerError)
			return
		}
	}

	// Get updated teams for response
	var teams []Team
	if err := db.Find(&teams).Error; err != nil {
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	// Get all matches for response, including played and upcoming
	var allMatches []Match
	if err := db.Preload("HomeTeam").Preload("AwayTeam").Order("week asc").Find(&allMatches).Error; err != nil {
		http.Error(w, "Error fetching all matches", http.StatusInternalServerError)
		return
	}

	log.Printf("Total matches in database: %d", len(allMatches))

	response := LeagueResponse{
		Teams:   teams,
		Matches: allMatches,
		Week:    currentWeek,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInitializeLeague creates initial teams and generates fixtures
func handleInitializeLeague(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create initial teams with different strengths
	teams := []Team{
		{Name: "Chelsea", Strength: 85},
		{Name: "Arsenal", Strength: 82},
		{Name: "Manchester City", Strength: 88},
		{Name: "Liverpool", Strength: 84},
	}

	// Save teams to database
	for i := range teams {
		if err := db.Create(&teams[i]).Error; err != nil {
			http.Error(w, "Error creating teams", http.StatusInternalServerError)
			return
		}
	}
	log.Printf("Created %d teams", len(teams))

	// Generate fixtures
	league := NewLeague(teams)
	matches := league.GenerateFixtures()
	log.Printf("Generated %d fixtures by league.GenerateFixtures()", len(matches))
	if len(matches) > 0 {
		// Log details of the first match as an example, ensure Week field exists and is accessible
		log.Printf("Example first generated match: Week %d, HomeTeamID %d, AwayTeamID %d", matches[0].Week, matches[0].HomeTeamID, matches[0].AwayTeamID)
	}

	// Save matches to database
	for _, match := range matches {
		if err := db.Create(&match).Error; err != nil {
			http.Error(w, "Error creating matches", http.StatusInternalServerError)
			return
		}
	}

	response := LeagueResponse{
		Teams:   teams,
		Matches: matches,
		Week:    1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
