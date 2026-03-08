import { View } from 'react-native';
import { Text } from './ui/Text';
import { useGridCellSize } from '../hooks/useGridCellSize';
import { colors } from '../constants/colors';

function formatSeasonShort(seasonYear: string): string {
  const parts = seasonYear.split(/[-\/]/);
  if (parts.length === 2) return `${parts[0].slice(-2)}/${parts[1].slice(-2)}`;
  return seasonYear;
}

interface SeasonBannerProps {
  seasonYear: string;
  competitionName: string;
  className?: string;
}

export function SeasonBanner({ seasonYear, competitionName, className }: SeasonBannerProps) {
  const cellSize = useGridCellSize();
  const label = `${formatSeasonShort(seasonYear)} · ${competitionName}`;

  return (
    <View style={{ width: '100%', alignItems: 'center' }} className={className}>
    <View
      style={{ height: cellSize, backgroundColor: colors.black, justifyContent: 'center', alignItems: 'center', width: cellSize * 7 }}
    >
      <Text className="font-hk-semibold text-white text-sm tracking-wide">
        {label}
      </Text>
    </View>
    </View>
  );
}
