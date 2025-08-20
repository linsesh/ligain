package api

import (
	"encoding/json"
	"ligain/backend/models"
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
			{
				ID:            1,
				FixtureID:     12345,
				TypeID:        1525,
				ParticipantID: 1,
				Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{
					Goals:       2,
					Participant: "home",
				},
				Description: "CURRENT",
			},
			{
				ID:            2,
				FixtureID:     12345,
				TypeID:        1525,
				ParticipantID: 2,
				Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{
					Goals:       1,
					Participant: "away",
				},
				Description: "CURRENT",
			},
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

func TestToMatch_StateTransitionsAndScores(t *testing.T) {
	tests := []struct {
		name               string
		jsonResponse       string
		expectedStatus     models.MatchStatus
		expectedHome       int
		expectedAway       int
		expectedFinished   bool
		expectedInProgress bool
		description        string
	}{
		{
			name: "Scheduled match with no scores (Monaco vs Le Havre)",
			jsonResponse: `{
				"id": 19433440,
				"state_id": 1,
				"name": "Monaco vs Le Havre",
				"starting_at": "2025-08-16 17:00:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [],
				"participants": [
					{"name": "Monaco", "meta": {"location": "home"}},
					{"name": "Le Havre", "meta": {"location": "away"}}
				],
				"odds": []
			}`,
			expectedStatus:     models.MatchStatusScheduled,
			expectedHome:       0,
			expectedAway:       0,
			expectedFinished:   false,
			expectedInProgress: false,
			description:        "Should handle scheduled match with no scores",
		},
		{
			name: "Scheduled match (Lens vs Lyon)",
			jsonResponse: `{
				"id": 19433445,
				"state_id": 1,
				"name": "Lens vs Olympique Lyonnais",
				"starting_at": "2025-08-16 15:00:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [],
				"participants": [
					{"name": "Lens", "meta": {"location": "home"}},
					{"name": "Olympique Lyonnais", "meta": {"location": "away"}}
				],
				"odds": []
			}`,
			expectedStatus:     models.MatchStatusScheduled,
			expectedHome:       0,
			expectedAway:       0,
			expectedFinished:   false,
			expectedInProgress: false,
			description:        "Should handle scheduled match with no scores",
		},
		{
			name: "Match in progress (1st half)",
			jsonResponse: `{
				"id": 19433446,
				"state_id": 2,
				"name": "Brest vs LOSC Lille",
				"starting_at": "2025-08-17 13:00:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 1, "fixture_id": 19433446, "type_id": 1525, "participant_id": 1, "score": {"goals": 1, "participant": "home"}, "description": "CURRENT"},
					{"id": 2, "fixture_id": 19433446, "type_id": 1525, "participant_id": 2, "score": {"goals": 0, "participant": "away"}, "description": "CURRENT"}
				],
				"participants": [
					{"name": "Brest", "meta": {"location": "home"}},
					{"name": "LOSC Lille", "meta": {"location": "away"}}
				],
				"odds": []
			}`,
			expectedStatus:     models.MatchStatusStarted,
			expectedHome:       1,
			expectedAway:       0,
			expectedFinished:   false,
			expectedInProgress: true,
			description:        "Should handle match in progress with current scores",
		},
		{
			name: "Match in progress (2nd half)",
			jsonResponse: `{
				"id": 19433444,
				"state_id": 4,
				"name": "Nice vs Toulouse",
				"starting_at": "2025-08-16 19:05:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 1, "fixture_id": 19433444, "type_id": 1525, "participant_id": 1, "score": {"goals": 2, "participant": "home"}, "description": "CURRENT"},
					{"id": 2, "fixture_id": 19433444, "type_id": 1525, "participant_id": 2, "score": {"goals": 1, "participant": "away"}, "description": "CURRENT"}
				],
				"participants": [
					{"name": "Nice", "meta": {"location": "home"}},
					{"name": "Toulouse", "meta": {"location": "away"}}
				],
				"odds": [
					{"bookmaker_id": 23, "market_id": 1, "label": "Home", "value": "2.14"},
					{"bookmaker_id": 23, "market_id": 1, "label": "Draw", "value": "3.5"},
					{"bookmaker_id": 23, "market_id": 1, "label": "Away", "value": "3.35"}
				]
			}`,
			expectedStatus:     models.MatchStatusStarted,
			expectedHome:       2,
			expectedAway:       1,
			expectedFinished:   false,
			expectedInProgress: true,
			description:        "Should handle match in progress (2nd half) with current scores",
		},
		{
			name: "Finished match (Rennes vs Marseille)",
			jsonResponse: `{
				"id": 19433447,
				"state_id": 5,
				"name": "Rennes vs Olympique Marseille",
				"starting_at": "2025-08-15 18:45:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 16782918, "fixture_id": 19433447, "type_id": 1, "participant_id": 44, "score": {"goals": 0, "participant": "away"}, "description": "1ST_HALF"},
					{"id": 16783671, "fixture_id": 19433447, "type_id": 48996, "participant_id": 44, "score": {"goals": 0, "participant": "away"}, "description": "2ND_HALF_ONLY"},
					{"id": 16782917, "fixture_id": 19433447, "type_id": 1525, "participant_id": 598, "score": {"goals": 1, "participant": "home"}, "description": "CURRENT"},
					{"id": 16782916, "fixture_id": 19433447, "type_id": 1525, "participant_id": 44, "score": {"goals": 0, "participant": "away"}, "description": "CURRENT"},
					{"id": 16782921, "fixture_id": 19433447, "type_id": 2, "participant_id": 598, "score": {"goals": 1, "participant": "home"}, "description": "2ND_HALF"},
					{"id": 16782920, "fixture_id": 19433447, "type_id": 2, "participant_id": 44, "score": {"goals": 0, "participant": "away"}, "description": "2ND_HALF"},
					{"id": 16783670, "fixture_id": 19433447, "type_id": 48996, "participant_id": 598, "score": {"goals": 1, "participant": "home"}, "description": "2ND_HALF_ONLY"},
					{"id": 16782919, "fixture_id": 19433447, "type_id": 1, "participant_id": 598, "score": {"goals": 0, "participant": "home"}, "description": "1ST_HALF"}
				],
				"participants": [
					{"name": "Rennes", "meta": {"location": "home", "winner": true}},
					{"name": "Olympique Marseille", "meta": {"location": "away", "winner": false}}
				],
				"odds": []
			}`,
			expectedStatus:     models.MatchStatusFinished,
			expectedHome:       1,
			expectedAway:       0,
			expectedFinished:   true,
			expectedInProgress: false,
			description:        "Should handle finished match with correct final score (Rennes 1-0 Marseille)",
		},
		{
			name: "Finished match with multiple score types",
			jsonResponse: `{
				"id": 19433446,
				"state_id": 5,
				"name": "Brest vs LOSC Lille",
				"starting_at": "2025-08-17 13:00:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 1, "fixture_id": 19433446, "type_id": 1, "participant_id": 1, "score": {"goals": 0, "participant": "home"}, "description": "1ST_HALF"},
					{"id": 2, "fixture_id": 19433446, "type_id": 1, "participant_id": 2, "score": {"goals": 0, "participant": "away"}, "description": "1ST_HALF"},
					{"id": 3, "fixture_id": 19433446, "type_id": 2, "participant_id": 1, "score": {"goals": 2, "participant": "home"}, "description": "2ND_HALF"},
					{"id": 4, "fixture_id": 19433446, "type_id": 2, "participant_id": 2, "score": {"goals": 1, "participant": "away"}, "description": "2ND_HALF"},
					{"id": 5, "fixture_id": 19433446, "type_id": 1525, "participant_id": 1, "score": {"goals": 2, "participant": "home"}, "description": "CURRENT"},
					{"id": 6, "fixture_id": 19433446, "type_id": 1525, "participant_id": 2, "score": {"goals": 1, "participant": "away"}, "description": "CURRENT"}
				],
				"participants": [
					{"name": "Brest", "meta": {"location": "home", "winner": true}},
					{"name": "LOSC Lille", "meta": {"location": "away", "winner": false}}
				],
				"odds": [
					{"bookmaker_id": 23, "market_id": 1, "label": "Home", "value": "3.35"},
					{"bookmaker_id": 23, "market_id": 1, "label": "Draw", "value": "2.2"},
					{"bookmaker_id": 23, "market_id": 1, "label": "Away", "value": "3.4"}
				]
			}`,
			expectedStatus:     models.MatchStatusFinished,
			expectedHome:       2,
			expectedAway:       1,
			expectedFinished:   true,
			expectedInProgress: false,
			description:        "Should extract CURRENT scores even when multiple score types are present",
		},
		{
			name: "Match with no CURRENT scores (fallback to scheduled)",
			jsonResponse: `{
				"id": 19433441,
				"state_id": 1,
				"name": "Auxerre vs Lorient",
				"starting_at": "2025-08-17 15:15:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 1, "fixture_id": 19433441, "type_id": 1, "participant_id": 1, "score": {"goals": 0, "participant": "home"}, "description": "1ST_HALF"},
					{"id": 2, "fixture_id": 19433441, "type_id": 1, "participant_id": 2, "score": {"goals": 0, "participant": "away"}, "description": "1ST_HALF"}
				],
				"participants": [
					{"name": "Auxerre", "meta": {"location": "home"}},
					{"name": "Lorient", "meta": {"location": "away"}}
				],
				"odds": [
					{"bookmaker_id": 23, "market_id": 1, "label": "Home", "value": "2.3"},
					{"bookmaker_id": 23, "market_id": 1, "label": "Draw", "value": "3.05"},
					{"bookmaker_id": 23, "market_id": 1, "label": "Away", "value": "3.45"}
				]
			}`,
			expectedStatus:     models.MatchStatusScheduled,
			expectedHome:       0,
			expectedAway:       0,
			expectedFinished:   false,
			expectedInProgress: false,
			description:        "Should handle match with no CURRENT scores as scheduled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the JSON response
			var fixture sportmonksFixture
			if err := json.Unmarshal([]byte(tt.jsonResponse), &fixture); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			// Convert to match
			match, err := fixture.toMatch()
			if err != nil {
				t.Fatalf("toMatch() error = %v", err)
			}

			// Test status
			if match.GetStatus() != tt.expectedStatus {
				t.Errorf("Status = %s, want %s", match.GetStatus(), tt.expectedStatus)
			}

			// Test scores
			if match.GetHomeGoals() != tt.expectedHome {
				t.Errorf("Home goals = %d, want %d", match.GetHomeGoals(), tt.expectedHome)
			}
			if match.GetAwayGoals() != tt.expectedAway {
				t.Errorf("Away goals = %d, want %d", match.GetAwayGoals(), tt.expectedAway)
			}

			// Test match state
			if match.IsFinished() != tt.expectedFinished {
				t.Errorf("IsFinished() = %t, want %t", match.IsFinished(), tt.expectedFinished)
			}
			if match.IsInProgress() != tt.expectedInProgress {
				t.Errorf("IsInProgress() = %t, want %t", match.IsInProgress(), tt.expectedInProgress)
			}

			// Test basic properties
			if match.GetHomeTeam() == "" {
				t.Error("Home team should not be empty")
			}
			if match.GetAwayTeam() == "" {
				t.Error("Away team should not be empty")
			}
			if match.Id() == "" {
				t.Error("Match ID should not be empty")
			}

			t.Logf("Test: %s - %s", tt.name, tt.description)
			t.Logf("Result: %s vs %s (%d-%d), Status: %s, Finished: %t, InProgress: %t",
				match.GetHomeTeam(), match.GetAwayTeam(),
				match.GetHomeGoals(), match.GetAwayGoals(),
				match.GetStatus(), match.IsFinished(), match.IsInProgress())
		})
	}
}

