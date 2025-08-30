import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Platform } from 'react-native';
import { AuthService } from '../services/authService';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'expo-router';
import { useTranslation } from '../hooks/useTranslation';

interface AppleSignInButtonProps {
  onSignInSuccess?: (result: any) => void;
  onSignInError?: (error: any) => void;
  onNewUser?: (result: any) => void; // Callback for new users who need to choose display name
  disabled?: boolean;
}

export const AppleSignInButton: React.FC<AppleSignInButtonProps> = ({
  onSignInSuccess,
  onSignInError,
  onNewUser,
  disabled = false,
}) => {
  const { signIn } = useAuth();
  const router = useRouter();
  const { t } = useTranslation();

  const handleAppleSignIn = async () => {
    if (Platform.OS !== 'ios') {
      console.error('Apple Sign-In is only available on iOS');
      return;
    }

    try {
      const result = await AuthService.signInWithApple();
      console.log('Apple Sign-In success:', result);
      
      try {
        // Try to authenticate without display name first (two-step flow)
        const authResult = await signIn(
          'apple',
          result.token,
          result.email,
          '' // Empty name to trigger display name requirement for new users
        );
        
        // Check if backend is requesting display name
        if (authResult && authResult.needDisplayName) {
          console.log('Apple Sign-In - New user detected, showing display name modal');
          if (onNewUser) {
            onNewUser(result);
          } else {
            // Re-throw if no callback is provided
            throw new Error('Display name required but no callback provided');
          }
          return;
        }
        
        // If successful, navigate to main app (existing user)
        router.replace('/(tabs)');
        
        if (onSignInSuccess) {
          onSignInSuccess(result);
        }
      } catch (authError: any) {
        console.log('Apple Sign-In - Authentication failed:', {
          message: authError.message,
          code: authError.code,
          stack: authError.stack
        });
        
        // Check if this is a "display name required" error for new users (two-step flow)
        if (authError.message && authError.message.startsWith('NEED_DISPLAY_NAME:')) {
          if (onNewUser) {
            onNewUser(result);
          } else {
            // Re-throw if no callback is provided
            throw authError;
          }
        } else if (authError.message && authError.message.includes('display name is required for new users')) {
          // Fallback for old error format
          if (onNewUser) {
            onNewUser(result);
          } else {
            // Re-throw if no callback is provided
            throw authError;
          }
        } else {
          // For other authentication errors, let the parent handle the error
          console.error('Apple Sign-In - Authentication error:', authError.message);
          if (onSignInError) {
            onSignInError(authError);
          }
        }
      }
    } catch (error: any) {
      console.log('Apple Sign-In - Caught error:', {
        message: error.message,
        code: error.code,
        stack: error.stack
      });
      
      // Check if this is a cancellation (normal behavior, not an error)
      const isCancellation = error.message && (
        error.message.includes('cancelled') || 
        error.message.includes('canceled') ||
        error.message.includes('Sign-in was cancelled') ||
        error.message.includes('Sign-in was canceled')
      );
      
      if (isCancellation) {
        console.log('Apple Sign-In - User cancelled sign-in (normal behavior)');
        return; // Don't show error alert for cancellation
      }
      
      console.error('Apple Sign-In error:', error);
      
      // Let the parent handle the error display
      if (onSignInError) {
        onSignInError(error);
      }
    }
  };

  // Only render on iOS
  if (Platform.OS !== 'ios') {
    return null;
  }

  return (
    <TouchableOpacity 
      style={[styles.button, disabled && styles.buttonDisabled]} 
      onPress={handleAppleSignIn}
      disabled={disabled}
    >
      <View style={styles.contentRow}>
        <Ionicons name="logo-apple" size={24} color="#FFFFFF" style={{ marginRight: 10 }} />
        <Text style={styles.buttonText}>{t('auth.continueWithApple')}</Text>
      </View>
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  button: {
    backgroundColor: '#000000',
    paddingVertical: 16,
    paddingHorizontal: 24,
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 16,
    minHeight: 56,
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  contentRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
});
