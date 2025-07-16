import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert } from 'react-native';
import { AuthService } from '../services/authService';
import { Ionicons } from '@expo/vector-icons';

interface GoogleSignInButtonProps {
  onSignInSuccess?: (result: any) => void;
  onSignInError?: (error: Error) => void;
  disabled?: boolean;
}

export const GoogleSignInButton: React.FC<GoogleSignInButtonProps> = ({
  onSignInSuccess,
  onSignInError,
  disabled = false,
}) => {
  const handleGoogleSignIn = async () => {
    try {
      const result = await AuthService.signInWithGoogle();
      console.log('Google Sign-In success:', result);
      
      if (onSignInSuccess) {
        onSignInSuccess(result);
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