func TestToMatch_ScoreExtractionEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		jsonResponse string
		expectedHome int
		expectedAway int
		description  string
	}{
		{
			name: "Empty scores array",
			jsonResponse: `{
				"id": 19433440,
				"state_id": 1,
				"name": "Monaco vs Le Havre",
				"starting_at": "2025-08-16 17:00:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [],
				"participants": [
					{"name": "Monaco", "meta": {"location": "home"}},
					{"name": "Le Havre", "meta": {"location": "away"}}
				],
				"odds": []
			}`,
			expectedHome: 0,
			expectedAway: 0,
			description:  "Should handle empty scores array",
		},
		{
			name: "Scores with non-CURRENT descriptions",
			jsonResponse: `{
				"id": 19433441,
				"state_id": 1,
				"name": "Auxerre vs Lorient",
				"starting_at": "2025-08-17 15:15:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 1, "fixture_id": 19433441, "type_id": 1, "participant_id": 1, "score": {"goals": 1, "participant": "home"}, "description": "1ST_HALF"},
					{"id": 2, "fixture_id": 19433441, "type_id": 1, "participant_id": 2, "score": {"goals": 0, "participant": "away"}, "description": "1ST_HALF"},
					{"id": 3, "fixture_id": 19433441, "type_id": 2, "participant_id": 1, "score": {"goals": 2, "participant": "home"}, "description": "2ND_HALF"},
					{"id": 4, "fixture_id": 19433441, "type_id": 2, "participant_id": 2, "score": {"goals": 1, "participant": "away"}, "description": "2ND_HALF"}
				],
				"participants": [
					{"name": "Auxerre", "meta": {"location": "home"}},
					{"name": "Lorient", "meta": {"location": "away"}}
				],
				"odds": []
			}`,
			expectedHome: 0,
			expectedAway: 0,
			description:  "Should ignore scores without CURRENT description",
		},
		{
			name: "Only home team CURRENT score",
			jsonResponse: `{
				"id": 19433442,
				"state_id": 2,
				"name": "Metz vs Strasbourg",
				"starting_at": "2025-08-17 15:15:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 1, "fixture_id": 19433442, "type_id": 1525, "participant_id": 1, "score": {"goals": 1, "participant": "home"}, "description": "CURRENT"}
				],
				"participants": [
					{"name": "Metz", "meta": {"location": "home"}},
					{"name": "Strasbourg", "meta": {"location": "away"}}
				],
				"odds": []
			}`,
			expectedHome: 1,
			expectedAway: 0,
			description:  "Should handle partial CURRENT scores (only home team)",
		},
		{
			name: "Only away team CURRENT score",
			jsonResponse: `{
				"id": 19433443,
				"state_id": 2,
				"name": "Nantes vs Paris Saint Germain",
				"starting_at": "2025-08-17 18:45:00",
				"has_odds": true,
				"league": {"name": "Ligue 1"},
				"season": {"name": "2025/2026"},
				"round": {"name": "1"},
				"scores": [
					{"id": 1, "fixture_id": 19433443, "type_id": 1525, "participant_id": 2, "score": {"goals": 2, "participant": "away"}, "description": "CURRENT"}
				],
				"participants": [
					{"name": "Nantes", "meta": {"location": "home"}},
					{"name": "Paris Saint Germain", "meta": {"location": "away"}}
				],
				"odds": []
			}`,
			expectedHome: 0,
			expectedAway: 2,
			description:  "Should handle partial CURRENT scores (only away team)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the JSON response
			var fixture sportmonksFixture
			if err := json.Unmarshal([]byte(tt.jsonResponse), &fixture); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			// Convert to match
			match, err := fixture.toMatch()
			if err != nil {
				t.Fatalf("toMatch() error = %v", err)
			}

			// Test scores
			if match.GetHomeGoals() != tt.expectedHome {
				t.Errorf("Home goals = %d, want %d", match.GetHomeGoals(), tt.expectedHome)
			}
			if match.GetAwayGoals() != tt.expectedAway {
				t.Errorf("Away goals = %d, want %d", match.GetAwayGoals(), tt.expectedAway)
			}

			t.Logf("Test: %s - %s", tt.name, tt.description)
			t.Logf("Result: %s vs %s (%d-%d)",
				match.GetHomeTeam(), match.GetAwayTeam(),
				match.GetHomeGoals(), match.GetAwayGoals())
		})
	}
}

