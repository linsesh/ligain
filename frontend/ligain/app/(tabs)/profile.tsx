import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  ScrollView,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../../src/contexts/AuthContext';
import { colors } from '../../src/constants/colors';
import { useTranslation } from '../../src/hooks/useTranslation';

export default function ProfileScreen() {
  const { player, signOut } = useAuth();
  const { t } = useTranslation();

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
          
          <View style={styles.infoRow}>
            <Ionicons name="person-outline" size={20} color={colors.textSecondary} />
            <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>{t('profile.displayName')}:</Text>
            <Text style={[styles.infoValue, { color: colors.text }]}>{player.name}</Text>
          </View>

          {player.email && (
            <View style={styles.infoRow}>
              <Ionicons name="mail-outline" size={20} color={colors.textSecondary} />
              <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>{t('profile.email')}:</Text>
              <Text style={[styles.infoValue, { color: colors.text }]}>{player.email}</Text>
            </View>
          )}

          {player.provider && (
            <View style={styles.infoRow}>
              <Ionicons 
                name={player.provider === 'google' ? 'logo-google' : 'logo-apple'} 
                size={20} 
                color={colors.textSecondary} 
              />
              <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>{t('profile.provider')}:</Text>
              <Text style={[styles.infoValue, { color: colors.text }]}>
                {player.provider.charAt(0).toUpperCase() + player.provider.slice(1)}
              </Text>
            </View>
          )}

          {player.created_at && (
            <View style={styles.infoRow}>
              <Ionicons name="calendar-outline" size={20} color={colors.textSecondary} />
              <Text style={[styles.infoLabel, { color: colors.textSecondary }]}>Member since:</Text>
              <Text style={[styles.infoValue, { color: colors.text }]}>
                {new Date(player.created_at).toLocaleDateString()}
              </Text>
            </View>
          )}
        </View>

        {/* Actions */}
        <View style={styles.section}>
          <Text style={[styles.sectionTitle, { color: colors.text }]}>Actions</Text>
          
          <TouchableOpacity style={styles.actionButton} onPress={handleSignOut}>
            <Ionicons name="log-out-outline" size={20} color={colors.error} />
            <Text style={[styles.actionButtonText, { color: colors.error }]}>{t('common.signOut')}</Text>
          </TouchableOpacity>
        </View>
      </View>
    </ScrollView>
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
    marginRight: 8,
    minWidth: 80,
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
}); 