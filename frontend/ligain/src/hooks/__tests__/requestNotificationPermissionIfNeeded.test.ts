import { requestNotificationPermissionIfNeeded } from '../useNotifications';

const mockGetPermissionsAsync = jest.fn();
const mockRequestPermissionsAsync = jest.fn();

jest.mock('expo-notifications', () => ({
  setNotificationHandler: jest.fn(),
  getPermissionsAsync: (...args: any[]) => mockGetPermissionsAsync(...args),
  requestPermissionsAsync: (...args: any[]) => mockRequestPermissionsAsync(...args),
}));

const mockGetItem = jest.fn();
const mockSetItem = jest.fn();

jest.mock('../../utils/storage', () => ({
  getItem: (...args: any[]) => mockGetItem(...args),
  setItem: (...args: any[]) => mockSetItem(...args),
}));

describe('requestNotificationPermissionIfNeeded', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockSetItem.mockResolvedValue(undefined);
  });

  it('does nothing when permission was already requested', async () => {
    mockGetItem.mockResolvedValue('true');

    await requestNotificationPermissionIfNeeded();

    expect(mockRequestPermissionsAsync).not.toHaveBeenCalled();
    expect(mockSetItem).not.toHaveBeenCalled();
  });

  it('requests permissions when flag is absent', async () => {
    mockGetItem.mockResolvedValue(null);
    mockGetPermissionsAsync.mockResolvedValue({ status: 'undetermined' });
    mockRequestPermissionsAsync.mockResolvedValue({ status: 'granted' });

    await requestNotificationPermissionIfNeeded();

    expect(mockRequestPermissionsAsync).toHaveBeenCalledTimes(1);
  });

  it('stores permission flag and enables preference when user grants', async () => {
    mockGetItem.mockResolvedValue(null);
    mockGetPermissionsAsync.mockResolvedValue({ status: 'undetermined' });
    mockRequestPermissionsAsync.mockResolvedValue({ status: 'granted' });

    await requestNotificationPermissionIfNeeded();

    expect(mockSetItem).toHaveBeenCalledWith('notification_permission_requested', 'true');
    expect(mockSetItem).toHaveBeenCalledWith('notification_preferences_enabled', 'true');
  });

  it('stores only permission flag when user denies', async () => {
    mockGetItem.mockResolvedValue(null);
    mockGetPermissionsAsync.mockResolvedValue({ status: 'undetermined' });
    mockRequestPermissionsAsync.mockResolvedValue({ status: 'denied' });

    await requestNotificationPermissionIfNeeded();

    expect(mockSetItem).toHaveBeenCalledWith('notification_permission_requested', 'true');
    expect(mockSetItem).not.toHaveBeenCalledWith('notification_preferences_enabled', 'true');
  });

  it('skips requestPermissionsAsync when already granted', async () => {
    mockGetItem.mockResolvedValue(null);
    mockGetPermissionsAsync.mockResolvedValue({ status: 'granted' });

    await requestNotificationPermissionIfNeeded();

    expect(mockRequestPermissionsAsync).not.toHaveBeenCalled();
    expect(mockSetItem).toHaveBeenCalledWith('notification_permission_requested', 'true');
    expect(mockSetItem).toHaveBeenCalledWith('notification_preferences_enabled', 'true');
  });
});
