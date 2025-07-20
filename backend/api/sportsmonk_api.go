package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"liguain/backend/models"
	"liguain/backend/utils"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// SportsmonkAPI is a wrapper around the Sportsmonk API
type SportsmonkAPI interface {
	// GetSeasonIds creates a mapping between the season code and the season ID, for a given competition ID
	GetSeasonIds(seasonCodes []string, competitionId int) (map[string]int, error)
	// GetFixturesInfos returns the fixtures infos for a given list of fixture IDs
	GetFixturesInfos(fixtureIds []int) (map[int]models.Match, error)
	// GetSeasonFixtures returns a map of fixtureId - models.Match for a given season ID
	GetSeasonFixtures(seasonId int) (map[int]models.Match, error)
	// GetCompetitionId returns the sportsmonk ID for a given competition code
	GetCompetitionId(competitionCode string) (int, error)
}

type SportsmonkAPIImpl struct {
	apiToken string
}

type seasonsResponse struct {
	Data []season `json:"data"`
}

// League represents the league data in the API response
type league struct {
	ID        int    `json:"id"`
	SportID   int    `json:"sport_id"`
	CountryID int    `json:"country_id"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	ShortCode string `json:"short_code"`
	ImagePath string `json:"image_path"`
	Type      string `json:"type"`
	SubType   string `json:"sub_type"`
	Category  int    `json:"category"`
}

// Season represents the season data in the API response
type season struct {
	ID         int    `json:"id"`
	SportID    int    `json:"sport_id"`
	LeagueID   int    `json:"league_id"`
	Name       string `json:"name"`
	Finished   bool   `json:"finished"`
	Pending    bool   `json:"pending"`
	IsCurrent  bool   `json:"is_current"`
	StartingAt string `json:"starting_at"`
	EndingAt   string `json:"ending_at"`
}

// Round represents the round data in the API response
type round struct {
	ID         int    `json:"id"`
	SportID    int    `json:"sport_id"`
	LeagueID   int    `json:"league_id"`
	SeasonID   int    `json:"season_id"`
	StageID    int    `json:"stage_id"`
	Name       string `json:"name"`
	Finished   bool   `json:"finished"`
	IsCurrent  bool   `json:"is_current"`
	StartingAt string `json:"starting_at"`
	EndingAt   string `json:"ending_at"`
}

// Score represents a score entry in the API response
type score struct {
	Type      string `json:"type"`
	HomeScore int    `json:"home_score"`
	AwayScore int    `json:"away_score"`
}

// Participant represents a team participant in the API response
type participant struct {
	ID        int    `json:"id"`
	SportID   int    `json:"sport_id"`
	CountryID int    `json:"country_id"`
	VenueID   int    `json:"venue_id"`
	Name      string `json:"name"`
	ShortCode string `json:"short_code"`
	Type      string `json:"type"`
	Meta      struct {
		Location string `json:"location"` // "home" or "away"
		Winner   *bool  `json:"winner"`
		Position int    `json:"position"`
	} `json:"meta"`
}

// Odd represents an odd entry in the API response
type odd struct {
	ID          int64  `json:"id"`
	FixtureID   int    `json:"fixture_id"`
	MarketID    int    `json:"market_id"`
	BookmakerID int    `json:"bookmaker_id"`
	Label       string `json:"label"`
	Value       string `json:"value"`
}

// sportmonksFixture represents the raw fixture data from Sportmonk API
type sportmonksFixture struct {
	ID           int           `json:"id"`
	LeagueID     int           `json:"league_id"`
	SeasonID     int           `json:"season_id"`
	RoundID      int           `json:"round_id"`
	StateID      int           `json:"state_id"`
	VenueID      int           `json:"venue_id"`
	Name         string        `json:"name"`
	StartingAt   string        `json:"starting_at"`
	HasOdds      bool          `json:"has_odds"`
	League       league        `json:"league"`
	Season       season        `json:"season"`
	Round        round         `json:"round"`
	Scores       []score       `json:"scores"`
	Participants []participant `json:"participants"`
	Odds         []odd         `json:"odds"`
}

type fixturesResponse struct {
	Data []sportmonksFixture `json:"data"`
}

type allFixturesResponse struct {
	Data       []sportmonksFixture `json:"data"`
	Pagination struct {
		Count       int    `json:"count"`
		PerPage     int    `json:"per_page"`
		CurrentPage int    `json:"current_page"`
		NextPage    string `json:"next_page"`
		HasMore     bool   `json:"has_more"`
	} `json:"pagination"`
}

// toMatch converts a sportmonksFixture to a models.Match
func (f *sportmonksFixture) toMatch() (models.Match, error) {
	// Parse the timestamp
	startTime, err := time.Parse("2006-01-02 15:04:05", f.StartingAt)
	if err != nil {
		return nil, err
	}

	// Extract home and away teams
	var homeTeam, awayTeam *participant
	for i, p := range f.Participants {
		if p.Meta.Location == "home" {
			homeTeam = &f.Participants[i]
		} else if p.Meta.Location == "away" {
			awayTeam = &f.Participants[i]
		}
	}
	if homeTeam == nil || awayTeam == nil {
		return nil, fmt.Errorf("could not find home or away team in participants")
	}

	// Extract full time score (if available)
	homeScore, awayScore := 0, 0
	for _, s := range f.Scores {
		if s.Type == "FT" || s.Type == "fulltime" || s.Type == "" { // fallback if type is missing
			homeScore = s.HomeScore
			awayScore = s.AwayScore
			break
		}
	}

	// Extract odds for home, draw, away (market_id=1, bookmaker_id=2 is bet365)
	var homeOdd, drawOdd, awayOdd float64
	for _, o := range f.Odds {
		if o.MarketID == 1 && o.BookmakerID == 2 {
			switch o.Label {
			case "Home":
				homeOdd, _ = strconv.ParseFloat(o.Value, 64)
			case "Draw":
				drawOdd, _ = strconv.ParseFloat(o.Value, 64)
			case "Away":
				awayOdd, _ = strconv.ParseFloat(o.Value, 64)
			}
		}
	}

	// Extract the matchday
	matchday := 1 // default value
	if f.Round.Name != "" {
		// Try to parse the round name as a number (e.g., "34" -> 34)
		if m, err := strconv.Atoi(f.Round.Name); err == nil {
			matchday = m
		}
	}

	// State ID 5 means the match is finished
	if f.StateID == 5 {
		return models.NewFinishedSeasonMatch(
			homeTeam.Name,
			awayTeam.Name,
			homeScore,
			awayScore,
			f.Season.Name,
			f.League.Name,
			startTime,
			matchday,
			homeOdd,
			awayOdd,
			drawOdd,
		), nil
	}

	if f.HasOdds {
		return models.NewSeasonMatchWithKnownOdds(
			homeTeam.Name,
			awayTeam.Name,
			f.Season.Name,
			f.League.Name,
			startTime,
			matchday,
			homeOdd,
			awayOdd,
			drawOdd,
		), nil
	}

	return models.NewSeasonMatch(
		homeTeam.Name,
		awayTeam.Name,
		f.Season.Name,
		f.League.Name,
		startTime,
		matchday,
	), nil
}

const baseURL = "https://api.sportmonks.com/v3/football/"
const ligue1LigueId = 301

func NewSportsmonkAPI(apiToken string) *SportsmonkAPIImpl {
	return &SportsmonkAPIImpl{apiToken: apiToken}
}

func (s *SportsmonkAPIImpl) GetSeasonIds(seasonCodes []string, competitionId int) (map[string]int, error) {
	seasons := make(chan []season)
	errChan := make(chan error)
	//should we forward the context from the caller?
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go s.fetchSeasons(competitionId, ctx, seasons, errChan)

	seasonIds := make(map[string]int)
	select {
	case seasons := <-seasons:
		for _, season := range seasons {
			if slices.Contains(seasonCodes, season.Name) {
				seasonIds[season.Name] = season.ID
			}
		}
		return seasonIds, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("request timed out after 1 second")
	}
}

func (s *SportsmonkAPIImpl) fetchSeasons(competitionId int, ctx context.Context, resultChan chan<- []season, errChan chan<- error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%sseasons?filters=seasonLeagues:%d", baseURL, competitionId), nil)
	if err != nil {
		errChan <- err
		return
	}
	query := s.basicQuery(req)
	req.URL.RawQuery = query.Encode()
	resp, err := s.makeRequest(req)
	if err != nil {
		errChan <- err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("unexpected status: %s", resp.Status)
		return
	}

	var result seasonsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		errChan <- err
		return
	}

	resultChan <- result.Data
}

func (s *SportsmonkAPIImpl) GetSeasonFixtures(seasonId int) (map[int]models.Match, error) {
	seasonFixtures := make(chan map[int]models.Match)
	errChan := make(chan error)
	//should we forward the context from the caller?
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go s.fetchSeasonFixtures(seasonId, ctx, seasonFixtures, errChan)

	select {
	case seasonFixtures := <-seasonFixtures:
		return seasonFixtures, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("request timed out after 30 seconds")
	}
}

func (s *SportsmonkAPIImpl) fetchSeasonFixtures(seasonId int, ctx context.Context, resultChan chan<- map[int]models.Match, errChan chan<- error) {
	// Map to store all fixtures across pages
	allFixtures := make(map[int]models.Match)
	// Map to track matches by their unique ID to handle duplicates
	matchIdToFixtureId := make(map[string]int)
	currentPage := 1
	hasMore := true

	log.Printf("Starting to fetch fixtures for season ID %d", seasonId)

	// Loop through all pages
	for hasMore {
		log.Printf("Fetching page %d of fixtures for season ID %d", currentPage, seasonId)

		// Use the fixtures endpoint with a season filter
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%sfixtures", baseURL), nil)
		if err != nil {
			errChan <- err
			return
		}
		query := s.basicQuery(req)
		// Add the season filter
		query.Add("filters", fmt.Sprintf("fixtureSeasons:%d;bookmakers:1;markets:1", seasonId))
		// Use semicolons for includes
		query.Add("include", "league;season;round;scores;participants;odds")
		// Add pagination parameters
		query.Add("page", strconv.Itoa(currentPage))
		query.Add("per_page", "25") // Set a reasonable page size

		req.URL.RawQuery = query.Encode()
		resp, err := s.makeRequest(req)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("unexpected status: %s", resp.Status)
			return
		}

		// Parse the response with pagination info
		var responseBody allFixturesResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			errChan <- err
			return
		}

		fixtureCount := len(responseBody.Data)
		log.Printf("Received %d fixtures on page %d for season ID %d", fixtureCount, currentPage, seasonId)

		// Log first few fixtures for debugging
		for i, fixture := range responseBody.Data {
			if i < 3 { // Only log first 3 fixtures per page to avoid spam
				log.Printf("  Fixture %d: ID=%d, Name='%s', StartingAt='%s', Round='%s'",
					i+1, fixture.ID, fixture.Name, fixture.StartingAt, fixture.Round.Name)
			}
		}

		// Convert the fixtures to a map of models.Match and add to our collection
		for _, fixture := range responseBody.Data {
			match, err := fixture.toMatch()
			if err != nil {
				errChan <- err
				return
			}

			matchId := match.Id()

			// Check if we already have a fixture for this match
			if existingFixtureId, exists := matchIdToFixtureId[matchId]; exists {
				existingMatch := allFixtures[existingFixtureId]
				log.Printf("Duplicate match found: %s", matchId)
				log.Printf("  Existing fixture ID %d: %s vs %s (%s)", existingFixtureId, existingMatch.GetHomeTeam(), existingMatch.GetAwayTeam(), existingMatch.GetDate().Format("2006-01-02 15:04"))
				log.Printf("  New fixture ID %d: %s vs %s (%s)", fixture.ID, match.GetHomeTeam(), match.GetAwayTeam(), match.GetDate().Format("2006-01-02 15:04"))

				// Keep the fixture with the later time (more likely to be the current schedule)
				if match.GetDate().After(existingMatch.GetDate()) {
					log.Printf("  Replacing with later time fixture")
					delete(allFixtures, existingFixtureId)
					matchIdToFixtureId[matchId] = fixture.ID
					allFixtures[fixture.ID] = match
				} else {
					log.Printf("  Keeping existing fixture (earlier time)")
				}
				continue
			}

			// This is a new match, add it
			matchIdToFixtureId[matchId] = fixture.ID
			allFixtures[fixture.ID] = match
		}

		// Check if there are more pages
		hasMore = responseBody.Pagination.HasMore
		currentPage++

		log.Printf("Processed page %d for season ID %d. Has more pages: %v", currentPage-1, seasonId, hasMore)
	}

	log.Printf("Completed fetching all fixtures for season ID %d. Total fixtures: %d (after deduplication)", seasonId, len(allFixtures))
	resultChan <- allFixtures
}

func (s *SportsmonkAPIImpl) GetFixturesInfos(fixtureIds []int) (map[int]models.Match, error) {
	fixtureIdsStr := utils.ConvertIntSliceToStringWithCommas(fixtureIds)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%sfixtures/multi/%s", baseURL, fixtureIdsStr), nil)
	if err != nil {
		return nil, err
	}
	query := s.basicQuery(req)
	// Use semicolons for includes - this is the key change
	query.Add("include", "league;season;round;scores;participants;odds")
	// Filter for specific bookmaker and market if needed
	query.Add("filters", "bookmakers:1;markets:1")
	req.URL.RawQuery = query.Encode()
	resp, err := s.makeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var fixturesResp fixturesResponse
	if err := json.NewDecoder(resp.Body).Decode(&fixturesResp); err != nil {
		return nil, err
	}

	fixtureIdToMatch := make(map[int]models.Match)
	for _, fixture := range fixturesResp.Data {
		match, err := fixture.toMatch()
		if err != nil {
			return nil, err
		}
		fixtureIdToMatch[fixture.ID] = match
	}
	return fixtureIdToMatch, nil
}

func (s *SportsmonkAPIImpl) GetCompetitionId(competitionCode string) (int, error) {
	if competitionCode == "Ligue 1" {
		return ligue1LigueId, nil
	}
	return -1, fmt.Errorf("competition code not supported: %s", competitionCode)
}

func (s *SportsmonkAPIImpl) makeRequest(req *http.Request) (resp *http.Response, err error) {
	log.Printf("Request: %s", req.URL.String())
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	// Read the body into a buffer
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Create a new reader with the bytes for later use
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Just print the status and raw response
	//fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(bodyBytes))

	return resp, nil
}

func (s *SportsmonkAPIImpl) basicQuery(req *http.Request) url.Values {
	query := req.URL.Query()
	query.Add("api_token", s.apiToken)
	return query
}
