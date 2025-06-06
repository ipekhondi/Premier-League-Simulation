package main

import (
	"math/rand"
	"time"
)

// League represents the football league simulation
type League struct {
	Teams       []Team
	Matches     []Match
	CurrentWeek int
}

// NewLeague creates a new league with the given teams
func NewLeague(teams []Team) *League {
	return &League{
		Teams:       teams,
		Matches:     make([]Match, 0),
		CurrentWeek: 1,
	}
}

// SimulateMatch simulates a match between two teams based on their strengths
func (l *League) SimulateMatch(homeTeam, awayTeam *Team) Match {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Base goals on team strengths with some randomness
	homeStrengthFactor := float64(homeTeam.Strength) * (1.2 + rand.Float64()*0.3) // Home advantage
	awayStrengthFactor := float64(awayTeam.Strength) * (1.0 + rand.Float64()*0.3)

	homeGoals := int(homeStrengthFactor / 20.0)
	awayGoals := int(awayStrengthFactor / 20.0)

	match := Match{
		Week:       l.CurrentWeek,
		HomeTeamID: homeTeam.ID,
		AwayTeamID: awayTeam.ID,
		HomeGoals:  &homeGoals,
		AwayGoals:  &awayGoals,
		HomeTeam:   *homeTeam,
		AwayTeam:   *awayTeam,
	}

	// Update team statistics
	homeTeam.UpdateStats(homeGoals, awayGoals)
	awayTeam.UpdateStats(awayGoals, homeGoals)

	return match
}

// GenerateFixtures generates all matches for the season
func (l *League) GenerateFixtures() []Match {
	matches := make([]Match, 0)
	numTeamsOriginal := len(l.Teams) // Use a different variable for the original team count
	if numTeamsOriginal < 2 {
		return matches
	}

	// Create a temporary slice for teams to handle potential BYE team
	tempTeams := make([]Team, numTeamsOriginal)
	copy(tempTeams, l.Teams)

	numTeamsForScheduling := numTeamsOriginal
	// If odd number of teams, add a dummy "bye" team for scheduling logic
	if numTeamsOriginal%2 != 0 {
		tempTeams = append(tempTeams, Team{ID: 0, Name: "BYE"}) // ID 0 or a special marker for bye
		numTeamsForScheduling++
	}

	rounds := numTeamsForScheduling - 1 // Number of weeks in one half of the season

	for round := 0; round < rounds; round++ {
		currentWeekInRound := round + 1
		for i := 0; i < numTeamsForScheduling/2; i++ {
			homeIdx := i
			awayIdx := numTeamsForScheduling - 1 - i

			homeTeam := tempTeams[homeIdx]
			awayTeam := tempTeams[awayIdx]

			// Skip if one of the teams is the BYE team (if it was added)
			if homeTeam.Name == "BYE" || awayTeam.Name == "BYE" {
				continue
			}

			// First half fixture
			matches = append(matches, Match{
				Week:       currentWeekInRound,
				HomeTeamID: homeTeam.ID,
				AwayTeamID: awayTeam.ID,
				HomeGoals:  nil, // Scores are nil for unplayed matches
				AwayGoals:  nil,
			})

			// Second half fixture (return match)
			// Ensure homeTeam.ID and awayTeam.ID are not zero if they were not BYE
			if homeTeam.ID != 0 && awayTeam.ID != 0 { // Redundant if BYE check above is sufficient
				matches = append(matches, Match{
					Week:       currentWeekInRound + rounds, // Week in the second half
					HomeTeamID: awayTeam.ID,                 // Swap home and away
					AwayTeamID: homeTeam.ID,
					HomeGoals:  nil,
					AwayGoals:  nil,
				})
			}
		}

		// Rotate teams for the next round (circling algorithm)
		// Keep the first team (index 0) fixed, rotate the others.
		if numTeamsForScheduling > 2 { // Rotation makes sense for 3+ teams
			lastTeamInRotation := tempTeams[numTeamsForScheduling-1]
			for k := numTeamsForScheduling - 1; k > 1; k-- {
				tempTeams[k] = tempTeams[k-1]
			}
			tempTeams[1] = lastTeamInRotation
		}
	}
	return matches
}

// GetLeagueTable returns the teams sorted by points
func (l *League) GetLeagueTable() []Team {
	// Create a copy of teams to sort
	table := make([]Team, len(l.Teams))
	copy(table, l.Teams)

	// Sort teams by points (descending), then goal difference, then goals scored
	for i := 0; i < len(table)-1; i++ {
		for j := i + 1; j < len(table); j++ {
			teamA := table[i]
			teamB := table[j]

			// Compare points
			if teamB.Points > teamA.Points {
				table[i], table[j] = table[j], table[i]
				continue
			}

			// If points are equal, compare goal difference
			goalDiffA := teamA.GoalsFor - teamA.GoalsAgainst
			goalDiffB := teamB.GoalsFor - teamB.GoalsAgainst
			if teamB.Points == teamA.Points && goalDiffB > goalDiffA {
				table[i], table[j] = table[j], table[i]
				continue
			}

			// If goal difference is equal, compare goals scored
			if teamB.Points == teamA.Points && goalDiffB == goalDiffA && teamB.GoalsFor > teamA.GoalsFor {
				table[i], table[j] = table[j], table[i]
			}
		}
	}

	return table
}