// TestBug_ScoreExtraction_ZeroZeroDraw tests for the bug where matches appear as 0-0 draws
// This test checks if the score extraction logic is working correctly
func TestBug_ScoreExtraction_ZeroZeroDraw(t *testing.T) {
	tests := []struct {
		name         string
		fixture      sportmonksFixture
		expectedHome int
		expectedAway int
		description  string
	}{
		{
			name: "Finished match with CURRENT scores",
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
				Scores: []score{
					{
						ID:            1,
						FixtureID:     12345,
						TypeID:        1525,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       2,
							Participant: "home",
						},
						Description: "CURRENT",
					},
					{
						ID:            2,
						FixtureID:     12345,
						TypeID:        1525,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       1,
							Participant: "away",
						},
						Description: "CURRENT",
					},
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    5, // Finished match
				HasOdds:    true,
			},
			expectedHome: 2,
			expectedAway: 1,
			description:  "Should extract CURRENT scores correctly for finished match",
		},
		{
			name: "Finished match with no CURRENT scores (BUG SCENARIO)",
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
				Scores: []score{
					{
						ID:            1,
						FixtureID:     12346,
						TypeID:        1,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       2,
							Participant: "home",
						},
						Description: "1ST_HALF", // Different description
					},
					{
						ID:            2,
						FixtureID:     12346,
						TypeID:        1,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       1,
							Participant: "away",
						},
						Description: "1ST_HALF", // Different description
					},
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    5, // Finished match
				HasOdds:    true,
			},
			expectedHome: 2, // Fixed: should extract from HALF scores
			expectedAway: 1, // Fixed: should extract from HALF scores
			description:  "Fixed: Should extract scores from HALF descriptions when CURRENT not available",
		},
		{
			name: "Finished match with empty scores array",
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
				Scores:     []score{}, // Empty scores array
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    5, // Finished match
				HasOdds:    true,
			},
			expectedHome: 0, // Empty scores array should still result in 0-0
			expectedAway: 0, // Empty scores array should still result in 0-0
			description:  "Should handle empty scores array for finished match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := tt.fixture.toMatch()
			if err != nil {
				t.Fatalf("toMatch() error = %v", err)
			}

			homeGoals := match.GetHomeGoals()
			awayGoals := match.GetAwayGoals()

			if homeGoals != tt.expectedHome {
				t.Errorf("Home goals = %d, want %d", homeGoals, tt.expectedHome)
			}
			if awayGoals != tt.expectedAway {
				t.Errorf("Away goals = %d, want %d", awayGoals, tt.expectedAway)
			}

			// Check if this would result in a 0-0 draw when it shouldn't (the bug scenario)
			// Only flag as bug if we have scores in the fixture but they're not being extracted
			if homeGoals == 0 && awayGoals == 0 && match.IsFinished() && len(tt.fixture.Scores) > 0 {
				// Check if we have any non-zero scores that should have been extracted
				hasNonZeroScores := false
				for _, s := range tt.fixture.Scores {
					if s.Score.Goals > 0 {
						hasNonZeroScores = true
						break
					}
				}

				if hasNonZeroScores {
					t.Errorf("BUG DETECTED: Finished match with non-zero scores appears as 0-0 draw: %s", tt.description)
					t.Logf("Match ID: %s", match.Id())
					t.Logf("Match status: %s", match.GetStatus())
					t.Logf("IsDraw: %t", match.IsDraw())
					t.Logf("Fixture has %d scores, some with non-zero goals", len(tt.fixture.Scores))
				}
			}

			t.Logf("Test: %s - %s", tt.name, tt.description)
			t.Logf("Result: %s vs %s (%d-%d), Status: %s, IsDraw: %t",
				match.GetHomeTeam(), match.GetAwayTeam(),
				homeGoals, awayGoals,
				match.GetStatus(), match.IsDraw())
		})
	}
}

