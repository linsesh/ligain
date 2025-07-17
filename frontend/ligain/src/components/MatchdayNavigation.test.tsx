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

  it('should display current matchday', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={2}
        sortedMatchdays={[1, 2, 3]}
        onNavigate={mockOnNavigate}
      />
    );

    expect(screen.getByText('Matchday 2')).toBeTruthy();
  });

  it('should enable both navigation buttons when in middle', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={2}
        sortedMatchdays={[1, 2, 3]}
        onNavigate={mockOnNavigate}
      />
    );

    const prevButton = screen.getByTestId('prev-matchday-button');
    const nextButton = screen.getByTestId('next-matchday-button');

    expect(prevButton.props.accessibilityState.disabled).toBe(false);
    expect(nextButton.props.accessibilityState.disabled).toBe(false);
  });

  it('should disable prev button on first matchday', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={1}
        sortedMatchdays={[1, 2, 3]}
        onNavigate={mockOnNavigate}
      />
    );

    const prevButton = screen.getByTestId('prev-matchday-button');
    const nextButton = screen.getByTestId('next-matchday-button');

    expect(prevButton.props.accessibilityState.disabled).toBe(true);
    expect(nextButton.props.accessibilityState.disabled).toBe(false);
  });

  it('should disable next button on last matchday', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={3}
        sortedMatchdays={[1, 2, 3]}
        onNavigate={mockOnNavigate}
      />
    );

    const prevButton = screen.getByTestId('prev-matchday-button');
    const nextButton = screen.getByTestId('next-matchday-button');

    expect(prevButton.props.accessibilityState.disabled).toBe(false);
    expect(nextButton.props.accessibilityState.disabled).toBe(true);
  });

  it('should call onNavigate with correct direction when buttons are pressed', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={2}
        sortedMatchdays={[1, 2, 3]}
        onNavigate={mockOnNavigate}
      />
    );

    const prevButton = screen.getByTestId('prev-matchday-button');
    const nextButton = screen.getByTestId('next-matchday-button');

    fireEvent.press(prevButton);
    expect(mockOnNavigate).toHaveBeenCalledWith('prev');

    fireEvent.press(nextButton);
    expect(mockOnNavigate).toHaveBeenCalledWith('next');
  });

  it('should handle single matchday correctly', () => {
    render(
      <MockMatchdayNavigation
        currentMatchday={1}
        sortedMatchdays={[1]}
        onNavigate={mockOnNavigate}
      />
    );

    const prevButton = screen.getByTestId('prev-matchday-button');
    const nextButton = screen.getByTestId('next-matchday-button');

    expect(prevButton.props.accessibilityState.disabled).toBe(true);
    expect(nextButton.props.accessibilityState.disabled).toBe(true);
  });
}); 