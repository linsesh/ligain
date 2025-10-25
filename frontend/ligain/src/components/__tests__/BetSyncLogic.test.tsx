// Mock colors first - before any imports
jest.mock('../../constants/colors', () => ({
  colors: {
    primary: '#007AFF',
    text: '#000000',
    textSecondary: '#666666',
    background: '#FFFFFF',
    card: '#F2F2F7',
    border: '#C6C6C8',
    success: '#34C759',
    error: '#FF3B30',
    warning: '#FF9500',
    disabled: '#E5E5EA',
    loadingBackground: '#F2F2F7',
  }
}));

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import { BetSyncModal, SyncResult } from '../../components/BetSyncModal';
import enTranslations from '../../i18n/locales/en.json';

// Mock the translation hook to use actual translations
jest.mock('../../hooks/useTranslation', () => ({
  useTranslation: () => ({
    t: (key: string, params?: any) => {
      // Simple translation function that handles interpolation
      let translation = key;
      const keys = key.split('.');
      let current: any = enTranslations;
      
      for (const k of keys) {
        if (current && typeof current === 'object' && k in current) {
          current = current[k];
        } else {
          return key; // Return key if translation not found
        }
      }
      
      if (typeof current === 'string') {
        translation = current;
        
        // Handle interpolation
        if (params) {
          Object.keys(params).forEach(param => {
            translation = translation.replace(new RegExp(`{{${param}}}`, 'g'), params[param]);
          });
        }
      }
      
      return translation;
    },
  }),
}));

// Mock react-native components
jest.mock('react-native', () => {
  const RN = jest.requireActual('react-native');
  return {
    ...RN,
    Modal: ({ children, visible }: any) => visible ? children : null,
    View: 'View',
    Text: 'Text',
    TouchableOpacity: 'TouchableOpacity',
    ScrollView: 'ScrollView',
    useWindowDimensions: () => ({ width: 400, height: 800 }),
  };
});

// Mock Ionicons
jest.mock('@expo/vector-icons', () => ({
  Ionicons: 'Ionicons',
}));


describe('BetSyncModal', () => {
  const mockSyncOpportunity = {
    sourceGameId: 'game-2',
    sourceGameName: 'Game 2',
    matchesToSync: [
      {
        matchId: 'match-1',
        homeTeam: 'Team A',
        awayTeam: 'Team B',
        matchday: 1,
        predictedHomeGoals: 2,
        predictedAwayGoals: 1
      },
      {
        matchId: 'match-2',
        homeTeam: 'Team C',
        awayTeam: 'Team D',
        matchday: 2,
        predictedHomeGoals: 1,
        predictedAwayGoals: 0
      }
    ]
  };

  const mockOnSynchronize = jest.fn();
  const mockOnNotNow = jest.fn();
  const mockOnRetryFailed = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render initial modal with sync opportunity', () => {
    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="initial"
      />
    );

    expect(screen.getByText('Synchronize Bets?')).toBeTruthy();
    expect(screen.getByText('Synchronize')).toBeTruthy();
    expect(screen.getByText('Not now')).toBeTruthy();
  });

  it('should render success modal', () => {
    const syncResult: SyncResult = {
      successful: mockSyncOpportunity.matchesToSync,
      failed: []
    };

    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="success"
        syncResult={syncResult}
      />
    );

    expect(screen.getByText('Synchronization Complete')).toBeTruthy();
    expect(screen.getByText('All 2 bets have been successfully synchronized.')).toBeTruthy();
    expect(screen.getByText('Close')).toBeTruthy();
  });

  it('should render partial success modal with retry option', () => {
    const syncResult: SyncResult = {
      successful: [mockSyncOpportunity.matchesToSync[0]],
      failed: [{
        match: mockSyncOpportunity.matchesToSync[1],
        error: new Error('Network error')
      }]
    };

    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        onRetryFailed={mockOnRetryFailed}
        mode="partialSuccess"
        syncResult={syncResult}
      />
    );

    expect(screen.getByText('Partial Synchronization')).toBeTruthy();
    expect(screen.getByText('1 out of 2 bets were synchronized successfully.')).toBeTruthy();
    expect(screen.getByText('Successfully synchronized:')).toBeTruthy();
    expect(screen.getByText('Failed to synchronize:')).toBeTruthy();
    expect(screen.getByText('Team A vs Team B')).toBeTruthy();
    expect(screen.getByText('Team C vs Team D')).toBeTruthy();
    expect(screen.getByText('Retry Failed')).toBeTruthy();
    expect(screen.getByText('Close')).toBeTruthy();
  });

  it('should render failure modal', () => {
    const syncResult: SyncResult = {
      successful: [],
      failed: [{
        match: mockSyncOpportunity.matchesToSync[0],
        error: new Error('Network error')
      }]
    };

    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="failure"
        syncResult={syncResult}
      />
    );

    expect(screen.getByText('Synchronization Failed')).toBeTruthy();
    expect(screen.getByText('Failed to synchronize any bets. Please try again.')).toBeTruthy();
    expect(screen.getByText('Close')).toBeTruthy();
    expect(screen.queryByText('Retry Failed')).toBeNull();
  });

  it('should call onSynchronize when synchronize button is clicked', () => {
    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="initial"
      />
    );

    fireEvent.press(screen.getByText('Synchronize'));
    expect(mockOnSynchronize).toHaveBeenCalledTimes(1);
  });

  it('should call onNotNow when not now button is clicked', () => {
    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="initial"
      />
    );

    fireEvent.press(screen.getByText('Not now'));
    expect(mockOnNotNow).toHaveBeenCalledTimes(1);
  });

  it('should call onRetryFailed when retry button is clicked', () => {
    const syncResult: SyncResult = {
      successful: [mockSyncOpportunity.matchesToSync[0]],
      failed: [{
        match: mockSyncOpportunity.matchesToSync[1],
        error: new Error('Network error')
      }]
    };

    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        onRetryFailed={mockOnRetryFailed}
        mode="partialSuccess"
        syncResult={syncResult}
      />
    );

    fireEvent.press(screen.getByText('Retry Failed'));
    expect(mockOnRetryFailed).toHaveBeenCalledTimes(1);
  });

  it('should show loading state when loading prop is true', () => {
    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="initial"
        loading={true}
      />
    );

    expect(screen.getByText('Loading...')).toBeTruthy();
    expect(screen.getByText('Synchronize')).toBeDisabled();
  });

  it('should not render when visible is false', () => {
    render(
      <BetSyncModal
        visible={false}
        syncOpportunity={mockSyncOpportunity}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="initial"
      />
    );

    expect(screen.queryByText('Synchronize Bets?')).toBeNull();
  });

  it('should not render when syncOpportunity is null', () => {
    render(
      <BetSyncModal
        visible={true}
        syncOpportunity={null}
        onSynchronize={mockOnSynchronize}
        onNotNow={mockOnNotNow}
        mode="initial"
      />
    );

    expect(screen.queryByText('Synchronize Bets?')).toBeNull();
  });
});
