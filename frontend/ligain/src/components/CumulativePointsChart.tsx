import React, { useMemo, useState, useCallback } from 'react';
import { View, Text } from 'react-native';
import Svg, { Line, Polyline, Text as SvgText } from 'react-native-svg';

export interface Series {
  playerId: string;
  playerName: string;
  values: number[]; // cumulative points per matchday index
}

interface CumulativePointsChartProps {
  matchdays: number[];
  series: Series[];
  width?: number;
  height?: number;
}

const COLOR_PALETTE: { name: string; hex: string }[] = [
  { name: 'Coral Red', hex: '#FF6B6B' },
  { name: 'Teal Mint', hex: '#4ECDC4' },
  { name: 'Sunflower', hex: '#FFD93D' },
  { name: 'Indigo', hex: '#6C5CE7' },
  { name: 'Sky Blue', hex: '#45B7D1' },
  { name: 'Rose', hex: '#F94D6A' },
  { name: 'Emerald', hex: '#2ECC71' },
  { name: 'Orange', hex: '#E67E22' },
  { name: 'Sea Green', hex: '#1ABC9C' },
  { name: 'Amethyst', hex: '#9B59B6' },
  { name: 'Fuchsia', hex: '#E84393' },
  { name: 'Aqua', hex: '#00CEC9' },
  { name: 'Mustard', hex: '#FDCB6E' },
  { name: 'Ocean', hex: '#0984E3' },
  { name: 'Mint', hex: '#55EFC4' },
];

export default function CumulativePointsChart({ matchdays, series, width = 0, height = 160 }: CumulativePointsChartProps) {
  const [containerWidth, setContainerWidth] = useState<number>(0);
  const onLayout = useCallback((e: any) => {
    try {
      const w = e?.nativeEvent?.layout?.width || 0;
      if (w && Math.abs(w - containerWidth) > 1) setContainerWidth(w);
    } catch (error) {
      console.warn('CumulativePointsChart onLayout error:', error);
    }
  }, [containerWidth]);

  // Use measured width if not provided
  const svgWidth = width > 0 ? width : Math.max(200, containerWidth);

  // Separate paddings to avoid label clipping
  const paddingLeft = 36;
  const paddingRight = 12;
  const paddingTop = 28;
  const paddingBottom = 24;
  const plotWidth = svgWidth - paddingLeft - paddingRight;
  const plotHeight = height - paddingTop - paddingBottom;

  const maxY = useMemo(() => {
    let maxVal = 0;
    for (const s of series) {
      for (const v of s.values) maxVal = Math.max(maxVal, v);
    }
    return maxVal || 1;
  }, [series]);

  const xForIndex = (i: number) => {
    if (matchdays.length <= 1) return paddingLeft;
    return paddingLeft + (i * plotWidth) / (matchdays.length - 1);
  };
  const yForValue = (v: number) => paddingTop + plotHeight - (v / maxY) * plotHeight;

  // Compute x-axis ticks to avoid overcrowding
  const maxTicks = 8;
  const xTickIndices = useMemo(() => {
    const n = matchdays.length;
    if (n === 0) return [] as number[];
    if (n <= maxTicks) return Array.from({ length: n }, (_, i) => i);
    const step = Math.ceil(n / maxTicks);
    const bases: number[] = [];
    for (let i = 0; i < n; i += step) bases.push(i);
    if (bases[bases.length - 1] !== n - 1) bases.push(n - 1);
    return bases;
  }, [matchdays.length]);

  // Add safety checks for data
  if (!matchdays || matchdays.length === 0 || !series || series.length === 0) {
    return (
      <View style={{ height, justifyContent: 'center', alignItems: 'center' }}>
        <Text style={{ color: '#999', fontSize: 14 }}>No data available</Text>
      </View>
    );
  }

  return (
    <View onLayout={onLayout} style={{ backgroundColor: 'transparent' }}>
      <Svg width={svgWidth} height={height}>
        {/* These two lines are the axes lines*/}
        <Line x1={paddingLeft} y1={paddingTop} x2={paddingLeft} y2={height - paddingBottom} stroke="#555" strokeWidth={1} />
        <Line x1={paddingLeft} y1={height - paddingBottom} x2={svgWidth - paddingRight} y2={height - paddingBottom} stroke="#555" strokeWidth={1} />

        {/* X labels (matchdays) */}
        {xTickIndices.map((i) => (
          <SvgText key={`x-${matchdays[i]}`} x={xForIndex(i)} y={height - paddingBottom + 14} fill="#aaa" fontSize="10" textAnchor="middle">
            {String(matchdays[i])}
          </SvgText>
        ))}

        {/* Optional vertical grid lines at ticks */}
        {xTickIndices.map((i) => (
          <Line
            key={`grid-${i}`}
            x1={xForIndex(i)}
            y1={paddingTop}
            x2={xForIndex(i)}
            y2={height - paddingBottom}
            stroke="#444"
            strokeWidth={1}
          />
        ))}

        {/* Y labels (0 and max) */}
        <SvgText x={paddingLeft - 8} y={height - paddingBottom} fill="#aaa" fontSize="10" textAnchor="end">0</SvgText>
        <SvgText x={paddingLeft - 8} y={paddingTop + 6} fill="#aaa" fontSize="10" textAnchor="end">{maxY}</SvgText>

        {series.map((s, idx) => {
          const { hex: color } = COLOR_PALETTE[idx % COLOR_PALETTE.length];
          const points = s.values.map((v, i) => `${xForIndex(i)},${yForValue(v)}`).join(' ');
          return (
            <Polyline
              key={s.playerId}
              points={points}
              fill="none"
              stroke={color}
              strokeWidth={2}
            />
          );
        })}
      </Svg>
      {/* Legend mapping colors to player names */}
      <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginTop: 8 }}>
        {series.map((s, idx) => {
          const { hex: color } = COLOR_PALETTE[idx % COLOR_PALETTE.length];
          return (
            <View key={`legend-${s.playerId}`} style={{ flexDirection: 'row', alignItems: 'center', marginRight: 12, marginBottom: 6 }}>
              <View style={{ width: 10, height: 10, borderRadius: 5, backgroundColor: color, marginRight: 6 }} />
              <Text style={{ color: '#fff', fontSize: 12 }}>{s.playerName || s.playerId}</Text>
            </View>
          );
        })}
      </View>
    </View>
  );
}


