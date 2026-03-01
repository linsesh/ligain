import { StyleSheet } from 'react-native';
import Svg, { Defs, Pattern, Rect } from 'react-native-svg';
import { colors } from '../constants/colors';

const CELL_SIZE = 25;

export function GridBackground() {
  return (
    <Svg style={StyleSheet.absoluteFillObject}>
      <Defs>
        <Pattern
          id="grid"
          width={CELL_SIZE}
          height={CELL_SIZE}
          patternUnits="userSpaceOnUse"
        >
          <Rect
            width={CELL_SIZE}
            height={CELL_SIZE}
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
