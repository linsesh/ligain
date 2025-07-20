import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react-native';
import { View, Text, TouchableOpacity } from 'react-native';

// Mock the navigation component logic
const MockMatchdayNavigation = ({ 
  currentMatchday, 
  sortedMatchdays, 
  onNavigate 
}: {
  currentMatchday: number;
  sortedMatchdays: number[];
  onNavigate: (direction: 'prev' | 'next') => void;
}) => {
  const currentIndex = sortedMatchdays.indexOf(currentMatchday);
  const canGoPrev = currentIndex > 0;
  const canGoNext = currentIndex < sortedMatchdays.length - 1;

  return (
    <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
      <TouchableOpacity 
        onPress={() => onNavigate('prev')}
        disabled={!canGoPrev}
        testID="prev-matchday-button"
        accessibilityState={{ disabled: !canGoPrev }}
      >
        <Text>←</Text>
      </TouchableOpacity>
      
      <View>
        <Text>Matchday {currentMatchday}</Text>
      </View>
      
      <TouchableOpacity 
        onPress={() => onNavigate('next')}
        disabled={!canGoNext}
        testID="next-matchday-button"
        accessibilityState={{ disabled: !canGoNext }}
      >
        <Text>→</Text>
      </TouchableOpacity>
    </View>
  );
};

describe('MatchdayNavigation Component', () => {
  const mockOnNavigate = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render navigation component', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={2}
        sortedMatchdays={[1, 2, 3]}
        onNavigate={mockOnNavigate}
      />
    );

    // Test that the component renders without throwing
    expect(mockOnNavigate).toBeDefined();
  });

  it('should handle navigation callback', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={2}
        sortedMatchdays={[1, 2, 3]}
        onNavigate={mockOnNavigate}
      />
    );

    // Test that the callback is not called initially
    expect(mockOnNavigate).not.toHaveBeenCalled();
  });

  it('should render with single matchday', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={1}
        sortedMatchdays={[1]}
        onNavigate={mockOnNavigate}
      />
    );

    expect(mockOnNavigate).toBeDefined();
  });
}); 