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
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../../src/contexts/AuthContext';
import { colors } from '../../src/constants/colors';
import { useTranslation } from '../../src/hooks/useTranslation';
import { formatShortDate } from '../../src/utils/dateUtils';
import { API_CONFIG, getApiHeaders, getAuthenticatedHeaders } from '../../src/config/api';

export default function ProfileScreen() {
  const { player, signOut, setPlayer } = useAuth();
  const { t } = useTranslation();
  const [showDisplayNameModal, setShowDisplayNameModal] = useState(false);
  const [newDisplayName, setNewDisplayName] = useState('');
  const [isUpdating, setIsUpdating] = useState(false);
  const nameInputRef = React.useRef<TextInput>(null);

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
      
      const headers = await getAuthenticatedHeaders({ 'Content-Type': 'application/json' });
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/profile/display-name`, {
        method: 'PUT',
        headers,
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
        setPlayer(data.player);
      }

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
            <View style={styles.avatarPlaceholder}>
              <Ionicons name="person" size={40} color={colors.text} />
            </View>
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

          {/* Actions */}
          <View style={styles.section}>
            <Text style={[styles.sectionTitle, { color: colors.text }]}>{t('common.actions')}</Text>
            
            <TouchableOpacity style={styles.actionButton} onPress={handleSignOut}>
              <Ionicons name="log-out-outline" size={20} color={colors.text} />
              <Text style={[styles.actionButtonText, { color: colors.text }]}>{t('common.signOut')}</Text>
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

  avatarPlaceholder: {
    width: 100,
    height: 100,
    borderRadius: 50,
    backgroundColor: colors.card,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
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
}); 