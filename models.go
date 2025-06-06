package main

//This is the database model for the teams in the league and the matches playes between them.
type Team struct {
	ID           uint `gorm:"primaryKey"`
	Name         string
	Strength     int
	Wins         int
	Draws        int
	Losses       int
	GoalsFor     int
	GoalsAgainst int
	Points       int
}

// UpdateStats updates the team's statistics after a match
func (t *Team) UpdateStats(goalsFor, goalsAgainst int) {
	t.GoalsFor += goalsFor
	t.GoalsAgainst += goalsAgainst

	if goalsFor > goalsAgainst {
		t.Wins++
		t.Points += 3
	} else if goalsFor == goalsAgainst {
		t.Draws++
		t.Points++
	} else {
		t.Losses++
	}
}

type Match struct {
	ID         uint `gorm:"primaryKey"`
	Week       int
	HomeTeamID uint
	AwayTeamID uint
	HomeGoals  *int
	AwayGoals  *int
	HomeTeam   Team `gorm:"foreignKey:HomeTeamID"`
	AwayTeam   Team `gorm:"foreignKey:AwayTeamID"`
}
