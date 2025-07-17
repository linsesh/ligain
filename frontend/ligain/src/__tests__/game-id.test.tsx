import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import { useLocalSearchParams } from 'expo-router';
import { SeasonMatch, MatchResult } from '../../../../src/types/match';
import { MockTimeService } from '../../../../src/services/timeService';
import { TimeServiceProvider } from '../../../../src/contexts/TimeServiceContext';
import { AuthProvider } from '../../../../src/contexts/AuthContext';
import GameScreen from './[id]';

// Mock expo-router
jest.mock('expo-router', () => ({
  useLocalSearchParams: jest.fn(),
  useRouter: jest.fn(() => ({
    push: jest.fn(),
  })),
}));

// Mock the hooks
jest.mock('../../../../hooks/useBetSubmission', () => ({
  useBetSubmission: jest.fn(() => ({
    submitBet: jest.fn(),
    error: null,
  })),
}));

// Mock fetch
global.fetch = jest.fn();

// Mock the API config
jest.mock('../../../../src/config/api', () => ({
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

describe('GameScreen - Matchday Grouping and Time Organization', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useLocalSearchParams as jest.Mock).mockReturnValue({ id: 'game1' });
    
    // Mock successful API response
    (fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockMatchesData),
    });
  });

  describe('Data Transformation Tests', () => {
    it('should group matches by matchday correctly', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Should show matchday 1 initially (first matchday)
        expect(screen.getByText('Matchday 1')).toBeTruthy();
      });
    });

    it('should sort matchdays in ascending order', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Should start with matchday 1
        expect(screen.getByText('Matchday 1')).toBeTruthy();
      });

      // Navigate to next matchday
      const nextButton = screen.getByTestId('next-matchday-button');
      fireEvent.press(nextButton);

      await waitFor(() => {
        expect(screen.getByText('Matchday 2')).toBeTruthy();
      });

      // Navigate to next matchday
      fireEvent.press(nextButton);

      await waitFor(() => {
        expect(screen.getByText('Matchday 3')).toBeTruthy();
      });
    });

    it('should group matches by time within each matchday', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Matchday 1 has matches at 19:00, 20:00, and 21:00
        expect(screen.getByText('19:00')).toBeTruthy();
        expect(screen.getByText('20:00')).toBeTruthy();
        expect(screen.getByText('21:00')).toBeTruthy();
      });
    });

    it('should sort matches by time in ascending order', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        const timeHeaders = screen.getAllByText(/^\d{2}:\d{2}$/);
        const times = timeHeaders.map((header: any) => header.props.children);
        
        // Should be sorted: 19:00, 20:00, 21:00
        expect(times).toEqual(['19:00', '20:00', '21:00']);
      });
    });

    it('should display match times in 24-hour format', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Should show times in 24-hour format
        expect(screen.getByText('20:00')).toBeTruthy();
        expect(screen.getByText('21:00')).toBeTruthy();
      });
    });
  });

  describe('Navigation Tests', () => {
    it('should navigate to next matchday when next button is pressed', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        expect(screen.getByText('Matchday 1')).toBeTruthy();
      });

      const nextButton = screen.getByTestId('next-matchday-button');
      fireEvent.press(nextButton);

      await waitFor(() => {
        expect(screen.getByText('Matchday 2')).toBeTruthy();
      });
    });

    it('should navigate to previous matchday when prev button is pressed', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      // First navigate to matchday 2
      await waitFor(() => {
        expect(screen.getByText('Matchday 1')).toBeTruthy();
      });

      const nextButton = screen.getByTestId('next-matchday-button');
      fireEvent.press(nextButton);

      await waitFor(() => {
        expect(screen.getByText('Matchday 2')).toBeTruthy();
      });

      // Then navigate back to matchday 1
      const prevButton = screen.getByTestId('prev-matchday-button');
      fireEvent.press(prevButton);

      await waitFor(() => {
        expect(screen.getByText('Matchday 1')).toBeTruthy();
      });
    });

    it('should disable prev button on first matchday', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        const prevButton = screen.getByTestId('prev-matchday-button');
        expect(prevButton.props.accessibilityState.disabled).toBe(true);
      });
    });

    it('should disable next button on last matchday', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      // Navigate to last matchday (matchday 3)
      await waitFor(() => {
        expect(screen.getByText('Matchday 1')).toBeTruthy();
      });

      const nextButton = screen.getByTestId('next-matchday-button');
      fireEvent.press(nextButton); // Go to matchday 2
      fireEvent.press(nextButton); // Go to matchday 3

      await waitFor(() => {
        expect(screen.getByText('Matchday 3')).toBeTruthy();
        expect(nextButton.props.accessibilityState.disabled).toBe(true);
      });
    });
  });

  describe('Time Grouping Tests', () => {
    it('should group matches with the same time together', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Matchday 2 has two matches at 20:00
        const nextButton = screen.getByTestId('next-matchday-button');
        fireEvent.press(nextButton);
      });

      await waitFor(() => {
        expect(screen.getByText('Matchday 2')).toBeTruthy();
        expect(screen.getByText('20:00')).toBeTruthy();
        
        // Should show both matches at 20:00
        expect(screen.getByText('Nice')).toBeTruthy();
        expect(screen.getByText('Bordeaux')).toBeTruthy();
      });
    });

    it('should display correct number of matches per time slot', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Matchday 1: 1 match at 19:00, 1 at 20:00, 1 at 21:00
        expect(screen.getByText('19:00')).toBeTruthy();
        expect(screen.getByText('20:00')).toBeTruthy();
        expect(screen.getByText('21:00')).toBeTruthy();
      });
    });
  });

  describe('Date Formatting Tests', () => {
    it('should display matchday date correctly', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Should show the date of the first match in the matchday
        expect(screen.getByText(/Wednesday, March 20/)).toBeTruthy();
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty matchday gracefully', async () => {
      const emptyMatchesData = {
        incomingMatches: {},
        pastMatches: {},
      };

      (fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(emptyMatchesData),
      });

      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        expect(screen.getByText('No matches available')).toBeTruthy();
      });
    });

    it('should handle single matchday correctly', async () => {
      const singleMatchdayData = {
        incomingMatches: {
          'match1': createTestMatchResult(
            createTestMatch('PSG', 'Marseille', 1, '2024-03-20T20:00:00Z')
          ),
        },
        pastMatches: {},
      };

      (fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(singleMatchdayData),
      });

      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        expect(screen.getByText('Matchday 1')).toBeTruthy();
        // Both navigation buttons should be disabled
        const prevButton = screen.getByTestId('prev-matchday-button');
        const nextButton = screen.getByTestId('next-matchday-button');
        expect(prevButton.props.accessibilityState.disabled).toBe(true);
        expect(nextButton.props.accessibilityState.disabled).toBe(true);
      });
    });

    it('should handle matches with same time but different matchdays', async () => {
      const sameTimeData = {
        incomingMatches: {
          'match1': createTestMatchResult(
            createTestMatch('PSG', 'Marseille', 1, '2024-03-20T20:00:00Z')
          ),
          'match2': createTestMatchResult(
            createTestMatch('Lyon', 'Monaco', 2, '2024-03-21T20:00:00Z')
          ),
        },
        pastMatches: {},
      };

      (fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(sameTimeData),
      });

      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Should show matchday 1 initially
        expect(screen.getByText('Matchday 1')).toBeTruthy();
        expect(screen.getByText('20:00')).toBeTruthy();
        expect(screen.getByText('PSG')).toBeTruthy();
      });

      // Navigate to matchday 2
      const nextButton = screen.getByTestId('next-matchday-button');
      fireEvent.press(nextButton);

      await waitFor(() => {
        expect(screen.getByText('Matchday 2')).toBeTruthy();
        expect(screen.getByText('20:00')).toBeTruthy();
        expect(screen.getByText('Lyon')).toBeTruthy();
      });
    });
  });

  describe('Integration with Existing Features', () => {
    it('should maintain betting functionality with new grouping', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Should still show betting inputs for future matches
        const inputs = screen.getAllByRole('textbox');
        expect(inputs.length).toBeGreaterThan(0);
      });
    });

    it('should maintain match card styling with new grouping', async () => {
      render(
        <TestWrapper>
          <GameScreen />
        </TestWrapper>
      );

      await waitFor(() => {
        // Should show match cards with proper styling
        expect(screen.getByText('PSG')).toBeTruthy();
        expect(screen.getByText('Marseille')).toBeTruthy();
      });
    });
  });
}); 