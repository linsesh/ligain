import { computeGridCellSize, GRID_COLUMNS } from './gridSize';

describe('computeGridCellSize', () => {
  it('returns exactly screenWidth / GRID_COLUMNS by default', () => {
    expect(computeGridCellSize(375)).toBe(375 / GRID_COLUMNS);
  });

  it('produces 25pt cells on a 375pt screen (iPhone SE)', () => {
    expect(computeGridCellSize(375)).toBe(25);
  });

  it('produces 26pt cells on a 390pt screen (iPhone 14)', () => {
    expect(computeGridCellSize(390)).toBe(26);
  });

  it('always fills the full width with no remainder', () => {
    const widths = [375, 390, 393, 428, 360, 412];
    widths.forEach((w) => {
      const size = computeGridCellSize(w);
      expect(size * GRID_COLUMNS).toBeCloseTo(w, 10);
    });
  });

  it('supports a custom column count', () => {
    expect(computeGridCellSize(300, 10)).toBe(30);
    expect(computeGridCellSize(400, 16)).toBe(25);
  });

  it('always returns a positive cell size', () => {
    expect(computeGridCellSize(320)).toBeGreaterThan(0);
    expect(computeGridCellSize(428)).toBeGreaterThan(0);
  });
});