// TestBug_ScoreExtraction_ShouldExtractAnyScores tests that the current behavior is incorrect
// This test will fail when the bug is fixed, showing that we should extract scores from any description
func TestBug_ScoreExtraction_ShouldExtractAnyScores(t *testing.T) {
	// This test demonstrates the bug: finished matches should have scores extracted
	// regardless of the score description, not just "CURRENT"

	fixture := sportmonksFixture{
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
		Scores: []score{
			{
				ID:            1,
				FixtureID:     12348,
				TypeID:        1,
				ParticipantID: 1,
				Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{
					Goals:       2,
					Participant: "home",
				},
				Description: "1ST_HALF", // Different description, but should still be extracted
			},
			{
				ID:            2,
				FixtureID:     12348,
				TypeID:        1,
				ParticipantID: 2,
				Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{
					Goals:       1,
					Participant: "away",
				},
				Description: "1ST_HALF", // Different description, but should still be extracted
			},
		},
		StartingAt: "2025-01-01 15:00:00",
		Season:     season{Name: "2024/2025"},
		League:     league{Name: "Ligue 1"},
		Round:      round{Name: "1"},
		StateID:    5, // Finished match
		HasOdds:    true,
	}

	match, err := fixture.toMatch()
	if err != nil {
		t.Fatalf("toMatch() error = %v", err)
	}

	homeGoals := match.GetHomeGoals()
	awayGoals := match.GetAwayGoals()

	// CURRENT BUG: This should be 2-1, but it's 0-0 because scores aren't extracted
	// When the bug is fixed, this test will fail and show the correct behavior
	if homeGoals == 0 && awayGoals == 0 {
		t.Logf("BUG CONFIRMED: Finished match with scores in 1ST_HALF description appears as 0-0")
		t.Logf("Match ID: %s", match.Id())
		t.Logf("Match status: %s", match.GetStatus())
		t.Logf("IsDraw: %t", match.IsDraw())
		t.Logf("Expected: 2-1, Got: %d-%d", homeGoals, awayGoals)
		t.Logf("This explains why players only get points for draws!")
	} else {
		t.Logf("Bug fixed! Match correctly shows %d-%d", homeGoals, awayGoals)
	}

	// This test documents the expected behavior when the bug is fixed
	expectedHome := 2
	expectedAway := 1

	if homeGoals != expectedHome {
		t.Errorf("Home goals = %d, want %d (bug not fixed yet)", homeGoals, expectedHome)
	}
	if awayGoals != expectedAway {
		t.Errorf("Away goals = %d, want %d (bug not fixed yet)", awayGoals, expectedAway)
	}

	t.Logf("Test result: %s vs %s (%d-%d), Status: %s, IsDraw: %t",
		match.GetHomeTeam(), match.GetAwayTeam(),
		homeGoals, awayGoals,
		match.GetStatus(), match.IsDraw())
}

