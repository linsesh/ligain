package api

import (
	"testing"
)

func TestToMatch_BookmakerPreference(t *testing.T) {
	tests := []struct {
		name         string
		fixture      sportmonksFixture
		expectedHome float64
		expectedDraw float64
		expectedAway float64
		description  string
	}{
		{
			name: "Unibet preferred over Bet365",
			fixture: sportmonksFixture{
				ID: 12345,
				Participants: []participant{
					{Name: "Home Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "home"}},
					{Name: "Away Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "away"}},
				},
				Odds: []odd{
					{BookmakerID: 2, MarketID: 1, Label: "Home", Value: "2.0"},  // Bet365
					{BookmakerID: 2, MarketID: 1, Label: "Draw", Value: "3.0"},  // Bet365
					{BookmakerID: 2, MarketID: 1, Label: "Away", Value: "4.0"},  // Bet365
					{BookmakerID: 23, MarketID: 1, Label: "Home", Value: "1.8"}, // Unibet
					{BookmakerID: 23, MarketID: 1, Label: "Draw", Value: "3.2"}, // Unibet
					{BookmakerID: 23, MarketID: 1, Label: "Away", Value: "4.2"}, // Unibet
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    1,
				HasOdds:    true,
			},
			expectedHome: 1.8,
			expectedDraw: 3.2,
			expectedAway: 4.2,
			description:  "Should prefer Unibet (ID 23) over Bet365 (ID 2)",
		},
		{
			name: "Bet365 used when Unibet not available",
			fixture: sportmonksFixture{
				ID: 12346,
				Participants: []participant{
					{Name: "Home Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "home"}},
					{Name: "Away Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "away"}},
				},
				Odds: []odd{
					{BookmakerID: 2, MarketID: 1, Label: "Home", Value: "2.0"}, // Bet365
					{BookmakerID: 2, MarketID: 1, Label: "Draw", Value: "3.0"}, // Bet365
					{BookmakerID: 2, MarketID: 1, Label: "Away", Value: "4.0"}, // Bet365
					{BookmakerID: 5, MarketID: 1, Label: "Home", Value: "1.9"}, // Other bookmaker
					{BookmakerID: 5, MarketID: 1, Label: "Draw", Value: "3.1"}, // Other bookmaker
					{BookmakerID: 5, MarketID: 1, Label: "Away", Value: "4.1"}, // Other bookmaker
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    1,
				HasOdds:    true,
			},
			expectedHome: 2.0,
			expectedDraw: 3.0,
			expectedAway: 4.0,
			description:  "Should use Bet365 (ID 2) when Unibet not available",
		},
		{
			name: "Fallback to any available bookmaker",
			fixture: sportmonksFixture{
				ID: 12347,
				Participants: []participant{
					{Name: "Home Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "home"}},
					{Name: "Away Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "away"}},
				},
				Odds: []odd{
					{BookmakerID: 5, MarketID: 1, Label: "Home", Value: "1.9"}, // Other bookmaker
					{BookmakerID: 5, MarketID: 1, Label: "Draw", Value: "3.1"}, // Other bookmaker
					{BookmakerID: 5, MarketID: 1, Label: "Away", Value: "4.1"}, // Other bookmaker
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    1,
				HasOdds:    true,
			},
			expectedHome: 1.9,
			expectedDraw: 3.1,
			expectedAway: 4.1,
			description:  "Should fallback to any available bookmaker when preferred ones not available",
		},
		{
			name: "No odds available",
			fixture: sportmonksFixture{
				ID: 12348,
				Participants: []participant{
					{Name: "Home Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "home"}},
					{Name: "Away Team", Meta: struct {
						Location string `json:"location"`
						Winner   *bool  `json:"winner"`
						Position int    `json:"position"`
					}{Location: "away"}},
				},
				Odds:       []odd{},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    1,
				HasOdds:    true,
			},
			expectedHome: 0.0,
			expectedDraw: 0.0,
			expectedAway: 0.0,
			description:  "Should return zero odds when no odds available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := tt.fixture.toMatch()
			if err != nil {
				t.Fatalf("toMatch() error = %v", err)
			}

			homeOdds := match.GetHomeTeamOdds()
			drawOdds := match.GetDrawOdds()
			awayOdds := match.GetAwayTeamOdds()

			if homeOdds != tt.expectedHome {
				t.Errorf("Home odds = %v, want %v", homeOdds, tt.expectedHome)
			}
			if drawOdds != tt.expectedDraw {
				t.Errorf("Draw odds = %v, want %v", drawOdds, tt.expectedDraw)
			}
			if awayOdds != tt.expectedAway {
				t.Errorf("Away odds = %v, want %v", awayOdds, tt.expectedAway)
			}

			t.Logf("Test: %s - %s", tt.name, tt.description)
			t.Logf("Result: Home=%.2f, Draw=%.2f, Away=%.2f", homeOdds, drawOdds, awayOdds)
		})
	}
}

func TestToMatch_CompleteFixture(t *testing.T) {
	f := sportmonksFixture{
		ID: 12345,
		Participants: []participant{
			{Name: "Home Team", Meta: struct {
				Location string `json:"location"`
				Winner   *bool  `json:"winner"`
				Position int    `json:"position"`
			}{Location: "home"}},
			{Name: "Away Team", Meta: struct {
				Location string `json:"location"`
				Winner   *bool  `json:"winner"`
				Position int    `json:"position"`
			}{Location: "away"}},
		},
		Scores: []score{
			{Type: "FT", HomeScore: 2, AwayScore: 1},
		},
		Odds: []odd{
			{BookmakerID: 23, MarketID: 1, Label: "Home", Value: "1.8"},
			{BookmakerID: 23, MarketID: 1, Label: "Draw", Value: "3.2"},
			{BookmakerID: 23, MarketID: 1, Label: "Away", Value: "4.2"},
		},
		StartingAt: "2025-01-01 15:00:00",
		Season:     season{Name: "2024/2025"},
		League:     league{Name: "Ligue 1"},
		Round:      round{Name: "1"},
		StateID:    5, // Finished match
		HasOdds:    true,
	}

	match, err := f.toMatch()
	if err != nil {
		t.Fatalf("toMatch() error = %v", err)
	}

	// Test basic match properties
	if match.Id() != "Ligue 1-2024/2025-Home Team-Away Team-1" {
		t.Errorf("Expected match ID 'Ligue 1-2024/2025-Home Team-Away Team-1', got '%s'", match.Id())
	}

	if match.GetHomeTeam() != "Home Team" {
		t.Errorf("Expected home team 'Home Team', got '%s'", match.GetHomeTeam())
	}

	if match.GetAwayTeam() != "Away Team" {
		t.Errorf("Expected away team 'Away Team', got '%s'", match.GetAwayTeam())
	}

	// Test odds
	if match.GetHomeTeamOdds() != 1.8 {
		t.Errorf("Expected home odds 1.8, got %f", match.GetHomeTeamOdds())
	}

	if match.GetDrawOdds() != 3.2 {
		t.Errorf("Expected draw odds 3.2, got %f", match.GetDrawOdds())
	}

	if match.GetAwayTeamOdds() != 4.2 {
		t.Errorf("Expected away odds 4.2, got %f", match.GetAwayTeamOdds())
	}
}
