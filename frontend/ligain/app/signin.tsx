import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  Platform,
  TextInput,
  Modal,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../src/contexts/AuthContext';
import { AuthService } from '../src/services/authService';
import { colors } from '../src/constants/colors';
import { API_CONFIG } from '../src/config/api';
import { GoogleSignInButton } from '../src/components/GoogleSignInButton';
import { PrivacyTermsModal } from '../src/components/PrivacyTermsModal';
import { setItem } from '../src/utils/storage';
import { useTranslation } from 'react-i18next';

export default function SignInScreen() {
  console.log('🔐 SignInScreen - Rendering signin screen');
  console.log('🔐 SignInScreen - Platform:', Platform.OS);
  console.log('🔐 SignInScreen - API_CONFIG:', {
    BASE_URL: API_CONFIG.BASE_URL,
    API_KEY: API_CONFIG.API_KEY ? 'configured' : 'NOT_CONFIGURED'
  });
  
  const [isLoading, setIsLoading] = useState(false);
  const [displayName, setDisplayName] = useState('');
  const [showPrivacyTermsModal, setShowPrivacyTermsModal] = useState(false);
  const { t } = useTranslation();
  
  const { signIn, player, setPlayer, showNameModal, setShowNameModal, authResult, setAuthResult, selectedProvider, setSelectedProvider } = useAuth();
  const router = useRouter();

  // Get available providers for current platform
  const availableProviders = AuthService.getAvailableProviders();
  console.log('🔐 SignInScreen - Available providers:', availableProviders);

  // Test provider availability in development
  React.useEffect(() => {
    if (__DEV__) {
      console.log('🔐 SignInScreen - Testing provider availability in dev mode');
      AuthService.testProviderAvailability();
    }
  }, []);

  // Debug modal state
  React.useEffect(() => {
    console.log('🔐 SignInScreen - Modal state changed - showNameModal:', showNameModal);
  }, [showNameModal]);

  const handleSignIn = async (provider: 'google' | 'apple' | 'guest') => {
    console.log(`🔐 SignInScreen - Starting sign in with ${provider}`);
    try {
      setIsLoading(true);
      setSelectedProvider(provider);
      
      let result;
      
      if (provider === 'guest') {
        // Guest authentication - always use real backend
        console.log('🔐 SignInScreen - Calling guest authentication');
        result = await AuthService.signInAsGuest('Player1');
        console.log('🔐 SignInScreen - Guest sign in result:', {
          provider: result.provider,
          email: result.email,
          name: result.name,
          token: result.token ? '***token***' : 'NO_TOKEN'
        });
        
        // For guest authentication, skip the name modal and sign in directly
        console.log('🔐 SignInScreen - Guest authentication, signing in directly');
        
        // Store the result directly since AuthService.signInAsGuest already made the API call
        console.log('🔐 SignInScreen - Storing guest authentication result');
        await setItem('auth_token', result.token);
        await setItem('player_data', JSON.stringify({ 
          id: result.playerId, // Use the actual player ID from backend
          name: result.name 
        }));
        
        // Update the auth context state directly to avoid double API call
        setPlayer({ 
          id: result.playerId || 'guest-player', // Use the actual player ID from backend
          name: result.name || 'Guest Player'
        });
        
        // Navigate to main app
        console.log('🔐 SignInScreen - Navigating to main app');
        router.replace('/(tabs)');
      } else {
        // OAuth authentication - handled by individual buttons
        console.log('🔐 SignInScreen - OAuth authentication should be handled by individual buttons');
        throw new Error('OAuth authentication should be handled by individual buttons');
      }
    } catch (error) {
      console.error('🔐 SignInScreen - Sign in error:', error);
      console.error('🔐 SignInScreen - Error details:', {
        name: error instanceof Error ? error.name : 'Unknown',
        message: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : 'No stack trace'
      });
      
      Alert.alert(
        'Sign In Failed',
        error instanceof Error ? error.message : 'An unexpected error occurred'
      );
    } finally {
      setIsLoading(false);
      console.log('🔐 SignInScreen - Sign in process completed');
    }
  };

  const handleNameSubmit = async () => {
    console.log('🔐 SignInScreen - Starting name submission');
    console.log('🔐 SignInScreen - Display name:', displayName);
    console.log('🔐 SignInScreen - Selected provider:', selectedProvider);
    console.log('🔐 SignInScreen - Auth result exists:', !!authResult);
    
    if (!displayName.trim()) {
      Alert.alert('Error', 'Please enter a display name');
      return;
    }

    if (displayName.trim().length < 2) {
      Alert.alert('Error', 'Display name must be at least 2 characters long');
      return;
    }

    if (displayName.trim().length > 20) {
      Alert.alert('Error', 'Display name must be 20 characters or less');
      return;
    }

    try {
      setIsLoading(true);
      
      console.log('🔐 SignInScreen - Calling signIn with context');
      console.log('🔐 SignInScreen - Sign in parameters:', {
        provider: selectedProvider,
        email: authResult?.email,
        name: displayName.trim(),
        token: authResult?.token ? '***token***' : 'NO_TOKEN'
      });
      
      if (!authResult?.token || !authResult?.email) {
        throw new Error('Missing authentication data');
      }
      
      await signIn(
        selectedProvider!,
        authResult.token,
        authResult.email,
        displayName.trim()
      );

      console.log('🔐 SignInScreen - Sign in successful, cleaning up');
      setShowNameModal(false);
      setDisplayName('');
      setAuthResult(null);
      setSelectedProvider(null);

      // Navigate to main app
      console.log('🔐 SignInScreen - Navigating to main app');
      router.replace('/(tabs)');
    } catch (error) {
      console.error('🔐 SignInScreen - Sign in error in handleNameSubmit:', error);
      console.error('🔐 SignInScreen - Error details:', {
        name: error instanceof Error ? error.name : 'Unknown',
        message: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : 'No stack trace'
      });
      
      Alert.alert(
        'Sign In Failed',
        error instanceof Error ? error.message : 'An unexpected error occurred'
      );
    } finally {
      setIsLoading(false);
      console.log('🔐 SignInScreen - Name submission completed');
    }
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      <View style={styles.content}>
        <View style={styles.header}>
          <Text style={[styles.title, { color: colors.text }]}>{t('auth.welcome')}</Text>
          <Text style={[styles.subtitle, { color: colors.text }]}>
            {t('auth.signInSubtitle')}
          </Text>
        </View>

        <View style={styles.buttonContainer}>
          {/* Google Sign In */}
          {availableProviders.includes('google') && (
            <GoogleSignInButton
              onSignInSuccess={(result) => {
                console.log('Google Sign-In success for existing user:', result);
              }}
              onNewUser={(result) => {
                console.log('Google Sign-In - User needs to choose display name:', result);
                console.log('🔐 SignInScreen - Setting modal state to true');
                setAuthResult(result);
                setSelectedProvider('google');
                setShowNameModal(true);
                console.log('🔐 SignInScreen - Modal state should now be true');
              }}
              onSignInError={(error) => {
                console.error('Google Sign-In error:', error);
                Alert.alert('Sign-In Failed', error.message);
              }}
              disabled={isLoading}
            />
          )}

          {/* Apple Sign In (iOS only) */}
          {availableProviders.includes('apple') && (
            <TouchableOpacity
              style={[styles.button, styles.appleButton]}
              onPress={async () => {
                try {
                  const result = await AuthService.signInWithApple();
                  console.log('Apple Sign-In success:', result);
                  
                  // Always try to authenticate without a display name first
                  // This will fail for new users, triggering the display name modal
                  try {
                    console.log('Apple Sign-In - Attempting authentication without display name');
                    await signIn(
                      'apple',
                      result.token,
                      result.email,
                      '' // Never use Apple name, always require user to choose
                    );
                    
                    // If successful, navigate to main app (existing user)
                    console.log('Apple Sign-In - Existing user authenticated successfully');
                    router.replace('/(tabs)');
                  } catch (authError: any) {
                    console.log('Apple Sign-In - Authentication failed:', authError.message);
                    
                    // Check if this is a "display name required" error for new users
                    if (authError.message === 'display name is required for new users') {
                      console.log('Apple Sign-In - New user detected, showing display name modal');
                      setAuthResult(result);
                      setSelectedProvider('apple');
                      setShowNameModal(true);
                      console.log('Apple Sign-In - Showing name modal for user');
                    } else {
                      // For other authentication errors, show error alert
                      console.error('Apple Sign-In - Authentication error:', authError.message);
                      Alert.alert('Authentication Error', authError.message);
                    }
                  }
                } catch (error: any) {
                  console.error('Apple Sign-In error:', error);
                  Alert.alert('Sign-In Failed', error.message);
                }
              }}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="#FFFFFF" />
              ) : (
                <>
                  <Ionicons name="logo-apple" size={24} color="#FFFFFF" />
                  <Text style={styles.appleButtonText}>{t('auth.continueWithApple')}</Text>
                </>
              )}
            </TouchableOpacity>
          )}
        </View>
        {/* Separation between main sign-in and guest */}
        {availableProviders.includes('guest') && (
          <View style={styles.guestSectionWrapper}>
            <View style={styles.guestDivider} />
            <View style={styles.guestSection}>
              <TouchableOpacity
                style={[styles.button, styles.guestButton]}
                onPress={() => handleSignIn('guest')}
                disabled={isLoading}
              >
                {isLoading ? (
                  <ActivityIndicator color="#333" />
                ) : (
                  <>
                    <Ionicons name="person" size={24} color="#333" />
                    <Text style={styles.guestButtonText}>{t('auth.continueAsGuest')}</Text>
                  </>
                )}
              </TouchableOpacity>
              <Text style={styles.guestNote}>
                {t('auth.guestNote')}
              </Text>
            </View>
          </View>
        )}
        <View style={styles.footer}>
          <TouchableOpacity onPress={() => setShowPrivacyTermsModal(true)}>
            <Text style={[styles.footerText, { color: colors.link }]}>
              {t('auth.termsAgreement')}
            </Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Name Selection Modal */}
      <Modal
        visible={showNameModal}
        animationType="slide"
        transparent={true}
        onRequestClose={() => {
          setShowNameModal(false);
        }}
      >
        <View style={styles.modalOverlay}>
          <View style={[styles.modalContent, { backgroundColor: colors.card }]}>
            <Text style={[styles.modalTitle, { color: colors.text }]}>
              {t('auth.chooseName')}
            </Text>
            <Text style={[styles.modalSubtitle, { color: colors.textSecondary }]}>
              {t('auth.chooseNameSubtitle')}
            </Text>
            
            <TextInput
              style={[styles.nameInput, { 
                backgroundColor: colors.background,
                color: colors.text,
                borderColor: colors.border
              }]}
              placeholder={t('auth.enterDisplayName')}
              placeholderTextColor={colors.textSecondary}
              value={displayName}
              onChangeText={setDisplayName}
              autoFocus={true}
              maxLength={20}
              autoCapitalize="words"
            />
            
            <View style={styles.modalButtons}>
              <TouchableOpacity
                style={[styles.modalButton, styles.cancelButton]}
                onPress={() => {
                  setShowNameModal(false);
                  setDisplayName('');
                  setAuthResult(null);
                  setSelectedProvider(null);
                }}
                disabled={isLoading}
              >
                <Text style={[styles.cancelButtonText, { color: colors.text }]}>{t('common.cancel')}</Text>
              </TouchableOpacity>
              
              <TouchableOpacity
                style={[styles.modalButton, styles.continueButton]}
                onPress={handleNameSubmit}
                disabled={isLoading || !displayName.trim()}
              >
                {isLoading ? (
                  <ActivityIndicator color="#FFFFFF" />
                ) : (
                  <Text style={styles.continueButtonText}>{t('common.continue')}</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>

      {/* Privacy and Terms Modal */}
      <PrivacyTermsModal
        visible={showPrivacyTermsModal}
        onClose={() => setShowPrivacyTermsModal(false)}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: 24,
  },
  header: {
    alignItems: 'center',
    marginBottom: 48,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    marginBottom: 8,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 16,
    textAlign: 'center',
    opacity: 0.8,
  },
  buttonContainer: {
    marginBottom: 32,
  },
  button: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    paddingHorizontal: 24,
    borderRadius: 12,
    marginBottom: 16,
    minHeight: 56,
  },
  googleButton: {
    backgroundColor: '#FFFFFF',
    borderWidth: 1,
    borderColor: '#E0E0E0',
  },
  googleButtonText: {
    marginLeft: 12,
    fontSize: 16,
    fontWeight: '600',
    color: '#333333',
  },
  appleButton: {
    backgroundColor: '#000000',
  },
  appleButtonText: {
    marginLeft: 12,
    fontSize: 16,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  guestButton: {
    backgroundColor: '#f0f0f0',
    borderWidth: 1,
    borderColor: '#cccccc',
    marginBottom: 0,
  },
  guestButtonText: {
    marginLeft: 12,
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  guestNote: {
    fontSize: 12,
    color: '#666666',
    textAlign: 'center',
    marginTop: 4,
    marginBottom: 16,
    fontStyle: 'italic',
  },
  footer: {
    alignItems: 'center',
  },
  footerText: {
    fontSize: 12,
    textAlign: 'center',
    opacity: 0.6,
    lineHeight: 18,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'flex-start',
    alignItems: 'center',
    padding: 20,
    paddingTop: 100, // Move modal higher on screen with more space
  },
  modalContent: {
    width: '100%',
    maxWidth: 400,
    borderRadius: 16,
    padding: 24,
    alignItems: 'center',
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 8,
    textAlign: 'center',
  },
  modalSubtitle: {
    fontSize: 14,
    textAlign: 'center',
    marginBottom: 24,
    lineHeight: 20,
  },
  nameInput: {
    width: '100%',
    height: 48,
    borderWidth: 1,
    borderRadius: 8,
    paddingHorizontal: 16,
    fontSize: 16,
    marginBottom: 24,
  },
  modalButtons: {
    flexDirection: 'row',
    width: '100%',
    gap: 12,
  },
  modalButton: {
    flex: 1,
    height: 48,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  cancelButton: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: '#666666',
  },
  cancelButtonText: {
    fontSize: 16,
    fontWeight: '600',
  },
  continueButton: {
    backgroundColor: '#4a9eff',
  },
  continueButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  guestSection: {
    alignItems: 'center',
    marginTop: 0,
  },
  guestSectionWrapper: {
    marginTop: 24,
    marginBottom: 8,
  },
  guestDivider: {
    height: 1,
    backgroundColor: '#e0e0e0',
    marginBottom: 16,
    marginHorizontal: -24,
  },
}); 