// TestBug_ScoreExtraction_TimingIssue tests the timing issue where scores appear differently
// depending on when the API is called relative to match completion
func TestBug_ScoreExtraction_TimingIssue(t *testing.T) {
	tests := []struct {
		name         string
		fixture      sportmonksFixture
		expectedHome int
		expectedAway int
		description  string
	}{
		{
			name: "Match just finished - only HALF scores available (BUG SCENARIO)",
			fixture: sportmonksFixture{
				ID: 12349,
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
					{
						ID:            1,
						FixtureID:     12349,
						TypeID:        1,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       1,
							Participant: "home",
						},
						Description: "1ST_HALF",
					},
					{
						ID:            2,
						FixtureID:     12349,
						TypeID:        1,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "away",
						},
						Description: "1ST_HALF",
					},
					{
						ID:            3,
						FixtureID:     12349,
						TypeID:        2,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       1,
							Participant: "home",
						},
						Description: "2ND_HALF",
					},
					{
						ID:            4,
						FixtureID:     12349,
						TypeID:        2,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "away",
						},
						Description: "2ND_HALF",
					},
					// No CURRENT scores yet - this is the timing issue!
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    5, // Finished match
				HasOdds:    true,
			},
			expectedHome: 2, // Should be 1+1 = 2 from HALF scores
			expectedAway: 0, // Should be 0+0 = 0 from HALF scores
			description:  "Should extract scores from HALF descriptions when CURRENT not available",
		},
		{
			name: "Match finished later - CURRENT scores available",
			fixture: sportmonksFixture{
				ID: 12350,
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
					{
						ID:            1,
						FixtureID:     12350,
						TypeID:        1525,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       2,
							Participant: "home",
						},
						Description: "CURRENT",
					},
					{
						ID:            2,
						FixtureID:     12350,
						TypeID:        1525,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "away",
						},
						Description: "CURRENT",
					},
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    5, // Finished match
				HasOdds:    true,
			},
			expectedHome: 2,
			expectedAway: 0,
			description:  "Should extract CURRENT scores when available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := tt.fixture.toMatch()
			if err != nil {
				t.Fatalf("toMatch() error = %v", err)
			}

			homeGoals := match.GetHomeGoals()
			awayGoals := match.GetAwayGoals()

			// Check if this would result in a 0-0 draw (the bug scenario)
			if homeGoals == 0 && awayGoals == 0 && match.IsFinished() {
				t.Errorf("BUG DETECTED: Finished match appears as 0-0 draw: %s", tt.description)
				t.Logf("Match ID: %s", match.Id())
				t.Logf("Match status: %s", match.GetStatus())
				t.Logf("IsDraw: %t", match.IsDraw())
				t.Logf("This timing issue explains why matches stop being 'live'!")
			}

			// Check if scores are correct
			if homeGoals != tt.expectedHome {
				t.Errorf("Home goals = %d, want %d", homeGoals, tt.expectedHome)
			}
			if awayGoals != tt.expectedAway {
				t.Errorf("Away goals = %d, want %d", awayGoals, tt.expectedAway)
			}

			t.Logf("Test: %s - %s", tt.name, tt.description)
			t.Logf("Result: %s vs %s (%d-%d), Status: %s, IsDraw: %t",
				match.GetHomeTeam(), match.GetAwayTeam(),
				homeGoals, awayGoals,
				match.GetStatus(), match.IsDraw())
		})
	}
}

