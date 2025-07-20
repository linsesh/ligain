import React from 'react';
import { SeasonMatch } from '../types/match';

describe('Odds Display with Clear Favorite Indicators', () => {
  test('shows star for home team favorite', () => {
    const match = new SeasonMatch(
      'Manchester United',
      'Liverpool',
      0,
      0,
      1.0,  // home odds (favorite)
      3.0,  // away odds
      2.0,  // draw odds
      'scheduled',
      '2024',
      'Premier League',
      new Date('2024-01-01T15:00:00Z'),
      1
    );

    // Check that the logic works correctly
    expect(match.hasClearFavorite()).toBe(true);
    expect(match.getFavoriteTeam()).toBe('Manchester United');
    
    // Check that the correct indicators would be shown
    const homeIsFavorite = match.getFavoriteTeam() === match.getHomeTeam();
    const awayIsFavorite = match.getFavoriteTeam() === match.getAwayTeam();
    
    expect(homeIsFavorite).toBe(true);
    expect(awayIsFavorite).toBe(false);
  });

  test('shows star for away team favorite', () => {
    const match = new SeasonMatch(
      'Arsenal',
      'Chelsea',
      0,
      0,
      3.0,  // home odds
      1.0,  // away odds (favorite)
      2.0,  // draw odds
      'scheduled',
      '2024',
      'Premier League',
      new Date('2024-01-01T15:00:00Z'),
      1
    );

    // Check that the logic works correctly
    expect(match.hasClearFavorite()).toBe(true);
    expect(match.getFavoriteTeam()).toBe('Chelsea');
    
    // Check that the correct indicators would be shown
    const homeIsFavorite = match.getFavoriteTeam() === match.getHomeTeam();
    const awayIsFavorite = match.getFavoriteTeam() === match.getAwayTeam();
    
    expect(homeIsFavorite).toBe(false);
    expect(awayIsFavorite).toBe(true);
  });

  test('no indicators when no clear favorite', () => {
    const match = new SeasonMatch(
      'Tottenham',
      'West Ham',
      0,
      0,
      1.5,  // home odds
      2.0,  // away odds (difference = 0.5 <= 1.5)
      3.0,  // draw odds
      'scheduled',
      '2024',
      'Premier League',
      new Date('2024-01-01T15:00:00Z'),
      1
    );

    // Check that the logic works correctly
    expect(match.hasClearFavorite()).toBe(false);
    expect(match.getFavoriteTeam()).toBe('');
  });

  test('odds difference calculation', () => {
    const match = new SeasonMatch(
      'Team A',
      'Team B',
      0,
      0,
      1.0,  // home odds
      3.0,  // away odds (difference = 2.0)
      2.0,  // draw odds
      'scheduled',
      '2024',
      'Premier League',
      new Date('2024-01-01T15:00:00Z'),
      1
    );

    expect(match.absoluteDifferenceOddsBetweenHomeAndAway()).toBe(2.0);
    expect(match.hasClearFavorite()).toBe(true);
  });

  test('borderline case - exactly 1.5 difference', () => {
    const match = new SeasonMatch(
      'Team A',
      'Team B',
      0,
      0,
      1.0,  // home odds
      2.5,  // away odds (difference = 1.5)
      2.0,  // draw odds
      'scheduled',
      '2024',
      'Premier League',
      new Date('2024-01-01T15:00:00Z'),
      1
    );

    expect(match.absoluteDifferenceOddsBetweenHomeAndAway()).toBe(1.5);
    expect(match.hasClearFavorite()).toBe(false); // Should be false for exactly 1.5
  });
}); 