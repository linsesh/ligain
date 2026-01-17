import React, { useState } from 'react';
import {
  View,
  Text,
  Image,
  TouchableOpacity,
  Modal,
  StyleSheet,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import * as ImagePicker from 'expo-image-picker';
import { useTranslation } from '../hooks/useTranslation';
import { colors } from '../constants/colors';
import { validateAvatarImage } from '../utils/avatarValidation';
import { AvatarErrorCode } from '../api/types';

interface AvatarEditorProps {
  currentAvatarUrl: string | null;
  onSave: (imageUri: string) => Promise<void>;
  onDelete: () => Promise<void>;
  visible: boolean;
  onClose: () => void;
}

export function AvatarEditor({
  currentAvatarUrl,
  onSave,
  onDelete,
  visible,
  onClose,
}: AvatarEditorProps) {
  const { t } = useTranslation();
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  const getErrorMessage = (code: AvatarErrorCode): string => {
    switch (code) {
      case 'FILE_TOO_LARGE':
        return t('avatar.error.tooLarge');
      case 'IMAGE_TOO_SMALL':
        return t('avatar.error.tooSmall');
      case 'INVALID_IMAGE':
        return t('avatar.error.invalidImage');
      case 'UPLOAD_FAILED':
        return t('avatar.error.uploadFailed');
      default:
        return t('avatar.error.uploadFailed');
    }
  };

  const pickImage = async (useCamera: boolean) => {
    try {
      // Request permissions
      if (useCamera) {
        const { status } = await ImagePicker.requestCameraPermissionsAsync();
        if (status !== 'granted') {
          Alert.alert(t('errors.error'), 'Camera permission is required');
          return;
        }
      } else {
        const { status } = await ImagePicker.requestMediaLibraryPermissionsAsync();
        if (status !== 'granted') {
          Alert.alert(t('errors.error'), 'Photo library permission is required');
          return;
        }
      }

      const result = useCamera
        ? await ImagePicker.launchCameraAsync({
            mediaTypes: ImagePicker.MediaTypeOptions.Images,
            allowsEditing: true,
            aspect: [1, 1],
            quality: 0.8,
          })
        : await ImagePicker.launchImageLibraryAsync({
            mediaTypes: ImagePicker.MediaTypeOptions.Images,
            allowsEditing: true,
            aspect: [1, 1],
            quality: 0.8,
          });

      if (result.canceled || !result.assets?.[0]?.uri) {
        return;
      }

      const imageUri = result.assets[0].uri;

      // Validate the image
      const validation = await validateAvatarImage(imageUri);
      if (!validation.valid) {
        Alert.alert(t('errors.error'), getErrorMessage(validation.error!));
        return;
      }

      setSelectedImage(imageUri);
    } catch (error) {
      console.error('Error picking image:', error);
      Alert.alert(t('errors.error'), t('avatar.error.invalidImage'));
    }
  };

  const handleSave = async () => {
    if (!selectedImage) return;

    try {
      setIsUploading(true);
      await onSave(selectedImage);
      setSelectedImage(null);
      onClose();
    } catch (error) {
      console.error('Error uploading avatar:', error);
      Alert.alert(t('errors.error'), t('avatar.error.uploadFailed'));
    } finally {
      setIsUploading(false);
    }
  };

  const handleDelete = () => {
    Alert.alert(
      t('avatar.removeConfirmTitle'),
      t('avatar.removeConfirmMessage'),
      [
        { text: t('common.cancel'), style: 'cancel' },
        {
          text: t('common.confirm'),
          style: 'destructive',
          onPress: async () => {
            try {
              setIsUploading(true);
              await onDelete();
              onClose();
            } catch (error) {
              console.error('Error deleting avatar:', error);
              Alert.alert(t('errors.error'), t('avatar.error.uploadFailed'));
            } finally {
              setIsUploading(false);
            }
          },
        },
      ]
    );
  };

  const handleClose = () => {
    setSelectedImage(null);
    onClose();
  };

  const previewImage = selectedImage || currentAvatarUrl;

  return (
    <Modal
      visible={visible}
      animationType="slide"
      transparent={true}
      onRequestClose={handleClose}
    >
      <View style={styles.overlay}>
        <View style={[styles.content, { backgroundColor: colors.card }]}>
          <Text style={[styles.title, { color: colors.text }]}>
            {t('avatar.editAvatar')}
          </Text>

          {/* Preview */}
          <View style={styles.previewContainer}>
            {previewImage ? (
              <Image
                source={{ uri: previewImage }}
                style={styles.previewImage}
                testID="preview-image"
              />
            ) : (
              <View style={[styles.placeholder, { backgroundColor: colors.border }]}>
                <Ionicons name="person" size={60} color={colors.textSecondary} />
              </View>
            )}
          </View>

          {/* Loading overlay */}
          {isUploading && (
            <View style={styles.loadingOverlay} testID="upload-loading">
              <ActivityIndicator size="large" color={colors.primary} />
              <Text style={[styles.loadingText, { color: colors.text }]}>
                {t('avatar.uploading')}
              </Text>
            </View>
          )}

          {/* Actions */}
          <View style={styles.actions}>
            <TouchableOpacity
              style={[styles.actionButton, { backgroundColor: colors.background }]}
              onPress={() => pickImage(true)}
              disabled={isUploading}
            >
              <Ionicons name="camera-outline" size={24} color={colors.text} />
              <Text style={[styles.actionText, { color: colors.text }]}>
                {t('avatar.takePhoto')}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.actionButton, { backgroundColor: colors.background }]}
              onPress={() => pickImage(false)}
              disabled={isUploading}
            >
              <Ionicons name="images-outline" size={24} color={colors.text} />
              <Text style={[styles.actionText, { color: colors.text }]}>
                {t('avatar.chooseFromLibrary')}
              </Text>
            </TouchableOpacity>

            {currentAvatarUrl && (
              <TouchableOpacity
                style={[styles.actionButton, styles.deleteButton]}
                onPress={handleDelete}
                disabled={isUploading}
              >
                <Ionicons name="trash-outline" size={24} color={colors.danger} />
                <Text style={[styles.actionText, { color: colors.danger }]}>
                  {t('avatar.removeAvatar')}
                </Text>
              </TouchableOpacity>
            )}
          </View>

          {/* Bottom buttons */}
          <View style={styles.bottomButtons}>
            <TouchableOpacity
              style={[styles.bottomButton, { backgroundColor: colors.border }]}
              onPress={handleClose}
              disabled={isUploading}
            >
              <Text style={[styles.bottomButtonText, { color: colors.text }]}>
                {t('common.cancel')}
              </Text>
            </TouchableOpacity>

            {selectedImage && (
              <TouchableOpacity
                style={[styles.bottomButton, { backgroundColor: colors.secondary }]}
                onPress={handleSave}
                disabled={isUploading}
              >
                <Text style={[styles.bottomButtonText, { color: '#fff' }]}>
                  {t('avatar.save')}
                </Text>
              </TouchableOpacity>
            )}
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    justifyContent: 'flex-end',
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
  },
  content: {
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 20,
    paddingBottom: 40,
  },
  title: {
    fontSize: 20,
    fontWeight: 'bold',
    textAlign: 'center',
    marginBottom: 20,
  },
  previewContainer: {
    alignItems: 'center',
    marginBottom: 24,
  },
  previewImage: {
    width: 120,
    height: 120,
    borderRadius: 60,
  },
  placeholder: {
    width: 120,
    height: 120,
    borderRadius: 60,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    borderRadius: 20,
    zIndex: 10,
  },
  loadingText: {
    marginTop: 12,
    fontSize: 16,
  },
  actions: {
    gap: 12,
    marginBottom: 20,
  },
  actionButton: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderRadius: 12,
    gap: 12,
  },
  deleteButton: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: colors.danger,
  },
  actionText: {
    fontSize: 16,
    fontWeight: '500',
  },
  bottomButtons: {
    flexDirection: 'row',
    gap: 12,
  },
  bottomButton: {
    flex: 1,
    padding: 16,
    borderRadius: 12,
    alignItems: 'center',
  },
  bottomButtonText: {
    fontSize: 16,
    fontWeight: '600',
  },
});
