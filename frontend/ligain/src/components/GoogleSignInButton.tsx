import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert } from 'react-native';
import { AuthService } from '../services/authService';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'expo-router';
import { useTranslation } from '../hooks/useTranslation';
import { translateError } from '../utils/errorMessages';

interface GoogleSignInButtonProps {
  onSignInSuccess?: (result: any) => void;
  onSignInError?: (error: any) => void;
  onNewUser?: (result: any) => void; // Callback for new users who need to choose display name
  disabled?: boolean;
}

export const GoogleSignInButton: React.FC<GoogleSignInButtonProps> = ({
  onSignInSuccess,
  onSignInError,
  onNewUser,
  disabled = false,
}) => {
  const { signIn } = useAuth();
  const router = useRouter();
  const { t } = useTranslation();

  const handleGoogleSignIn = async () => {
    try {
      const result = await AuthService.signInWithGoogle();
      console.log('Google Sign-In success:', result);
      
      try {
        // Try to authenticate without display name first (two-step flow)
        const authResult = await signIn(
          'google',
          result.token,
          result.email,
          '' // Empty name to trigger display name requirement for new users
        );
        
        // Check if backend is requesting display name
        if (authResult && authResult.needDisplayName) {
          console.log('Google Sign-In - New user detected, showing display name modal');
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
        console.log('Google Sign-In - Authentication failed:', {
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
          console.error('Google Sign-In - Authentication error:', authError.message);
          if (onSignInError) {
            onSignInError(authError);
          }
        }
      }
    } catch (error: any) {
      console.log('Google Sign-In - Caught error:', {
        message: error.message,
        code: error.code,
        stack: error.stack
      });
      
      // Check if this is a cancellation (normal behavior, not an error)
      // Use error codes when available, fallback to message checking
      const isCancellation = 
        error.code === 'SIGN_IN_CANCELLED' || 
        error.code === 'ERR_CANCELED' ||
        (error.message && (
          error.message.toLowerCase().includes('cancelled') || 
          error.message.toLowerCase().includes('canceled') ||
          error.message.toLowerCase().includes('sign-in was cancelled') ||
          error.message.toLowerCase().includes('sign-in was canceled')
        ));
      
      if (isCancellation) {
        console.log('Google Sign-In - User cancelled sign-in (normal behavior)');
        return; // Don't show error alert for cancellation
      }
      
      console.error('Google Sign-In error:', error);
      
      // Let the parent handle the error display
      if (onSignInError) {
        onSignInError(error);
      }
    }
  };

  return (
    <TouchableOpacity 
      style={[styles.button, disabled && styles.buttonDisabled]} 
      onPress={handleGoogleSignIn}
      disabled={disabled}
    >
      <View style={styles.contentRow}>
        <Ionicons name="logo-google" size={24} color="#fff" style={{ marginRight: 10 }} />
        <Text style={styles.buttonText}>{t('auth.continueWithGoogleButton')}</Text>
      </View>
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  button: {
    backgroundColor: '#4285F4',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
    marginVertical: 10,
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  contentRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
}); 