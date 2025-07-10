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
  Switch,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../src/contexts/AuthContext';
import { AuthService } from '../src/services/authService';
import { colors } from '../src/constants/colors';
import { API_CONFIG } from '../src/config/api';
import { GoogleSignInButton } from '../src/components/GoogleSignInButton';

export default function SignInScreen() {
  console.log('🔐 SignInScreen - Rendering signin screen');
  console.log('🔐 SignInScreen - Platform:', Platform.OS);
  console.log('🔐 SignInScreen - API_CONFIG:', {
    BASE_URL: API_CONFIG.BASE_URL,
    API_KEY: API_CONFIG.API_KEY ? 'configured' : 'NOT_CONFIGURED',
    GAME_ID: API_CONFIG.GAME_ID
  });
  
  const [isLoading, setIsLoading] = useState(false);
  const [showNameModal, setShowNameModal] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<'google' | 'apple' | null>(null);
  const [authResult, setAuthResult] = useState<any>(null);
  const [displayName, setDisplayName] = useState('');
  const [useRealAuth, setUseRealAuth] = useState(false); // Toggle for real OAuth in dev

  const { signIn } = useAuth();
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

  const handleSignIn = async (provider: 'google' | 'apple') => {
    console.log(`🔐 SignInScreen - Starting sign in with ${provider}`);
    try {
      setIsLoading(true);
      setSelectedProvider(provider);
      
      // Determine whether to use real or mock authentication
      const shouldUseRealAuth = !__DEV__ || useRealAuth;
      console.log(`🔐 SignInScreen - Using ${shouldUseRealAuth ? 'real' : 'mock'} authentication`);
      
      let result;
      
      if (shouldUseRealAuth) {
        // Use real authentication
        console.log('🔐 SignInScreen - Calling real authentication');
        if (provider === 'google') {
          result = await AuthService.signInWithGoogle();
        } else {
          result = await AuthService.signInWithApple();
        }
        console.log('🔐 SignInScreen - Real sign in result:', {
          provider: result.provider,
          email: result.email,
          name: result.name,
          token: result.token ? '***token***' : 'NO_TOKEN'
        });
      } else {
        // Use mock authentication for development
        console.log('🔐 SignInScreen - Calling mockSignIn');
        result = await AuthService.mockSignIn(provider);
        console.log('🔐 SignInScreen - Mock sign in result:', {
          provider: result.provider,
          email: result.email,
          name: result.name,
          token: result.token ? '***token***' : 'NO_TOKEN'
        });
      }

      setAuthResult(result);
      setShowNameModal(true);
      console.log('🔐 SignInScreen - Showing name modal');
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
          <Text style={[styles.title, { color: colors.text }]}>Welcome to Ligain</Text>
          <Text style={[styles.subtitle, { color: colors.text }]}>
            Sign in to start betting on football matches
          </Text>
        </View>

        <View style={styles.buttonContainer}>
          {/* Google Sign In */}
          {availableProviders.includes('google') && (
            useRealAuth ? (
              <GoogleSignInButton
                onSignInSuccess={(result) => {
                  console.log('Google Sign-In success:', result);
                  setAuthResult(result);
                  setSelectedProvider('google');
                  setShowNameModal(true);
                }}
                onSignInError={(error) => {
                  console.error('Google Sign-In error:', error);
                  Alert.alert('Sign-In Failed', error.message);
                }}
                disabled={isLoading}
              />
            ) : (
              <TouchableOpacity
                style={[styles.button, styles.googleButton]}
                onPress={() => handleSignIn('google')}
                disabled={isLoading}
              >
                {isLoading ? (
                  <ActivityIndicator color="#4285F4" />
                ) : (
                  <>
                    <Ionicons name="logo-google" size={24} color="#4285F4" />
                    <Text style={styles.googleButtonText}>Continue with Google (Mock)</Text>
                  </>
                )}
              </TouchableOpacity>
            )
          )}

          {/* Apple Sign In (iOS only) */}
          {availableProviders.includes('apple') && (
            <TouchableOpacity
              style={[styles.button, styles.appleButton]}
              onPress={() => handleSignIn('apple')}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="#FFFFFF" />
              ) : (
                <>
                  <Ionicons name="logo-apple" size={24} color="#FFFFFF" />
                  <Text style={styles.appleButtonText}>Continue with Apple {!useRealAuth && '(Mock)'}</Text>
                </>
              )}
            </TouchableOpacity>
          )}

          {/* Development notice */}
          {__DEV__ && (
            <View style={styles.devNotice}>
              <View style={styles.devToggleContainer}>
                <Text style={[styles.devText, { color: colors.text }]}>
                  🧪 Development Mode: {useRealAuth ? 'Real OAuth' : 'Mock Authentication'}
                </Text>
                <View style={styles.toggleContainer}>
                  <Text style={[styles.toggleLabel, { color: colors.textSecondary }]}>
                    Mock
                  </Text>
                  <Switch
                    value={useRealAuth}
                    onValueChange={setUseRealAuth}
                    trackColor={{ false: colors.border, true: colors.link }}
                    thumbColor={useRealAuth ? '#FFFFFF' : '#FFFFFF'}
                  />
                  <Text style={[styles.toggleLabel, { color: colors.textSecondary }]}>
                    Real
                  </Text>
                </View>
              </View>
              <Text style={[styles.devSubtext, { color: colors.textSecondary }]}>
                {useRealAuth 
                  ? '⚠️ Will open Google/Apple sign-in (requires real OAuth setup)'
                  : '✅ Using mock data for development'
                }
              </Text>
            </View>
          )}


        </View>

        <View style={styles.footer}>
          <Text style={[styles.footerText, { color: colors.text }]}>
            By signing in, you agree to our Terms of Service and Privacy Policy
          </Text>
        </View>
      </View>

      {/* Name Selection Modal */}
      <Modal
        visible={showNameModal}
        animationType="slide"
        transparent={true}
        onRequestClose={() => setShowNameModal(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={[styles.modalContent, { backgroundColor: colors.card }]}>
            <Text style={[styles.modalTitle, { color: colors.text }]}>
              Choose Your Display Name
            </Text>
            <Text style={[styles.modalSubtitle, { color: colors.textSecondary }]}>
              This name will be displayed to other players
            </Text>
            
            <TextInput
              style={[styles.nameInput, { 
                backgroundColor: colors.background,
                color: colors.text,
                borderColor: colors.border
              }]}
              placeholder="Enter your display name"
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
                <Text style={[styles.cancelButtonText, { color: colors.text }]}>Cancel</Text>
              </TouchableOpacity>
              
              <TouchableOpacity
                style={[styles.modalButton, styles.continueButton]}
                onPress={handleNameSubmit}
                disabled={isLoading || !displayName.trim()}
              >
                {isLoading ? (
                  <ActivityIndicator color="#FFFFFF" />
                ) : (
                  <Text style={styles.continueButtonText}>Continue</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
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
  devNotice: {
    backgroundColor: '#1a1a1a',
    padding: 12,
    borderRadius: 8,
    marginTop: 16,
    borderWidth: 1,
    borderColor: '#4a9eff',
  },
  devToggleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  devText: {
    fontSize: 12,
    color: '#4a9eff',
    flex: 1,
  },
  toggleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  toggleLabel: {
    fontSize: 10,
    fontWeight: '500',
  },
  devSubtext: {
    fontSize: 10,
    textAlign: 'center',
    lineHeight: 14,
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
}); 