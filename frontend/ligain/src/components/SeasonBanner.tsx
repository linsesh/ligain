import { View } from 'react-native';
import { Text } from './ui/Text';
import { GRID_CELL_SIZE } from './GridBackground';

function formatSeasonShort(seasonYear: string): string {
  const parts = seasonYear.split('-');
  if (parts.length === 2) return `${parts[0].slice(-2)}/${parts[1].slice(-2)}`;
  return seasonYear;
}

interface SeasonBannerProps {
  seasonYear: string;
  competitionName: string;
  className?: string;
}

export function SeasonBanner({ seasonYear, competitionName, className }: SeasonBannerProps) {
  const label = `${formatSeasonShort(seasonYear)} · ${competitionName}`;

  return (
    <View
      style={{ height: GRID_CELL_SIZE * 2, backgroundColor: '#000', justifyContent: 'center', alignItems: 'center', width: '100%' }}
      className={className}
    >
      <Text className="font-hk-semibold text-white text-sm tracking-wide">
        {label}
      </Text>
    </View>
  );
}
