import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert } from 'react-native';
import { AuthService } from '../services/authService';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../contexts/AuthContext';
import { useRouter } from 'expo-router';

interface GoogleSignInButtonProps {
  onSignInSuccess?: (result: any) => void;
  onSignInError?: (error: Error) => void;
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

  const handleGoogleSignIn = async () => {
    try {
      const result = await AuthService.signInWithGoogle();
      console.log('Google Sign-In success:', result);
      
      // Always try to authenticate without a display name first
      // This will fail for new users, triggering the display name modal
      try {
        console.log('Google Sign-In - Attempting authentication without display name');
        await signIn(
          'google',
          result.token,
          result.email,
          '' // Never use Google name, always require user to choose
        );
        
        // If successful, navigate to main app (existing user)
        console.log('Google Sign-In - Existing user authenticated successfully');
        router.replace('/(tabs)');
        
        if (onSignInSuccess) {
          onSignInSuccess(result);
        }
      } catch (authError: any) {
        console.log('Google Sign-In - Authentication failed:', authError.message);
        
        // Check if this is a "display name required" error for new users
        if (authError.message === 'display name is required for new users') {
          console.log('Google Sign-In - New user detected, showing display name modal');
          if (onNewUser) {
            onNewUser(result);
          } else {
            // Re-throw if no callback is provided
            throw authError;
          }
        } else {
          // For other authentication errors, show error alert
          console.error('Google Sign-In - Authentication error:', authError.message);
          Alert.alert('Authentication Error', authError.message);
          if (onSignInError) {
            onSignInError(authError);
          }
        }
      }
    } catch (error: any) {
      console.error('Google Sign-In error:', error);
      
      Alert.alert('Sign-In Error', error.message);
      
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
        <Text style={styles.buttonText}>Continue with Google</Text>
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