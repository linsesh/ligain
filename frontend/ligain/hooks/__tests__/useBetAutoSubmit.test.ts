import { renderHook, act } from '@testing-library/react-native';
import { useBetAutoSubmit } from '../useBetAutoSubmit';

describe('useBetAutoSubmit', () => {
  let onSubmit: jest.Mock;

  beforeEach(() => {
    onSubmit = jest.fn();
  });

  it('does not call onSubmit on initial render even when both fields are pre-filled', () => {
    renderHook(() => useBetAutoSubmit('2', '1', onSubmit));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('does not call onSubmit when only home is filled', () => {
    const { rerender } = renderHook(
      ({ home, away }: { home: string; away: string }) =>
        useBetAutoSubmit(home, away, onSubmit),
      { initialProps: { home: '', away: '' } },
    );
    act(() => { rerender({ home: '2', away: '' }); });
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('does not call onSubmit when only away is filled', () => {
    const { rerender } = renderHook(
      ({ home, away }: { home: string; away: string }) =>
        useBetAutoSubmit(home, away, onSubmit),
      { initialProps: { home: '', away: '' } },
    );
    act(() => { rerender({ home: '', away: '1' }); });
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('calls onSubmit with correct numbers when both fields become non-empty', () => {
    const { rerender } = renderHook(
      ({ home, away }: { home: string; away: string }) =>
        useBetAutoSubmit(home, away, onSubmit),
      { initialProps: { home: '', away: '' } },
    );
    act(() => { rerender({ home: '2', away: '' }); });
    act(() => { rerender({ home: '2', away: '1' }); });
    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith(2, 1);
  });

  it('calls onSubmit again when user edits one field while the other is already filled', () => {
    const { rerender } = renderHook(
      ({ home, away }: { home: string; away: string }) =>
        useBetAutoSubmit(home, away, onSubmit),
      { initialProps: { home: '', away: '' } },
    );
    act(() => { rerender({ home: '2', away: '1' }); });
    act(() => { rerender({ home: '3', away: '1' }); });
    expect(onSubmit).toHaveBeenCalledTimes(2);
    expect(onSubmit).toHaveBeenLastCalledWith(3, 1);
  });

  it('does not call onSubmit when a field is cleared back to empty', () => {
    const { rerender } = renderHook(
      ({ home, away }: { home: string; away: string }) =>
        useBetAutoSubmit(home, away, onSubmit),
      { initialProps: { home: '', away: '' } },
    );
    act(() => { rerender({ home: '2', away: '1' }); });
    onSubmit.mockClear();
    act(() => { rerender({ home: '', away: '1' }); });
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('does not call onSubmit for negative values', () => {
    const { rerender } = renderHook(
      ({ home, away }: { home: string; away: string }) =>
        useBetAutoSubmit(home, away, onSubmit),
      { initialProps: { home: '', away: '' } },
    );
    act(() => { rerender({ home: '-1', away: '1' }); });
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('does not call onSubmit for non-numeric values', () => {
    const { rerender } = renderHook(
      ({ home, away }: { home: string; away: string }) =>
        useBetAutoSubmit(home, away, onSubmit),
      { initialProps: { home: '', away: '' } },
    );
    act(() => { rerender({ home: 'abc', away: '1' }); });
    expect(onSubmit).not.toHaveBeenCalled();
  });
});
