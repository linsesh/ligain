import { View } from 'react-native';
import { Text } from './Text';
import { useGridCellSize } from '../../hooks/useGridCellSize';
import { colors } from '../../constants/colors';

interface GridTagProps {
  label: string;
  backgroundColor?: string;
  textColor?: string;
}

export function GridTag({ label, backgroundColor = colors.black, textColor = '#ffffff' }: GridTagProps) {
  const cellSize = useGridCellSize();
  return (
    <View style={{ height: cellSize, backgroundColor, justifyContent: 'center', alignItems: 'center', paddingHorizontal: 12 }}>
      <Text className="font-hk-semibold text-sm tracking-wide" style={{ color: textColor }}>
        {label}
      </Text>
    </View>
  );
}
