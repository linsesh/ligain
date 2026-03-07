import { StyleSheet } from 'react-native';
import Svg, { Defs, Pattern, Rect } from 'react-native-svg';
import { colors } from '../constants/colors';
import { useGridCellSize } from '../hooks/useGridCellSize';

export function GridBackground() {
  const cellSize = useGridCellSize();

  return (
    <Svg style={StyleSheet.absoluteFillObject}>
      <Defs>
        <Pattern
          id="grid"
          width={cellSize}
          height={cellSize}
          patternUnits="userSpaceOnUse"
        >
          <Rect
            width={cellSize}
            height={cellSize}
            fill={colors.background}
            stroke={colors.gridLine}
            strokeWidth={0.5}
          />
        </Pattern>
      </Defs>
      <Rect width="100%" height="100%" fill={`url(#grid)`} />
    </Svg>
  );
}