// TestBug_ScoreExtraction_LegitimateZeroZero tests that legitimate 0-0 draws are handled correctly
func TestBug_ScoreExtraction_LegitimateZeroZero(t *testing.T) {
	tests := []struct {
		name         string
		fixture      sportmonksFixture
		expectedHome int
		expectedAway int
		description  string
	}{
		{
			name: "Legitimate 0-0 draw with CURRENT scores",
			fixture: sportmonksFixture{
				ID: 12351,
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
					{
						ID:            1,
						FixtureID:     12351,
						TypeID:        1525,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "home",
						},
						Description: "CURRENT",
					},
					{
						ID:            2,
						FixtureID:     12351,
						TypeID:        1525,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "away",
						},
						Description: "CURRENT",
					},
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    5, // Finished match
				HasOdds:    true,
			},
			expectedHome: 0,
			expectedAway: 0,
			description:  "Should correctly identify legitimate 0-0 draw",
		},
		{
			name: "Legitimate 0-0 draw with HALF scores",
			fixture: sportmonksFixture{
				ID: 12352,
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
					{
						ID:            1,
						FixtureID:     12352,
						TypeID:        1,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "home",
						},
						Description: "1ST_HALF",
					},
					{
						ID:            2,
						FixtureID:     12352,
						TypeID:        1,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "away",
						},
						Description: "1ST_HALF",
					},
					{
						ID:            3,
						FixtureID:     12352,
						TypeID:        2,
						ParticipantID: 1,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "home",
						},
						Description: "2ND_HALF",
					},
					{
						ID:            4,
						FixtureID:     12352,
						TypeID:        2,
						ParticipantID: 2,
						Score: struct {
							Goals       int    `json:"goals"`
							Participant string `json:"participant"`
						}{
							Goals:       0,
							Participant: "away",
						},
						Description: "2ND_HALF",
					},
				},
				StartingAt: "2025-01-01 15:00:00",
				Season:     season{Name: "2024/2025"},
				League:     league{Name: "Ligue 1"},
				Round:      round{Name: "1"},
				StateID:    5, // Finished match
				HasOdds:    true,
			},
			expectedHome: 0,
			expectedAway: 0,
			description:  "Should correctly identify legitimate 0-0 draw from HALF scores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := tt.fixture.toMatch()
			if err != nil {
				t.Fatalf("toMatch() error = %v", err)
			}

			homeGoals := match.GetHomeGoals()
			awayGoals := match.GetAwayGoals()

			// Check if scores are correct
			if homeGoals != tt.expectedHome {
				t.Errorf("Home goals = %d, want %d", homeGoals, tt.expectedHome)
			}
			if awayGoals != tt.expectedAway {
				t.Errorf("Away goals = %d, want %d", awayGoals, tt.expectedAway)
			}

			// Verify this is correctly identified as a draw
			if !match.IsDraw() {
				t.Errorf("Match should be identified as a draw, but IsDraw() returned false")
			}

			t.Logf("Test: %s - %s", tt.name, tt.description)
			t.Logf("Result: %s vs %s (%d-%d), Status: %s, IsDraw: %t",
				match.GetHomeTeam(), match.GetAwayTeam(),
				homeGoals, awayGoals,
				match.GetStatus(), match.IsDraw())
		})
	}
}

