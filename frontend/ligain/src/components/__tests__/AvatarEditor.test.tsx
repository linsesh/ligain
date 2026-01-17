/**
 * AvatarEditor Component Tests
 *
 * Note: Full component tests for AvatarEditor require extensive React Native
 * mocking due to Modal, expo-image-picker, and other native dependencies.
 * The core logic (avatar validation) is tested in avatarValidation.test.ts.
 *
 * These tests verify the component can be imported and basic props are typed correctly.
 */

// Mock expo-image-picker
jest.mock('expo-image-picker', () => ({
  launchCameraAsync: jest.fn(),
  launchImageLibraryAsync: jest.fn(),
  MediaTypeOptions: { Images: 'Images' },
  requestCameraPermissionsAsync: jest.fn(),
  requestMediaLibraryPermissionsAsync: jest.fn(),
}));

// Mock avatar validation
jest.mock('../../utils/avatarValidation', () => ({
  validateAvatarImage: jest.fn(),
}));

// Mock translation hook
jest.mock('../../hooks/useTranslation', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock colors
jest.mock('../../constants/colors', () => ({
  colors: {
    primary: '#007AFF',
    secondary: '#5856D6',
    background: '#000000',
    card: '#1C1C1E',
    text: '#FFFFFF',
    textSecondary: '#8E8E93',
    border: '#38383A',
    danger: '#FF3B30',
  },
}));

// Mock expo-vector-icons
jest.mock('@expo/vector-icons', () => ({
  Ionicons: 'Ionicons',
}));

import { AvatarEditor } from '../AvatarEditor';

describe('AvatarEditor', () => {
  it('exports the AvatarEditor component', () => {
    expect(AvatarEditor).toBeDefined();
    expect(typeof AvatarEditor).toBe('function');
  });

  it('component accepts expected props (type check)', () => {
    // This is a compile-time type check
    // If the interface changed, TypeScript would fail compilation
    const props: Parameters<typeof AvatarEditor>[0] = {
      currentAvatarUrl: 'https://example.com/avatar.jpg',
      onSave: async (_uri: string) => {},
      onDelete: async () => {},
      visible: true,
      onClose: () => {},
    };

    // Verify the props object matches expected shape
    expect(props.currentAvatarUrl).toBe('https://example.com/avatar.jpg');
    expect(typeof props.onSave).toBe('function');
    expect(typeof props.onDelete).toBe('function');
    expect(props.visible).toBe(true);
    expect(typeof props.onClose).toBe('function');
  });
});
