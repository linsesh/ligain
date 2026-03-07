export const GRID_COLUMNS = 15;

/**
 * Computes a grid cell size that divides the screen width evenly into `columns` columns.
 * screenWidth / columns is exact — no partial cells at the edges.
 */
export function computeGridCellSize(screenWidth: number, columns: number = GRID_COLUMNS): number {
  return screenWidth / columns;
}