func TestScoreExtraction_AllScoreDescriptions(t *testing.T) {
	tests := []struct {
		name         string
		scores       []score
		expectedHome int
		expectedAway int
		description  string
	}{
		{
			name: "CURRENT scores only",
			scores: []score{
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 2}},
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 1}},
			},
			expectedHome: 2,
			expectedAway: 1,
			description:  "Should use CURRENT scores when available",
		},
		{
			name: "2ND_HALF scores (final score after 2nd half)",
			scores: []score{
				{Description: "2ND_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 3}},
				{Description: "2ND_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 2}},
			},
			expectedHome: 3,
			expectedAway: 2,
			description:  "Should use 2ND_HALF scores when CURRENT not available",
		},
		{
			name: "EXTRA TIME scores",
			scores: []score{
				{Description: "ET", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 2}},
				{Description: "ET", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 1}},
			},
			expectedHome: 2,
			expectedAway: 1,
			description:  "Should use ET scores for extra time matches",
		},
		{
			name: "PENALTIES scores",
			scores: []score{
				{Description: "PENALTIES", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 4}},
				{Description: "PENALTIES", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 3}},
			},
			expectedHome: 4,
			expectedAway: 3,
			description:  "Should use PENALTIES scores for penalty shootouts",
		},
		{
			name: "Sum 1ST_HALF and 2ND_HALF_ONLY",
			scores: []score{
				{Description: "1ST_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 1}},
				{Description: "1ST_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 0}},
				{Description: "2ND_HALF_ONLY", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 1}},
				{Description: "2ND_HALF_ONLY", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 1}},
			},
			expectedHome: 2,
			expectedAway: 1,
			description:  "Should sum 1ST_HALF and 2ND_HALF_ONLY when 2ND_HALF not available",
		},
		{
			name: "Fallback to any available score",
			scores: []score{
				{Description: "1ST_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 1}},
				{Description: "1ST_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 0}},
			},
			expectedHome: 1,
			expectedAway: 0,
			description:  "Should use any available score as last resort",
		},
		{
			name: "Mixed score types with CURRENT priority",
			scores: []score{
				{Description: "1ST_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 0}},
				{Description: "1ST_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 0}},
				{Description: "2ND_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 2}},
				{Description: "2ND_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 1}},
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 2}},
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 1}},
			},
			expectedHome: 2,
			expectedAway: 1,
			description:  "Should prioritize CURRENT over other score types",
		},
		{
			name:         "Zero-zero draw with no scores",
			scores:       []score{},
			expectedHome: 0,
			expectedAway: 0,
			description:  "Should default to 0-0 when no scores available",
		},
		{
			name: "Zero-zero draw with zero scores",
			scores: []score{
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 0}},
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 0}},
			},
			expectedHome: 0,
			expectedAway: 0,
			description:  "Should correctly handle legitimate 0-0 draws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := sportmonksFixture{
				ID:      1,
				StateID: 5, // Finished match
				Scores:  tt.scores,
			}

			match, err := fixture.toMatch()
			if err != nil {
				t.Fatalf("toMatch() error = %v", err)
			}

			if match.GetHomeGoals() != tt.expectedHome {
				t.Errorf("Home goals = %d, expected %d. %s", match.GetHomeGoals(), tt.expectedHome, tt.description)
			}
			if match.GetAwayGoals() != tt.expectedAway {
				t.Errorf("Away goals = %d, expected %d. %s", match.GetAwayGoals(), tt.expectedAway, tt.description)
			}
		})
	}
}

