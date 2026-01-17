import React from 'react';
import { View, Text, Image, TouchableOpacity, StyleSheet } from 'react-native';

interface PlayerAvatarProps {
  player: { name: string; avatarUrl?: string | null };
  displaySize: 'small' | 'medium' | 'large';
  onPress?: () => void;
}

const SIZES = {
  small: { container: 32, fontSize: 12 },
  medium: { container: 48, fontSize: 18 },
  large: { container: 80, fontSize: 28 },
};

// Color palette for avatar backgrounds
const AVATAR_COLORS = [
  '#E57373', // Red
  '#64B5F6', // Blue
  '#81C784', // Green
  '#FFD54F', // Yellow
  '#BA68C8', // Purple
  '#4DB6AC', // Teal
  '#FF8A65', // Orange
  '#90A4AE', // Blue Grey
  '#F06292', // Pink
  '#7986CB', // Indigo
];

/**
 * Generate a consistent color based on player name
 */
export function getColorForName(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
}

/**
 * Generate initials from player name
 */
export function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) {
    return parts[0].charAt(0).toUpperCase();
  }
  // First and last name initials
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
}

export function PlayerAvatar({ player, displaySize, onPress }: PlayerAvatarProps) {
  const size = SIZES[displaySize];
  const hasAvatar = player.avatarUrl && player.avatarUrl.length > 0;
  const backgroundColor = getColorForName(player.name);

  const containerStyle = [
    styles.container,
    {
      width: size.container,
      height: size.container,
      borderRadius: size.container / 2,
      backgroundColor: hasAvatar ? 'transparent' : backgroundColor,
    },
  ];

  const initials = getInitials(player.name);
  const content = hasAvatar ? (
    <Image
      source={{ uri: player.avatarUrl! }}
      style={[
        styles.image,
        {
          width: size.container,
          height: size.container,
          borderRadius: size.container / 2,
        },
      ]}
      testID="avatar-image"
    />
  ) : (
    <Text style={[styles.initials, { fontSize: size.fontSize }]} testID="avatar-initials">
      {initials}
    </Text>
  );

  if (onPress) {
    return (
      <TouchableOpacity
        onPress={onPress}
        testID="avatar-touchable"
        activeOpacity={0.7}
      >
        <View style={containerStyle} testID="avatar-container">
          {content}
        </View>
      </TouchableOpacity>
    );
  }

  return (
    <View style={containerStyle} testID="avatar-container">
      {content}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    justifyContent: 'center',
    alignItems: 'center',
    overflow: 'hidden',
  },
  initials: {
    color: '#FFFFFF',
    fontWeight: '600',
  },
  image: {
    resizeMode: 'cover',
  },
});
