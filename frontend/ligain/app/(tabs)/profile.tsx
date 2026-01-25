import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  ScrollView,
  Modal,
  TextInput,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  TouchableWithoutFeedback,
  Keyboard,
  Switch,
  Linking,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../../src/contexts/AuthContext';
import { useGames } from '../../src/contexts/GamesContext';
import { colors } from '../../src/constants/colors';
import { useTranslation } from '../../src/hooks/useTranslation';
import { formatShortDate } from '../../src/utils/dateUtils';
import { API_CONFIG, getApiHeaders, getAuthenticatedHeaders, authenticatedFetch } from '../../src/config/api';
import { mapPlayerFromBackend } from '../../src/api';
import { useNotifications } from '../../src/hooks/useNotifications';
import { PlayerAvatar } from '../../src/components/PlayerAvatar';
import { AvatarEditor } from '../../src/components/AvatarEditor';

export default function ProfileScreen() {
  const { player, signOut, setPlayer, uploadAvatar, deleteAvatar } = useAuth();
  const { refresh: refreshGames } = useGames();
  const { t, isFrench } = useTranslation();
  // Notification preferences management
  // This hook provides permission status and toggle functionality
  const { preferences, setNotificationEnabled, requestPermissions } = useNotifications();
  const [showDisplayNameModal, setShowDisplayNameModal] = useState(false);
  const [newDisplayName, setNewDisplayName] = useState('');
  const [isUpdating, setIsUpdating] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showDeleteConfirmModal, setShowDeleteConfirmModal] = useState(false);
  const [deleteConfirmText, setDeleteConfirmText] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const [isTogglingNotifications, setIsTogglingNotifications] = useState(false);
  const [showAvatarEditor, setShowAvatarEditor] = useState(false);
  const nameInputRef = React.useRef<TextInput>(null);
  const deleteInputRef = React.useRef<TextInput>(null);

  const handleEditDisplayName = () => {
    setNewDisplayName(player?.name || '');
    setShowDisplayNameModal(true);
  };

  const handleUpdateDisplayName = async () => {
    if (!newDisplayName.trim()) {
      Alert.alert(t('errors.error'), t('auth.pleaseEnterDisplayName'));
      return;
    }

    if (newDisplayName.trim().length < 2) {
      Alert.alert(t('errors.error'), t('auth.displayNameTooShort'));
      return;
    }

    if (newDisplayName.trim().length > 20) {
      Alert.alert(t('errors.error'), t('auth.displayNameTooLong'));
      return;
    }

    if (newDisplayName.trim() === player?.name) {
      setShowDisplayNameModal(false);
      return;
    }

    try {
      setIsUpdating(true);
      
      const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/players/me/display-name`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          displayName: newDisplayName.trim()
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to update display name');
      }


      // Update the auth context with the new player data
      if (setPlayer) {
        setPlayer(mapPlayerFromBackend(data.player));
      }

      // Refresh games data to update player names in leaderboards and match lists
      refreshGames();

      setShowDisplayNameModal(false);
      setNewDisplayName('');
    
    } catch (error) {
      console.error('Error updating display name:', error);
      const errorMessage = error instanceof Error ? error.message : 'An unknown error occurred';
      Alert.alert(t('errors.error'), errorMessage || t('profile.displayNameUpdateFailed'));
    } finally {
      setIsUpdating(false);
    }
  };

  const handleSignOut = () => {
    Alert.alert(
      t('common.signOut'),
      t('auth.signOutConfirm'),
      [
        { text: t('common.cancel'), style: 'cancel' },
        {
          text: t('common.signOut'),
          style: 'destructive',
          onPress: async () => {
            try {
              await signOut();
            } catch (error) {
              Alert.alert(t('common.error'), t('auth.signOutFailed'));
            }
          },
        },
      ]
    );
  };

  const handleDeleteAccount = () => {
    // Check if this is a test account (guest account or test pattern)
    const isTestAccount = !player?.email && !player?.provider;
    
    if (isTestAccount) {
      Alert.alert(
        t('profile.deleteAccountTestAccount'),
        t('profile.deleteAccountTestAccountMessage'),
        [{ text: t('common.ok'), style: 'default' }]
      );
      return;
    }
    
    setShowDeleteModal(true);
  };

  const handleDeleteAccountConfirm = () => {
    setShowDeleteModal(false);
    setShowDeleteConfirmModal(true);
  };

  const handleDeleteAccountFinal = async () => {
    const expectedText = isFrench ? 'SUPPRIMER' : 'DELETE';
    
    if (deleteConfirmText.trim().toUpperCase() !== expectedText) {
      Alert.alert(t('errors.error'), `${t('profile.typeDeleteToConfirm')}`);
      return;
    }

    try {
      setIsDeleting(true);
      
      const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/auth/account`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to delete account');
      }

      // Account deleted successfully - sign out the user
      Alert.alert(
        t('common.success'),
        t('profile.deleteAccountSuccess'),
        [
          {
            text: t('common.ok'),
            onPress: async () => {
              try {
                await signOut();
              } catch (error) {
                // Even if signOut fails, the account is deleted, so we can ignore this error
                console.warn('Failed to sign out after account deletion:', error);
              }
            },
          },
        ]
      );
      
      setShowDeleteConfirmModal(false);
      setDeleteConfirmText('');
    
    } catch (error) {
      console.error('Error deleting account:', error);
      const errorMessage = error instanceof Error ? error.message : 'An unknown error occurred';
      Alert.alert(t('errors.error'), errorMessage || t('profile.deleteAccountFailed'));
    } finally {
      setIsDeleting(false);
    }
  };

  const handleDeleteModalCancel = () => {
    setShowDeleteModal(false);
  };

  const handleDeleteConfirmModalCancel = () => {
    setShowDeleteConfirmModal(false);
    setDeleteConfirmText('');
    if (deleteInputRef.current) deleteInputRef.current.blur();
  };

  const handleModalSave = () => {
    if (nameInputRef.current) nameInputRef.current.blur();
    setTimeout(() => handleUpdateDisplayName(), 50);
  };
  const handleModalCancel = () => {
    if (nameInputRef.current) nameInputRef.current.blur();
    setTimeout(() => {
      setShowDisplayNameModal(false);
      setNewDisplayName('');
    }, 50);
  };

  const handleNotificationToggle = async (value: boolean) => {
    setIsTogglingNotifications(true);
    try {
      if (value) {
        const granted = await requestPermissions();
        if (granted) {
          await setNotificationEnabled(true);
        } else {
          // Permission denied: Show alert with option to open settings
          Alert.alert(
            t('notifications.permissionDenied'),
            t('notifications.permissionDeniedMessage'),
            [
              { text: t('common.cancel'), style: 'cancel' },
              {
                text: t('notifications.openSettings'),
                onPress: () => {
                  Linking.openSettings();
                },
              },
            ]
          );
        }
      } else {
        await setNotificationEnabled(false);
      }
    } catch (error) {
      console.error('Error toggling notifications:', error);
      Alert.alert(t('errors.error'), t('notifications.toggleFailed'));
    } finally {
      setIsTogglingNotifications(false);
    }
  };

  if (!player) {
    return (
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        <Text style={[styles.errorText, { color: colors.text }]}>
          {t('profile.noUserInfo')}
        </Text>
      </View>
    );
  }

  return (
    <>
      <ScrollView style={[styles.container, { backgroundColor: colors.background }]}> 
        <View style={styles.content}> 
          {/* Guest Testing Banner */}
          {!player.email && !player.provider && (
            <View style={styles.guestBanner}>
              <Ionicons name="warning" size={20} color="#FFA500" />
                      <Text style={styles.guestBannerText}>
              {t('profile.guestAccountBanner')}
            </Text>
            </View>
          )}

          {/* Profile Header */}
          <View style={styles.profileHeader}>
            <TouchableOpacity
              onPress={() => setShowAvatarEditor(true)}
              style={styles.avatarContainer}
              activeOpacity={0.7}
            >
              <PlayerAvatar
                player={player}
                displaySize="large"
              />
              <View style={styles.editBadge}>
                <Ionicons name="pencil" size={14} color="#fff" />
              </View>
            </TouchableOpacity>
            <Text style={[styles.name, { color: colors.text }]}>{player.name}</Text>
            {player.email && (
              <Text style={[styles.email, { color: colors.textSecondary }]}>
                {player.email}
              </Text>
            )}
          </View>

          {/* Account Information */}
          <View style={styles.section}>
            <Text style={[styles.sectionTitle, { color: colors.text }]}>{t('profile.accountInfo')}</Text>
            
            <TouchableOpacity
              style={styles.infoRow}
              onPress={handleEditDisplayName}
              activeOpacity={0.7}
            >
              <Ionicons name="person-outline" size={20} color={colors.textSecondary} />
              <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>{t('profile.displayName')}:</Text>
              <Text style={[styles.infoValue, { color: colors.text }]}>{player.name}</Text>
              <Ionicons name="pencil" size={22} color={colors.textSecondary} style={styles.pencilIcon} />
            </TouchableOpacity>

            {player.email && (
              <View style={styles.infoRow}>
                <Ionicons name="mail-outline" size={20} color={colors.textSecondary} />
                <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>{t('profile.email')}:</Text>
                <Text style={[styles.infoValue, { color: colors.text }]}>{player.email}</Text>
              </View>
            )}



            {player.created_at && (
              <View style={styles.infoRow}>
                <Ionicons name="calendar-outline" size={20} color={colors.textSecondary} />
                <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>{t('profile.memberSince')}</Text>
                <Text style={[styles.infoValue, { color: colors.text }]}>
                  {formatShortDate(new Date(player.created_at))}
                </Text>
              </View>
            )}
          </View>

          <View style={styles.section}>
            <Text style={[styles.sectionTitle, { color: colors.text }]}>
              {t('notifications.title')}
            </Text>
            
            <View style={styles.infoRow}>
              <Ionicons name="notifications-outline" size={20} color={colors.textSecondary} />
              <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>
                {t('notifications.enableNotifications')}:
              </Text>
              <View style={{ flex: 0.7 }} />
              <Switch
                value={preferences.enabled}
                onValueChange={handleNotificationToggle}
                disabled={isTogglingNotifications}
                trackColor={{ false: colors.border, true: colors.secondary }}
                thumbColor={preferences.enabled ? colors.primary : colors.textSecondary}
              />
            </View>

            {!preferences.permissionGranted && preferences.enabled && (
              <View style={styles.permissionWarning}>
                <Ionicons name="warning-outline" size={16} color={colors.warning} />
                <Text style={[styles.permissionWarningText, { color: colors.textSecondary }]}>
                  {t('notifications.permissionWarning')}
                </Text>
              </View>
            )}
          </View>

          {/* Actions */}
          <View style={styles.section}>
            <Text style={[styles.sectionTitle, { color: colors.text }]}>{t('common.actions')}</Text>
            
            <TouchableOpacity style={styles.actionButton} onPress={handleSignOut}>
              <Ionicons name="log-out-outline" size={20} color={colors.text} />
              <Text style={[styles.actionButtonText, { color: colors.text }]}>{t('common.signOut')}</Text>
            </TouchableOpacity>

            <TouchableOpacity 
              style={[
                styles.actionButton, 
                styles.deleteButton,
                (!player?.email && !player?.provider) && styles.disabledButton
              ]} 
              onPress={handleDeleteAccount}
              testID="delete-account-button"
              disabled={!player?.email && !player?.provider}
            >
              <Ionicons 
                name="trash-outline" 
                size={20} 
                color={(!player?.email && !player?.provider) ? colors.textSecondary : colors.danger} 
              />
              <Text style={[
                styles.actionButtonText, 
                styles.deleteButtonText,
                (!player?.email && !player?.provider) && { color: colors.textSecondary }
              ]}>
                {t('profile.deleteAccount')}
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </ScrollView>
      {/* Display Name Change Modal */}
      <Modal
        visible={showDisplayNameModal}
        animationType="slide"
        transparent={true}
        onRequestClose={() => {
          setShowDisplayNameModal(false);
        }}
      >
        <View style={styles.modalOverlay}>
          <View style={[styles.modalContent, { backgroundColor: colors.card }]}> 
                <Text style={[styles.modalTitle, { color: colors.text }]}>
                  {t('profile.editDisplayName')}
                </Text>
                <Text style={[styles.modalSubtitle, { color: colors.textSecondary }]}>
                  {t('profile.editDisplayNameSubtitle')}
                </Text>
                
                <TextInput
                  ref={nameInputRef}
                  style={[styles.nameInput, { 
                    backgroundColor: colors.background,
                    color: colors.text,
                    borderColor: colors.border
                  }]}
                  placeholder={t('auth.enterDisplayName')}
                  placeholderTextColor={colors.textSecondary}
                  value={newDisplayName}
                  onChangeText={setNewDisplayName}
                  autoFocus={true}
                  maxLength={20}
                  autoCapitalize="words"
                />
                
                <View style={styles.modalButtons}>
                  <TouchableOpacity
                    style={[styles.modalButton, styles.cancelButton]}
                    onPress={handleModalCancel}
                    disabled={isUpdating}
                  >
                    <Text style={[styles.cancelButtonText, { color: colors.text }]}>{t('common.cancel')}</Text>
                  </TouchableOpacity>
                  
                  <TouchableOpacity
                    style={[styles.modalButton, styles.continueButton]}
                    onPress={handleModalSave}
                    disabled={isUpdating || !newDisplayName.trim()}
                  >
                    {isUpdating ? (
                      <ActivityIndicator color={colors.primary} />
                    ) : (
                      <Text style={styles.continueButtonText}>{t('common.save')}</Text>
                    )}
                  </TouchableOpacity>
                </View>
              </View>
        </View>
      </Modal>

      {/* Delete Account Warning Modal */}
      <Modal
        visible={showDeleteModal}
        animationType="slide"
        transparent={true}
        onRequestClose={handleDeleteModalCancel}
      >
        <View style={styles.modalOverlay}>
          <View style={[styles.modalContent, { backgroundColor: colors.card }]}>
            <Ionicons name="warning" size={60} color={colors.danger} style={{ marginBottom: 20 }} />
            <Text style={[styles.modalTitle, { color: colors.danger }]}>
              {t('profile.deleteAccount')}
            </Text>
            <Text style={[styles.modalSubtitle, { color: colors.text, textAlign: 'center' }]}>
              {t('profile.deleteAccountWarning')}
            </Text>
            <Text style={[styles.modalSubtitle, { color: colors.text, textAlign: 'center', marginTop: 10, fontWeight: 'bold' }]}>
              {t('profile.deleteAccountConfirm')}
            </Text>
            
            <View style={styles.modalButtons}>
              <TouchableOpacity
                style={[styles.modalButton, styles.cancelButton]}
                onPress={handleDeleteModalCancel}
              >
                <Text style={[styles.cancelButtonText, { color: colors.text }]}>{t('common.cancel')}</Text>
              </TouchableOpacity>
              
              <TouchableOpacity
                style={[styles.modalButton, { backgroundColor: colors.danger }]}
                onPress={handleDeleteAccountConfirm}
              >
                <Text style={[styles.continueButtonText, { color: '#fff' }]}>{t('common.continue')}</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>

      {/* Delete Account Final Confirmation Modal */}
      <Modal
        visible={showDeleteConfirmModal}
        animationType="slide"
        transparent={true}
        onRequestClose={handleDeleteConfirmModalCancel}
      >
        <View style={styles.modalOverlay}>
          <View style={[styles.modalContent, { backgroundColor: colors.card }]}>
            <Ionicons name="alert-circle" size={60} color={colors.danger} style={{ marginBottom: 20 }} />
            <Text style={[styles.modalTitle, { color: colors.danger }]}>
              {t('profile.deleteAccountFinalConfirm')}
            </Text>
            
            <TextInput
              ref={deleteInputRef}
              style={[styles.nameInput, { 
                backgroundColor: colors.background,
                color: colors.text,
                borderColor: colors.border
              }]}
              placeholder={t('profile.typeDeleteToConfirm')}
              placeholderTextColor={colors.textSecondary}
              value={deleteConfirmText}
              onChangeText={setDeleteConfirmText}
              autoFocus={true}
              autoCapitalize="characters"
              autoComplete="off"
              autoCorrect={false}
            />
            
            <View style={styles.modalButtons}>
              <TouchableOpacity
                style={[styles.modalButton, styles.cancelButton]}
                onPress={handleDeleteConfirmModalCancel}
                disabled={isDeleting}
              >
                <Text style={[styles.cancelButtonText, { color: colors.text }]}>{t('common.cancel')}</Text>
              </TouchableOpacity>
              
              <TouchableOpacity
                style={[styles.modalButton, { backgroundColor: colors.danger }]}
                onPress={handleDeleteAccountFinal}
                disabled={isDeleting || !deleteConfirmText.trim()}
              >
                {isDeleting ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <Text style={[styles.continueButtonText, { color: '#fff' }]}>{t('profile.deleteAccount')}</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>

      {/* Avatar Editor Modal */}
      <AvatarEditor
        currentAvatarUrl={player?.avatarUrl || null}
        onSave={uploadAvatar}
        onDelete={deleteAvatar}
        visible={showAvatarEditor}
        onClose={() => setShowAvatarEditor(false)}
      />
    </>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    padding: 20,
  },
  profileHeader: {
    alignItems: 'center',
    marginBottom: 32,
    paddingTop: 20,
  },
  avatarContainer: {
    position: 'relative',
    marginBottom: 16,
  },
  editBadge: {
    position: 'absolute',
    bottom: 0,
    right: 0,
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: colors.secondary,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: colors.background,
  },
  name: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  email: {
    fontSize: 16,
    opacity: 0.8,
  },
  section: {
    marginBottom: 32,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  infoLabel: {
    fontSize: 16,
    marginLeft: 12,
    marginRight: 0,
    minWidth: 120,
    flexShrink: 0,
  },
  infoValue: {
    fontSize: 16,
    flex: 1,
  },
  actionButton: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 20,
    backgroundColor: colors.card,
    borderRadius: 12,
    marginBottom: 12,
  },
  actionButtonText: {
    fontSize: 16,
    fontWeight: '600',
    marginLeft: 12,
  },
  errorText: {
    fontSize: 16,
    textAlign: 'center',
    marginTop: 50,
  },
  guestBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.warning,
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 12,
    marginBottom: 20,
  },
  guestBannerText: {
    fontSize: 16,
    color: colors.text,
    marginLeft: 10,
    fontWeight: '600',
  },
  editButton: {
    marginLeft: 10,
  },
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-start',
    paddingTop: 60,
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
  },
  modalContent: {
    width: '90%',
    minHeight: 320,
    padding: 30,
    borderRadius: 15,
    alignItems: 'center',
    justifyContent: 'flex-start',
  },
  modalTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  modalSubtitle: {
    fontSize: 16,
    marginBottom: 20,
    textAlign: 'center',
  },
  nameInput: {
    width: '100%',
    height: 50,
    borderWidth: 1,
    borderRadius: 10,
    paddingHorizontal: 15,
    fontSize: 18,
    marginBottom: 20,
  },
  modalButtons: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    width: '100%',
  },
  modalButton: {
    width: '45%',
    paddingVertical: 12,
    paddingHorizontal: 20,
    borderRadius: 10,
    alignItems: 'center',
  },
  cancelButton: {
    backgroundColor: colors.border,
  },
  cancelButtonText: {
    fontSize: 18,
    fontWeight: '600',
  },
  continueButton: {
    backgroundColor: colors.secondary,
  },
  continueButtonText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#fff',
  },
  pencilIcon: {
    marginLeft: 10,
  },
  deleteButton: {
    borderWidth: 1,
    borderColor: colors.danger,
    backgroundColor: 'transparent',
  },
  deleteButtonText: {
    color: colors.danger,
  },
  disabledButton: {
    opacity: 0.5,
  },
  permissionWarning: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 8,
    paddingLeft: 32,
  },
  permissionWarningText: {
    fontSize: 12,
    marginLeft: 8,
    flex: 1,
  },
}); 