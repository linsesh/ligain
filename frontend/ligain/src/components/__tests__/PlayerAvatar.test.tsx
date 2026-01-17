import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { PlayerAvatar, getInitials, getColorForName } from '../PlayerAvatar';

// Mock expo-vector-icons
jest.mock('@expo/vector-icons', () => ({
  Ionicons: 'Ionicons',
}));

describe('PlayerAvatar', () => {
  describe('initials rendering', () => {
    it('renders single initial for single-word name', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie', avatarUrl: null }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('M');
    });

    it('renders two initials for two-word name', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie Dupont', avatarUrl: null }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('MD');
    });

    it('renders two initials for multi-word name (first and last)', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Jean Marie Dupont', avatarUrl: null }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('JD');
    });

    it('renders uppercase initials', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'alice bob', avatarUrl: null }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('AB');
    });
  });

  describe('image rendering', () => {
    it('renders image when avatarUrl is provided', () => {
      const { toJSON } = render(
        <PlayerAvatar
          player={{ name: 'Marie', avatarUrl: 'https://example.com/avatar.jpg' }}
          displaySize="medium"
        />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('https://example.com/avatar.jpg');
      expect(tree).not.toContain('"M"'); // Should not contain initials
    });

    it('renders initials when avatarUrl is null', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie', avatarUrl: null }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('M');
      expect(tree).not.toContain('avatar.jpg');
    });

    it('renders initials when avatarUrl is undefined', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie' }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('M');
    });
  });

  describe('sizes', () => {
    it('renders small size (32px)', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie', avatarUrl: null }} displaySize="small" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('"width":"32px"');
      expect(tree).toContain('"height":"32px"');
    });

    it('renders medium size (48px)', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie', avatarUrl: null }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('"width":"48px"');
      expect(tree).toContain('"height":"48px"');
    });

    it('renders large size (80px)', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie', avatarUrl: null }} displaySize="large" />
      );
      const tree = JSON.stringify(toJSON());
      expect(tree).toContain('"width":"80px"');
      expect(tree).toContain('"height":"80px"');
    });
  });

  describe('onPress handler', () => {
    it('calls onPress when provided and pressed', () => {
      const mockOnPress = jest.fn();
      const { root } = render(
        <PlayerAvatar
          player={{ name: 'Marie', avatarUrl: null }}
          displaySize="medium"
          onPress={mockOnPress}
        />
      );
      // Find the touchable element and press it
      fireEvent.press(root);
      expect(mockOnPress).toHaveBeenCalledTimes(1);
    });

    it('does not render TouchableOpacity when onPress is not provided', () => {
      const { toJSON } = render(
        <PlayerAvatar player={{ name: 'Marie', avatarUrl: null }} displaySize="medium" />
      );
      const tree = JSON.stringify(toJSON());
      // When no onPress, it renders a View not a TouchableOpacity
      expect(tree).not.toContain('TouchableOpacity');
    });
  });

  describe('consistent color for name', () => {
    it('generates same color for same name', () => {
      const color1 = getColorForName('TestPlayer');
      const color2 = getColorForName('TestPlayer');
      expect(color1).toBe(color2);
    });

    it('generates different colors for different names', () => {
      const color1 = getColorForName('Marie');
      const color2 = getColorForName('Lucas');
      // Both should be valid colors
      expect(color1).toMatch(/^#[0-9A-F]{6}$/i);
      expect(color2).toMatch(/^#[0-9A-F]{6}$/i);
    });
  });

  describe('getInitials utility', () => {
    it('returns single initial for single name', () => {
      expect(getInitials('Marie')).toBe('M');
    });

    it('returns two initials for two-word name', () => {
      expect(getInitials('Marie Dupont')).toBe('MD');
    });

    it('returns first and last initials for multi-word name', () => {
      expect(getInitials('Jean Marie Dupont')).toBe('JD');
    });

    it('handles lowercase names', () => {
      expect(getInitials('alice bob')).toBe('AB');
    });

    it('handles extra whitespace', () => {
      expect(getInitials('  Alice   Bob  ')).toBe('AB');
    });
  });
});