func TestExtractScores_Function(t *testing.T) {
	tests := []struct {
		name         string
		scores       []score
		expectedHome int
		expectedAway int
		description  string
	}{
		{
			name: "CURRENT scores only",
			scores: []score{
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 2}},
				{Description: "CURRENT", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 1}},
			},
			expectedHome: 2,
			expectedAway: 1,
			description:  "Should use CURRENT scores when available",
		},
		{
			name: "2ND_HALF scores (final score after 2nd half)",
			scores: []score{
				{Description: "2ND_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "home", Goals: 3}},
				{Description: "2ND_HALF", Score: struct {
					Goals       int    `json:"goals"`
					Participant string `json:"participant"`
				}{Participant: "away", Goals: 2}},
			},
			expectedHome: 3,
			expectedAway: 2,
			description:  "Should use 2ND_HALF scores when CURRENT not available",
		},
		{
			name:         "Zero-zero draw with no scores",
			scores:       []score{},
			expectedHome: 0,
			expectedAway: 0,
			description:  "Should default to 0-0 when no scores available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := sportmonksFixture{
				ID:      1,
				StateID: 5, // Finished match
				Scores:  tt.scores,
			}

			homeScore, awayScore := fixture.extractScores()

			if *homeScore != tt.expectedHome {
				t.Errorf("Home goals = %d, expected %d. %s", *homeScore, tt.expectedHome, tt.description)
			}
			if *awayScore != tt.expectedAway {
				t.Errorf("Away goals = %d, expected %d. %s", *awayScore, tt.expectedAway, tt.description)
			}
		})
	}
}
