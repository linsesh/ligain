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

type season struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type fixtures struct {
	Data []string `json:"data"`
}

type seasonsResponse struct {
	Data []season `json:"data"`
}

// sportmonkFixture represents the raw fixture data from Sportmonk API
type sportmonksFixture struct {
	ID         int    `json:"id"`
	SportID    int    `json:"sport_id"`
	LeagueID   int    `json:"league_id"`
	SeasonID   int    `json:"season_id"`
	StageID    int    `json:"stage_id"`
	RoundID    int    `json:"round_id"`
	StateID    int    `json:"state_id"`
	VenueID    int    `json:"venue_id"`
	Name       string `json:"name"`
	StartingAt string `json:"starting_at"`
	ResultInfo string `json:"result_info"`
	Leg        string `json:"leg"`
	Length     int    `json:"length"`
	HasOdds    bool   `json:"has_odds"`
	// Include relationships
	League struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"league"`
	Season struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"season"`
	Round struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Matchday int    `json:"matchday"`
	} `json:"round"`
	Scores struct {
		HomeScore int `json:"home_score"`
		AwayScore int `json:"away_score"`
	} `json:"scores"`
	HomeTeam struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"home"`
	AwayTeam struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"away"`
	Odds struct {
		HomeWin float64 `json:"home_win"`
		Draw    float64 `json:"draw"`
		AwayWin float64 `json:"away_win"`
	} `json:"odds"`
}

type fixturesResponse struct {
	Data []sportmonksFixture `json:"data"`
}

// toMatch converts a sportmonkFixture to a models.Match
func (f *sportmonksFixture) toMatch() (models.Match, error) {
	// Parse the timestamp
	startTime, err := time.Parse("2006-01-02 15:04:05", f.StartingAt)
	if err != nil {
		return nil, err
	}

	// State ID 5 means the match is finished
	if f.StateID == 5 {
		return models.NewFinishedSeasonMatch(
			f.HomeTeam.Name,
			f.AwayTeam.Name,
			f.Scores.HomeScore,
			f.Scores.AwayScore,
			f.Season.Name,
			f.League.Name,
			startTime,
			f.Round.Matchday,
			f.Odds.HomeWin,
			f.Odds.AwayWin,
			f.Odds.Draw,
		), nil
	}

	if f.HasOdds {
		return models.NewSeasonMatchWithKnownOdds(
			f.HomeTeam.Name,
			f.AwayTeam.Name,
			f.Season.Name,
			f.League.Name,
			startTime,
			f.Round.Matchday,
			f.Odds.HomeWin,
			f.Odds.AwayWin,
			f.Odds.Draw,
		), nil
	}

	return models.NewSeasonMatch(
		f.HomeTeam.Name,
		f.AwayTeam.Name,
		f.Season.Name,
		f.League.Name,
		startTime,
		f.Round.Matchday,
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go s.fetchSeasonFixtures(seasonId, ctx, seasonFixtures, errChan)

	select {
	case seasonFixtures := <-seasonFixtures:
		return seasonFixtures, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("request timed out after 1 second")
	}
}

func (s *SportsmonkAPIImpl) fetchSeasonFixtures(seasonId int, ctx context.Context, resultChan chan<- map[int]models.Match, errChan chan<- error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%sseasons/%d/fixtures", baseURL, seasonId), nil)
	if err != nil {
		errChan <- err
		return
	}
	query := s.basicQuery(req)
	query.Add("include", "league,season,round,scores,home,away,odds")
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

	var fixturesResp fixturesResponse
	if err := json.NewDecoder(resp.Body).Decode(&fixturesResp); err != nil {
		errChan <- err
		return
	}

	// Convert the fixtures to a map of models.Match
	fixtureIdToMatch := make(map[int]models.Match)
	for _, fixture := range fixturesResp.Data {
		match, err := fixture.toMatch()
		if err != nil {
			errChan <- err
			return
		}
		fixtureIdToMatch[fixture.ID] = match
	}

	resultChan <- fixtureIdToMatch
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
	query.Add("include", "league,season,round,scores,home,away,odds")
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
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(bodyBytes))

	return resp, nil
}

func (s *SportsmonkAPIImpl) basicQuery(req *http.Request) url.Values {
	query := req.URL.Query()
	query.Add("api_token", s.apiToken)
	return query
}
