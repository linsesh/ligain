import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import { useLocalSearchParams } from 'expo-router';
import { SeasonMatch, MatchResult } from '../types/match';
import { MockTimeService } from '../services/timeService';
import { TimeServiceProvider } from '../contexts/TimeServiceContext';
import { AuthProvider } from '../contexts/AuthContext';

// Mock expo-router
jest.mock('expo-router', () => ({
  useLocalSearchParams: jest.fn(),
  useRouter: jest.fn(() => ({
    push: jest.fn(),
  })),
}));

// Mock the hooks
jest.mock('../../hooks/useBetSubmission', () => ({
  useBetSubmission: jest.fn(() => ({
    submitBet: jest.fn(),
    error: null,
  })),
}));

// Mock fetch
global.fetch = jest.fn();

// Mock the API config
jest.mock('../config/api', () => ({
  API_CONFIG: {
    BASE_URL: 'http://localhost:8080',
    API_KEY: 'test-api-key',
  },
  getAuthenticatedHeaders: jest.fn(() => Promise.resolve({
    'X-API-Key': 'test-api-key',
    'Authorization': 'Bearer test-token',
  })),
}));

// Helper function to create test matches
const createTestMatch = (
  homeTeam: string,
  awayTeam: string,
  matchday: number,
  date: string,
  status: 'scheduled' | 'in-progress' | 'finished' = 'scheduled'
): SeasonMatch => {
  return new SeasonMatch(
    homeTeam,
    awayTeam,
    0,
    0,
    1.5,
    2.0,
    3.0,
    status,
    '2024',
    'L1',
    date,
    matchday
  );
};

// Helper function to create test match results
const createTestMatchResult = (
  match: SeasonMatch,
  bets: any = null,
  scores: any = null
): MatchResult => ({
  match,
  bets,
  scores,
});

// Mock data for testing
const mockMatchesData = {
  incomingMatches: {
    'match1': createTestMatchResult(
      createTestMatch('PSG', 'Marseille', 1, '2024-03-20T20:00:00Z')
    ),
    'match2': createTestMatchResult(
      createTestMatch('Lyon', 'Monaco', 1, '2024-03-20T21:00:00Z')
    ),
    'match3': createTestMatchResult(
      createTestMatch('Nice', 'Lille', 2, '2024-03-21T20:00:00Z')
    ),
    'match4': createTestMatchResult(
      createTestMatch('Bordeaux', 'Toulouse', 2, '2024-03-21T20:00:00Z')
    ),
    'match5': createTestMatchResult(
      createTestMatch('Nantes', 'Rennes', 3, '2024-03-22T19:00:00Z')
    ),
  },
  pastMatches: {
    'match6': createTestMatchResult(
      createTestMatch('Lens', 'Strasbourg', 1, '2024-03-20T19:00:00Z', 'finished')
    ),
  },
};

// Mock player data
const mockPlayer = {
  id: 'player1',
  name: 'Test Player',
  email: 'test@example.com',
};

// Test wrapper component
const TestWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const mockTimeService = new MockTimeService(new Date('2024-03-20T20:10:00'));
  
  return (
    <AuthProvider>
      <TimeServiceProvider service={mockTimeService}>
        {children}
      </TimeServiceProvider>
    </AuthProvider>
  );
};

// Remove the entire describe('GameScreen - Basic Functionality', ...) block and any references to GameScreen 