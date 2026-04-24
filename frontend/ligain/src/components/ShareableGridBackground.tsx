import React from 'react';
import { View, StyleSheet } from 'react-native';
import { colors } from '../constants/colors';

const GRID_COLUMNS = 15;

interface ShareableGridBackgroundProps {
  width?: number;
  height: number;
}

export function ShareableGridBackground({ width = 1080, height }: ShareableGridBackgroundProps) {
  const cellSize = width / GRID_COLUMNS;
  const rows = Math.ceil(height / cellSize);

  return (
    <View style={[StyleSheet.absoluteFillObject, { backgroundColor: colors.background }]}>
      {Array.from({ length: rows }, (_, row) => (
        <View key={row} style={{ flexDirection: 'row', height: cellSize }}>
          {Array.from({ length: GRID_COLUMNS }, (_, col) => (
            <View
              key={col}
              style={{
                width: cellSize,
                height: cellSize,
                borderRightWidth: 0.5,
                borderBottomWidth: 0.5,
                borderColor: colors.gridLine,
              }}
            />
          ))}
        </View>
      ))}
    </View>
  );
}
