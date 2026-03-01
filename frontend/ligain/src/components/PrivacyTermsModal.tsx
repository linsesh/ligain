import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Modal,
  ScrollView,
  useWindowDimensions,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../constants/colors';
import { useTranslation } from 'react-i18next';

interface PrivacyTermsModalProps {
  visible: boolean;
  onClose: () => void;
}

export const PrivacyTermsModal: React.FC<PrivacyTermsModalProps> = ({
  visible,
  onClose,
}) => {
  const { width } = useWindowDimensions();
  const { t, i18n } = useTranslation();
  const [activeTab, setActiveTab] = useState<'privacy' | 'terms'>('privacy');

  const isFrench = i18n.language === 'fr';

  const privacyPolicy = {
    fr: `🔐 Politique de confidentialité

Dernière mise à jour : 20 juillet 2025

Nous accordons une grande importance à la confidentialité de vos données. Cette politique explique quelles informations nous collectons, pourquoi, et comment elles sont utilisées.

1. Données collectées

Nous collectons les données suivantes lors de l'utilisation de l'application :
• Votre adresse email (via Google ou Apple)
• Un pseudonyme
• Vos pronostics
• Des informations techniques de connexion (ex : date et heure de dernière connexion)

2. Finalité de la collecte

Ces données sont utilisées pour :
• Vous identifier et vous permettre d'accéder à l'application
• Afficher vos pronostics à vos amis dans le cadre des ligues
• Améliorer la stabilité et la sécurité de l'application

Aucune donnée n'est utilisée à des fins commerciales ou publicitaires.

3. Base légale

La collecte est fondée sur votre consentement (inscription via Google ou Apple), et sur notre intérêt légitime à faire fonctionner l'application.

4. Hébergement des données

Vos données sont hébergées en Europe :
• Serveurs et logs : Google Cloud Platform (Europe)
• Base de données : Neon, hébergée en Allemagne

5. Partage des données

Nous ne partageons aucune donnée avec des tiers, sauf obligation légale.

6. Durée de conservation

Vos données sont conservées tant que vous avez un compte actif. Si vous demandez la suppression de votre compte, toutes vos données seront supprimées sous 30 jours.

7. Vos droits

Conformément au RGPD, vous disposez des droits suivants :
• Accès à vos données
• Rectification de vos données
• Suppression de votre compte et de vos données
• Limitation ou opposition au traitement

Vous pouvez exercer ces droits en nous contactant à : contact@ligain.com

8. Contact

Pour toute question concernant la protection de vos données :
Benoît Linsey Fazi
Email : contact@ligain.com`,

    en: `🔐 Privacy Policy

Last updated: July 20, 2025

We place great importance on the confidentiality of your data. This policy explains what information we collect, why, and how it is used.

1. Data Collected

We collect the following data when using the application:
• Your email address (via Google or Apple)
• A nickname
• Your predictions
• Technical connection information (e.g., date and time of last connection)

2. Purpose of Collection

This data is used to:
• Identify you and allow you to access the application
• Display your predictions to your friends within leagues
• Improve the stability and security of the application

No data is used for commercial or advertising purposes.

3. Legal Basis

The collection is based on your consent (registration via Google or Apple), and on our legitimate interest in making the application work.

4. Data Hosting

Your data is hosted in Europe:
• Servers and logs: Google Cloud Platform (Europe)
• Database: Neon, hosted in Germany

5. Data Sharing

We do not share any data with third parties, except legal obligation.

6. Data Retention

Your data is retained as long as you have an active account. If you request account deletion, all your data will be deleted within 30 days.

7. Your Rights

In accordance with GDPR, you have the following rights:
• Access to your data
• Rectification of your data
• Deletion of your account and data
• Limitation or opposition to processing

You can exercise these rights by contacting us at: contact@ligain.com

8. Contact

For any questions regarding the protection of your data:
Benoît Linsey Fazi
Email: contact@ligain.com`
  };

  const termsOfService = {
    fr: `📄 Conditions Générales d'Utilisation (CGU)

Dernière mise à jour : 20 juillet 2025

1. Objet

Les présentes conditions régissent l'utilisation de l'application de pronostics entre amis proposée par Ligain.

2. Inscription

L'inscription se fait via un compte Google ou Apple. L'utilisateur doit fournir une adresse email et un pseudonyme.

3. Fonctionnement

L'application permet de créer ou rejoindre des ligues privées de pronostics. Les résultats et classements sont visibles uniquement par les membres des ligues.

4. Responsabilités

L'utilisateur s'engage à :
• Fournir des informations exactes
• Ne pas usurper l'identité d'un tiers
• Respecter les autres joueurs (pas de pseudos inappropriés)

Le développeur se réserve le droit de suspendre un compte en cas de triche ou d'abus manifeste.

5. Propriété

L'application et son contenu sont protégés par le droit de la propriété intellectuelle. Vous ne pouvez pas la copier, la modifier ou la distribuer sans autorisation.

6. Suppression de compte

Vous pouvez demander la suppression de votre compte à tout moment en nous contactant à : contact@ligain.com. La suppression est effective sous 30 jours.

7. Limitation de responsabilité

L'application est fournie "en l'état". Aucune garantie n'est donnée quant à la disponibilité ou l'exactitude des données (par exemple, en cas de bug ou de panne serveur).

8. Droit applicable

Les présentes conditions sont régies par le droit français. En cas de litige, les tribunaux compétents seront ceux du ressort de Paris, France.`,

    en: `📄 Terms of Service

Last updated: July 20, 2025

1. Purpose

These terms govern the use of the friend betting application offered by Ligain.

2. Registration

Registration is done via a Google or Apple account. The user must provide an email address and a nickname.

3. Operation

The application allows you to create or join private betting leagues. Results and rankings are only visible to league members.

4. Responsibilities

The user agrees to:
• Provide accurate information
• Not impersonate a third party
• Respect other players (no inappropriate nicknames)

The developer reserves the right to suspend an account in case of cheating or obvious abuse.

5. Ownership

The application and its content are protected by intellectual property law. You may not copy, modify, or distribute it without permission.

6. Account Deletion

You can request deletion of your account at any time by contacting us at: contact@ligain.com. Deletion is effective within 30 days.

7. Limitation of Liability

The application is provided "as is". No warranty is given as to the availability or accuracy of data (for example, in case of bugs or server failure).

8. Applicable Law

These terms are governed by French law. In case of dispute, the competent courts will be those of Paris, France.`
  };

  const currentContent = activeTab === 'privacy' 
    ? (isFrench ? privacyPolicy.fr : privacyPolicy.en)
    : (isFrench ? termsOfService.fr : termsOfService.en);

  return (
    <Modal
      visible={visible}
      animationType="slide"
      transparent={true}
      onRequestClose={onClose}
    >
      <View style={styles.modalOverlay}>
        <View style={[styles.modalContent, { width: width - 20, height: '95%' }]}>
          {/* Header */}
          <View style={styles.header}>
            <Text className="font-hk-bold" style={[styles.title, { color: colors.text }]}>
              {activeTab === 'privacy'
                ? (isFrench ? 'Politique de confidentialité' : 'Privacy Policy')
                : (isFrench ? 'Conditions d\'utilisation' : 'Terms of Service')
              }
            </Text>
            <TouchableOpacity onPress={onClose} style={styles.closeButton}>
              <Ionicons name="close" size={24} color={colors.text} />
            </TouchableOpacity>
          </View>

          {/* Tab Navigation */}
          <View style={styles.tabContainer}>
            <TouchableOpacity
              style={[
                styles.tab,
                activeTab === 'privacy' && styles.activeTab
              ]}
              onPress={() => setActiveTab('privacy')}
            >
              <Text className="font-hk-semibold" style={[
                styles.tabText,
                { color: activeTab === 'privacy' ? colors.primary : colors.textSecondary }
              ]}>
                {isFrench ? 'Confidentialité' : 'Privacy'}
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[
                styles.tab,
                activeTab === 'terms' && styles.activeTab
              ]}
              onPress={() => setActiveTab('terms')}
            >
              <Text className="font-hk-semibold" style={[
                styles.tabText,
                { color: activeTab === 'terms' ? colors.primary : colors.textSecondary }
              ]}>
                {isFrench ? 'Conditions' : 'Terms'}
              </Text>
            </TouchableOpacity>
          </View>

          {/* Content */}
          <ScrollView 
            style={styles.content}
            showsVerticalScrollIndicator={false}
          >
            <Text style={[styles.contentText, { color: colors.text }]}>
              {currentContent}
            </Text>
          </ScrollView>

          {/* Footer */}
          <View style={styles.footer}>
            <TouchableOpacity
              style={[styles.closeModalButton, { backgroundColor: colors.primary }]}
              onPress={onClose}
            >
              <Text className="font-hk-semibold" style={styles.closeModalButtonText}>
                {t('common.ok')}
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 10,
  },
  modalContent: {
    backgroundColor: colors.card,
    borderRadius: 16,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  title: {
    fontSize: 20,
    flex: 1,
  },
  closeButton: {
    padding: 4,
  },
  tabContainer: {
    flexDirection: 'row',
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  tab: {
    flex: 1,
    paddingVertical: 16,
    alignItems: 'center',
  },
  activeTab: {
    borderBottomWidth: 2,
    borderBottomColor: colors.primary,
  },
  tabText: {
    fontSize: 18,
  },
  content: {
    flex: 1,
    padding: 20,
  },
  contentText: {
    fontSize: 16,
    lineHeight: 24,
    textAlign: 'left',
  },
  footer: {
    padding: 20,
    borderTopWidth: 1,
    borderTopColor: colors.border,
  },
  closeModalButton: {
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
  },
  closeModalButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
  },
}); 