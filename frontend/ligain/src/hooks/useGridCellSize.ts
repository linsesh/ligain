import { useWindowDimensions } from 'react-native';
import { computeGridCellSize } from '../utils/gridSize';

export function useGridCellSize(): number {
  const { width } = useWindowDimensions();
  return computeGridCellSize(width);
}
