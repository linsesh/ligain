import React, { useState } from 'react';
import {
  View,
  StyleSheet,
  TouchableOpacity,
  Alert,
  ScrollView,
  Modal,
  TextInput,
  ActivityIndicator,
  Switch,
  Linking,
} from 'react-native';
import { Text } from '../../src/components/ui/Text';
import { SectionCard } from '../../src/components/ui/SectionCard';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { useGames } from '../../src/contexts/GamesContext';
import { colors } from '../../src/constants/colors';
import { useTranslation } from '../../src/hooks/useTranslation';
import { formatShortDate } from '../../src/utils/dateUtils';
import { API_CONFIG, authenticatedFetch } from '../../src/config/api';
import { mapPlayerFromBackend } from '../../src/api';
import { useNotifications } from '../../src/hooks/useNotifications';
import { useAutoReplicateBets } from '../../src/hooks/useAutoReplicateBets';
import { PlayerAvatar } from '../../src/components/PlayerAvatar';
import { AvatarEditor } from '../../src/components/AvatarEditor';

export default function ProfileScreen() {
  const { player, signOut, setPlayer, uploadAvatar, deleteAvatar } = useAuth();
  const { refresh: refreshGames } = useGames();
  const { t, isFrench } = useTranslation();
  const router = useRouter();
  const { preferences, setNotificationEnabled, requestPermissions } = useNotifications();
  const { enabled: autoReplicate, isLoading: autoReplicateLoading, toggle: toggleAutoReplicate } = useAutoReplicateBets();
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

  const isGuest = !player?.email && !player?.provider;

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

      if (setPlayer) {
        setPlayer(mapPlayerFromBackend(data.player));
      }

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
    if (isGuest) {
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
      <View className="flex-1 items-center justify-center" style={{ backgroundColor: 'transparent' }}>
        <Text className="text-base text-foreground-secondary">{t('profile.noUserInfo')}</Text>
      </View>
    );
  }

  return (
    <>
      <ScrollView
        className="flex-1"
        style={{ backgroundColor: 'transparent' }}
        contentContainerClassName="px-4 pb-8 gap-4"
        showsVerticalScrollIndicator={false}
      >
        {/* Guest Banner */}
        {isGuest && (
          <View
            className="flex-row items-center gap-3 rounded-xl px-4 py-3"
            style={{ backgroundColor: colors.warning + '33' }}
          >
            <Ionicons name="warning" size={20} color={colors.warning} />
            <Text className="flex-1 text-foreground text-sm font-hk-semibold">
              {t('profile.guestAccountBanner')}
            </Text>
          </View>
        )}

        {/* Profile Header */}
        <View className="bg-background rounded-2xl p-6 items-center">
          <TouchableOpacity
            onPress={() => setShowAvatarEditor(true)}
            activeOpacity={0.7}
            style={{ position: 'relative' }}
          >
            <PlayerAvatar player={player} displaySize="large" />
            <View
              className="absolute bottom-0 right-0 w-7 h-7 bg-secondary rounded-full items-center justify-center"
              style={{ borderWidth: 2, borderColor: colors.background }}
            >
              <Ionicons name="pencil" size={12} color="#fff" />
            </View>
          </TouchableOpacity>
          <Text className="text-2xl font-hk-bold text-foreground mt-4">{player.name}</Text>
          {player.email && (
            <Text className="text-sm text-foreground-secondary mt-1">{player.email}</Text>
          )}
          {player.created_at && (
            <Text className="text-xs text-foreground-secondary mt-2">
              {t('profile.memberSince')} {formatShortDate(new Date(player.created_at))}
            </Text>
          )}
        </View>

        {/* Account Info */}
        <SectionCard title={t('profile.accountInfo')}>
          <TouchableOpacity
            className="flex-row items-center gap-3 py-3 border-b border-border"
            onPress={handleEditDisplayName}
            activeOpacity={0.6}
          >
            <Ionicons name="person-outline" size={18} color={colors.textSecondary} />
            <Text className="text-sm text-foreground-secondary font-hk-semibold w-28">
              {t('profile.displayName')}:
            </Text>
            <Text className="flex-1 text-foreground text-sm">{player.name}</Text>
            <Ionicons name="pencil" size={16} color={colors.textSecondary} />
          </TouchableOpacity>

          {player.email && (
            <View className="flex-row items-center gap-3 py-3">
              <Ionicons name="mail-outline" size={18} color={colors.textSecondary} />
              <Text className="text-sm text-foreground-secondary font-hk-semibold w-28">
                {t('profile.email')}:
              </Text>
              <Text className="flex-1 text-foreground text-sm">{player.email}</Text>
            </View>
          )}
        </SectionCard>

        {/* Notifications & Settings */}
        <SectionCard title={t('notifications.title')}>
          <View className="flex-row items-center gap-3 py-3 border-b border-border">
            <Ionicons name="notifications-outline" size={18} color={colors.textSecondary} />
            <Text className="flex-1 text-sm text-foreground-secondary font-hk-semibold">
              {t('notifications.enableNotifications')}
            </Text>
            <Switch
              value={preferences.enabled}
              onValueChange={handleNotificationToggle}
              disabled={isTogglingNotifications}
              trackColor={{ false: colors.border, true: colors.secondary }}
              thumbColor={preferences.enabled ? colors.primary : colors.textSecondary}
            />
          </View>

          {!preferences.permissionGranted && preferences.enabled && (
            <View className="flex-row items-center gap-2 mt-2 px-2 py-2 bg-surface rounded-lg">
              <Ionicons name="warning-outline" size={14} color={colors.warning} />
              <Text className="flex-1 text-xs text-foreground-secondary">
                {t('notifications.permissionWarning')}
              </Text>
            </View>
          )}

          <View className="flex-row items-center gap-3 py-3">
            <Ionicons name="copy-outline" size={18} color={colors.textSecondary} />
            <Text className="flex-1 text-sm text-foreground-secondary font-hk-semibold">
              {t('settings.autoReplicateBets')}
            </Text>
            <Switch
              value={autoReplicate}
              onValueChange={toggleAutoReplicate}
              trackColor={{ false: colors.border, true: colors.secondary }}
              thumbColor={autoReplicate ? colors.primary : colors.textSecondary}
            />
          </View>
        </SectionCard>

        {/* Actions */}
        <SectionCard title={t('common.actions')}>
          <View className="gap-3">
            <TouchableOpacity
              className="flex-row items-center gap-3 px-4 py-4 bg-surface rounded-xl"
              onPress={() => router.push('/about')}
              activeOpacity={0.7}
            >
              <Ionicons name="information-circle-outline" size={20} color={colors.text} />
              <Text className="flex-1 text-foreground text-base font-hk-semibold">
                {t('navigation.about')}
              </Text>
              <Ionicons name="chevron-forward" size={16} color={colors.textSecondary} />
            </TouchableOpacity>

            <TouchableOpacity
              className="flex-row items-center gap-3 px-4 py-4 bg-surface rounded-xl"
              onPress={handleSignOut}
              activeOpacity={0.7}
            >
              <Ionicons name="log-out-outline" size={20} color={colors.text} />
              <Text className="flex-1 text-foreground text-base font-hk-semibold">
                {t('common.signOut')}
              </Text>
              <Ionicons name="chevron-forward" size={16} color={colors.textSecondary} />
            </TouchableOpacity>

            <TouchableOpacity
              className={`flex-row items-center gap-3 px-4 py-4 rounded-xl border ${
                isGuest ? 'opacity-50 border-border' : 'border-error'
              }`}
              onPress={handleDeleteAccount}
              testID="delete-account-button"
              disabled={isGuest}
              activeOpacity={0.7}
            >
              <Ionicons
                name="trash-outline"
                size={20}
                color={isGuest ? colors.textSecondary : colors.danger}
              />
              <Text className={`flex-1 text-base font-hk-semibold ${
                isGuest ? 'text-foreground-secondary' : 'text-error'
              }`}>
                {t('profile.deleteAccount')}
              </Text>
            </TouchableOpacity>
          </View>
        </SectionCard>
      </ScrollView>

      {/* Display Name Change Modal */}
      <Modal
        visible={showDisplayNameModal}
        animationType="slide"
        transparent={true}
        onRequestClose={() => setShowDisplayNameModal(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={[styles.modalContent, { backgroundColor: colors.card }]}>
            <View className="w-full gap-4">
              <Text className="text-xl font-hk-bold text-foreground text-center">
                {t('profile.editDisplayName')}
              </Text>
              <Text className="text-sm text-foreground-secondary text-center">
                {t('profile.editDisplayNameSubtitle')}
              </Text>

              <TextInput
                ref={nameInputRef}
                className="w-full border border-border rounded-xl px-4 text-base text-foreground"
                style={{ height: 50, backgroundColor: colors.background }}
                placeholder={t('auth.enterDisplayName')}
                placeholderTextColor={colors.textSecondary}
                value={newDisplayName}
                onChangeText={setNewDisplayName}
                autoFocus={true}
                maxLength={20}
                autoCapitalize="words"
              />

              <View className="flex-row gap-3">
                <TouchableOpacity
                  className="flex-1 py-3 rounded-xl items-center justify-center"
                  style={{ backgroundColor: colors.border }}
                  onPress={handleModalCancel}
                  disabled={isUpdating}
                >
                  <Text className="text-base font-hk-semibold text-foreground">
                    {t('common.cancel')}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  className="flex-1 py-3 rounded-xl items-center justify-center"
                  style={{ backgroundColor: colors.secondary, opacity: (!newDisplayName.trim() || isUpdating) ? 0.5 : 1 }}
                  onPress={handleModalSave}
                  disabled={isUpdating || !newDisplayName.trim()}
                >
                  {isUpdating ? (
                    <ActivityIndicator color="#fff" />
                  ) : (
                    <Text className="text-base font-hk-semibold text-white">
                      {t('common.save')}
                    </Text>
                  )}
                </TouchableOpacity>
              </View>
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
            <View className="w-full gap-4 items-center">
              <Ionicons name="warning" size={56} color={colors.danger} />
              <Text className="text-xl font-hk-bold text-center" style={{ color: colors.danger }}>
                {t('profile.deleteAccount')}
              </Text>
              <Text className="text-base text-foreground-secondary text-center">
                {t('profile.deleteAccountWarning')}
              </Text>
              <Text className="text-base font-hk-bold text-foreground text-center">
                {t('profile.deleteAccountConfirm')}
              </Text>

              <View className="flex-row gap-3 w-full">
                <TouchableOpacity
                  className="flex-1 py-3 rounded-xl items-center justify-center"
                  style={{ backgroundColor: colors.border }}
                  onPress={handleDeleteModalCancel}
                >
                  <Text className="text-base font-hk-semibold text-foreground">
                    {t('common.cancel')}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  className="flex-1 py-3 rounded-xl items-center justify-center"
                  style={{ backgroundColor: colors.danger }}
                  onPress={handleDeleteAccountConfirm}
                >
                  <Text className="text-base font-hk-semibold text-white">
                    {t('common.continue')}
                  </Text>
                </TouchableOpacity>
              </View>
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
            <View className="w-full gap-4 items-center">
              <Ionicons name="alert-circle" size={56} color={colors.danger} />
              <Text className="text-xl font-hk-bold text-center" style={{ color: colors.danger }}>
                {t('profile.deleteAccountFinalConfirm')}
              </Text>

              <TextInput
                ref={deleteInputRef}
                className="w-full border border-border rounded-xl px-4 text-base text-foreground"
                style={{ height: 50, backgroundColor: colors.background }}
                placeholder={t('profile.typeDeleteToConfirm')}
                placeholderTextColor={colors.textSecondary}
                value={deleteConfirmText}
                onChangeText={setDeleteConfirmText}
                autoFocus={true}
                autoCapitalize="characters"
                autoComplete="off"
                autoCorrect={false}
              />

              <View className="flex-row gap-3 w-full">
                <TouchableOpacity
                  className="flex-1 py-3 rounded-xl items-center justify-center"
                  style={{ backgroundColor: colors.border }}
                  onPress={handleDeleteConfirmModalCancel}
                  disabled={isDeleting}
                >
                  <Text className="text-base font-hk-semibold text-foreground">
                    {t('common.cancel')}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  className="flex-1 py-3 rounded-xl items-center justify-center"
                  style={{
                    backgroundColor: colors.danger,
                    opacity: (isDeleting || !deleteConfirmText.trim()) ? 0.5 : 1,
                  }}
                  onPress={handleDeleteAccountFinal}
                  disabled={isDeleting || !deleteConfirmText.trim()}
                >
                  {isDeleting ? (
                    <ActivityIndicator color="#fff" />
                  ) : (
                    <Text className="text-base font-hk-semibold text-white">
                      {t('profile.deleteAccount')}
                    </Text>
                  )}
                </TouchableOpacity>
              </View>
            </View>
          </View>
        </View>
      </Modal>

      {/* Avatar Editor */}
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
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-start',
    paddingTop: 60,
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
  },
  modalContent: {
    width: '90%',
    padding: 28,
    borderRadius: 20,
    alignItems: 'center',
  },
});
