import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Modal,
  Linking,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../constants/colors';
import { useTranslation } from 'react-i18next';
import { useUpdateRequired } from '../contexts/UpdateRequiredContext';

export const UpdateRequiredModal = () => {
  const { t } = useTranslation();
  const { isUpdateRequired, storeUrl } = useUpdateRequired();

  const handleUpdate = async () => {
    if (storeUrl) {
      try {
        await Linking.openURL(storeUrl);
      } catch (error) {
        console.error('Failed to open store URL:', error);
      }
    }
  };

  if (!isUpdateRequired) {
    return null;
  }

  return (
    <Modal
      visible={isUpdateRequired}
      animationType="fade"
      transparent={true}
      // No onRequestClose - modal cannot be dismissed
    >
      <View style={styles.modalOverlay}>
        <View style={styles.modalContent}>
          {/* Icon */}
          <View style={styles.iconContainer}>
            <Ionicons name="arrow-up-circle" size={64} color={colors.primary} />
          </View>

          {/* Title */}
          <Text style={styles.title}>{t('update.title')}</Text>

          {/* Message */}
          <Text style={styles.message}>{t('update.message')}</Text>

          {/* Update Button */}
          <TouchableOpacity
            style={styles.updateButton}
            onPress={handleUpdate}
            testID="update-button"
          >
            <Text style={styles.updateButtonText}>{t('update.updateButton')}</Text>
          </TouchableOpacity>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.85)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  modalContent: {
    backgroundColor: colors.card,
    borderRadius: 16,
    padding: 32,
    alignItems: 'center',
    maxWidth: 340,
    width: '100%',
  },
  iconContainer: {
    marginBottom: 24,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: colors.text,
    textAlign: 'center',
    marginBottom: 16,
  },
  message: {
    fontSize: 16,
    color: colors.textSecondary,
    textAlign: 'center',
    lineHeight: 24,
    marginBottom: 32,
  },
  updateButton: {
    backgroundColor: colors.primary,
    paddingVertical: 14,
    paddingHorizontal: 32,
    borderRadius: 8,
    width: '100%',
    alignItems: 'center',
  },
  updateButtonText: {
    color: colors.background,
    fontSize: 16,
    fontWeight: '600',
  },
});
