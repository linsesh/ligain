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
import { API_CONFIG, getApiHeaders } from '../src/config/api';
import { GoogleSignInButton } from '../src/components/GoogleSignInButton';
import { PrivacyTermsModal } from '../src/components/PrivacyTermsModal';
import { setItem } from '../src/utils/storage';
import { useTranslation } from 'react-i18next';
import { translateError } from '../src/utils/errorMessages';

export default function SignInScreen() {
  const { signIn, player, setPlayer, showNameModal, setShowNameModal, authResult, setAuthResult, selectedProvider, setSelectedProvider } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [displayName, setDisplayName] = useState('');
  const [showPrivacyTermsModal, setShowPrivacyTermsModal] = useState(false);
  const [showGuestPasswordModal, setShowGuestPasswordModal] = useState(false);
  const [guestPassword, setGuestPassword] = useState('');
  const { t } = useTranslation();
  
  const router = useRouter();

  // Get available providers for current platform
  const availableProviders = AuthService.getAvailableProviders();

  // Test provider availability in development
  React.useEffect(() => {
    if (__DEV__) {
      console.log('🔐 SignInScreen - Testing provider availability in dev mode');
      AuthService.testProviderAvailability();
    }
  }, []);

  // Store pending auth info for two-step flow
  const [pendingAuth, setPendingAuth] = useState<{ provider: string, token: string, email: string } | null>(null);

  // OAuth sign-in handler
  const handleOAuthSignIn = async (provider: 'google' | 'apple', token: string, email: string, displayName?: string) => {
    setIsLoading(true);
    try {
      const result = await signIn(provider, token, email, displayName || '');
      
      // Check if backend is requesting display name
      if (result && result.needDisplayName) {
        setShowNameModal(true);
        setDisplayName(result.suggestedName || '');
        setAuthResult({ provider, token, email });
        // Don't show error alert for display name requirement - it's part of normal flow
        return;
      }

      // Success: navigate to main app
      router.replace('/(tabs)');
    } catch (error: any) {
      console.error('OAuth sign-in error:', error);
      
      // Check if this is a display name requirement (should not show as error)
      if (error.message && error.message.includes('display name is required for new users')) {
        console.log('OAuth sign-in - Display name required (normal flow)');
        return;
      }
      
      Alert.alert(t('errors.signInFailed'), translateError(error.message || 'Unknown error'));
    } finally {
      setIsLoading(false);
    }
  };

  // Google sign-in button handler
  const handleGoogleSignIn = async () => {
    // You need to implement the logic to get token and email from Google
    // For now, this is a placeholder
    // Example:
    // const { token, email } = await AuthService.signInWithGoogle();
    // await handleOAuthSignIn('google', token, email);
  };

  // Apple sign-in button handler
  const handleAppleSignIn = async () => {
    // You need to implement the logic to get token and email from Apple
    // For now, this is a placeholder
    // Example:
    // const { token, email } = await AuthService.signInWithApple();
    // await handleOAuthSignIn('apple', token, email);
  };

  // Guest sign-in handler
  const handleGuestSignIn = async () => {
    setShowGuestPasswordModal(true);
  };

  // Handle guest password submission
  const handleGuestPasswordSubmit = async () => {
    const correctPassword = 'H1F03ogLxPtf5mG';
    
    if (guestPassword !== correctPassword) {
      Alert.alert(t('errors.error'), t('auth.incorrectPassword'));
      return;
    }

    setIsLoading(true);
    try {
      const result = await AuthService.signInAsGuest('Player1');
      await setItem('auth_token', result.token);
      await setItem('player_data', JSON.stringify({ id: result.playerId, name: result.name }));
      setPlayer({ id: result.playerId || 'guest-player', name: result.name || 'Guest Player' });
      setShowGuestPasswordModal(false);
      setGuestPassword('');
      router.replace('/(tabs)');
    } catch (error: any) {
      console.error('Guest sign-in error:', error);
      Alert.alert(t('errors.signInFailed'), translateError(error.message || 'Unknown error'));
    } finally {
      setIsLoading(false);
    }
  };

  const handleNameSubmit = async () => {
    console.log('🔐 handleNameSubmit - Starting submission');
    console.log('🔐 handleNameSubmit - displayName:', displayName);
    console.log('�� handleNameSubmit - authResult:', authResult);
    
    if (!displayName.trim()) {
      Alert.alert(t('errors.error'), t('auth.pleaseEnterDisplayName'));
      return;
    }
    if (displayName.trim().length < 2) {
      Alert.alert(t('errors.error'), t('auth.displayNameTooShort'));
      return;
    }
    if (displayName.trim().length > 20) {
      Alert.alert(t('errors.error'), t('auth.displayNameTooLong'));
      return;
    }
    if (!authResult) {
      console.log('🔐 handleNameSubmit - ERROR: authResult is null');
      Alert.alert(t('errors.error'), t('errors.missingAuthContext'));
      setShowNameModal(false);
      return;
    }
    
    try {
      console.log('🔐 handleNameSubmit - Calling handleOAuthSignIn with:', {
        provider: authResult.provider,
        email: authResult.email,
        displayName: displayName.trim()
      });
      
      // Retry sign-in with display name
      await handleOAuthSignIn(
        authResult.provider as 'google' | 'apple',
        authResult.token,
        authResult.email,
        displayName.trim()
      );
      
      // If successful, clean up
      setShowNameModal(false);
      setAuthResult(null);
      setDisplayName('');
    } catch (error) {
      console.error('Display name submission error:', error);
      // Keep modal open on error so user can try again
    }
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.loadingBackground }]}>
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
                console.log('🔐 GoogleSignInButton onNewUser - Called with result:', result);
                console.log('🔐 GoogleSignInButton onNewUser - Setting authResult to:', {
                  provider: result.provider,
                  token: result.token ? `${result.token.substring(0, 10)}...` : 'null',
                  email: result.email
                });
                
                setAuthResult(result);
                setSelectedProvider('google');
                setShowNameModal(true);
                
                console.log('🔐 GoogleSignInButton onNewUser - Modal should now be visible');
              }}
              onSignInError={(error: any) => {
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
                Alert.alert(t('errors.signInFailed'), translateError(error.message));
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
                    await handleOAuthSignIn(
                      'apple',
                      result.token,
                      result.email,
                      '' // Never use Apple name, always require user to choose
                    );
                    
                    // If successful, navigate to main app (existing user)
                    console.log('Apple Sign-In - Existing user authenticated successfully');
                  } catch (authError: any) {
                    console.log('Apple Sign-In - Authentication failed:', authError.message);
                    
                    // Check if this is a "display name required" error for new users (two-step flow)
                    if (authError.message && authError.message.startsWith('NEED_DISPLAY_NAME:')) {
                      console.log('Apple Sign-In - New user detected, showing display name modal');
                      setAuthResult({
                        provider: result.provider,
                        token: result.token,
                        email: result.email
                      });
                      setSelectedProvider('apple');
                      setShowNameModal(true);
                      console.log('Apple Sign-In - Showing name modal for user');
                    } else if (authError.message && authError.message.includes('display name is required for new users')) {
                      // Fallback for old error format
                      console.log('Apple Sign-In - New user detected (legacy), showing display name modal');
                      setAuthResult({
                        provider: result.provider,
                        token: result.token,
                        email: result.email
                      });
                      setSelectedProvider('apple');
                      setShowNameModal(true);
                      console.log('Apple Sign-In - Showing name modal for user');
                    } else {
                      // For other authentication errors, show error alert
                      console.error('Apple Sign-In - Authentication error:', authError.message);
                      Alert.alert(t('errors.authenticationError'), translateError(authError.message));
                    }
                  }
                } catch (error: any) {
                  // Check if this is a cancellation (normal behavior, not an error)
                  if (error.message && (
                    error.message.includes('cancelled') || 
                    error.message.includes('canceled') ||
                    error.message.includes('Sign-in was cancelled') ||
                    error.message.includes('Sign-in was canceled')
                  )) {
                    console.log('Apple Sign-In - User cancelled sign-in (normal behavior)');
                    return; // Don't show error alert for cancellation
                  }
                  
                  console.error('Apple Sign-In error:', error);
                  Alert.alert(t('errors.signInFailed'), translateError(error.message));
                }
              }}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color={colors.primary} />
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
                onPress={handleGuestSignIn}
                disabled={isLoading}
              >
                {isLoading ? (
                  <ActivityIndicator color={colors.primary} />
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
                  <ActivityIndicator color={colors.primary} />
                ) : (
                  <Text style={styles.continueButtonText}>{t('common.continue')}</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>

      {/* Guest Password Modal */}
      <Modal
        visible={showGuestPasswordModal}
        animationType="slide"
        transparent={true}
        onRequestClose={() => {
          setShowGuestPasswordModal(false);
          setGuestPassword('');
        }}
      >
        <View style={styles.modalOverlay}>
          <View style={[styles.modalContent, { backgroundColor: colors.card }]}>
            <Text style={[styles.modalTitle, { color: colors.text }]}>
              {t('auth.guestPasswordTitle')}
            </Text>
            <Text style={[styles.modalSubtitle, { color: colors.textSecondary }]}>
              {t('auth.guestPasswordSubtitle')}
            </Text>
            
            <TextInput
              style={[styles.nameInput, { 
                backgroundColor: colors.background,
                color: colors.text,
                borderColor: colors.border
              }]}
              placeholder={t('auth.enterPassword')}
              placeholderTextColor={colors.textSecondary}
              value={guestPassword}
              onChangeText={setGuestPassword}
              autoFocus={true}
              secureTextEntry={true}
              autoCapitalize="none"
            />
            
            <View style={styles.modalButtons}>
              <TouchableOpacity
                style={[styles.modalButton, styles.cancelButton]}
                onPress={() => {
                  setShowGuestPasswordModal(false);
                  setGuestPassword('');
                }}
                disabled={isLoading}
              >
                <Text style={[styles.cancelButtonText, { color: colors.text }]}>{t('common.cancel')}</Text>
              </TouchableOpacity>
              
              <TouchableOpacity
                style={[styles.modalButton, styles.continueButton]}
                onPress={handleGuestPasswordSubmit}
                disabled={isLoading || !guestPassword.trim()}
              >
                {isLoading ? (
                  <ActivityIndicator color={colors.primary} />
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