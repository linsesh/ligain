import { Text, View, StyleSheet, Linking, TouchableOpacity } from 'react-native';
import { FontAwesome } from '@expo/vector-icons';
import { colors } from '../../src/constants/colors';
import { useTranslation } from '../../src/hooks/useTranslation';

export default function AboutScreen() {
  const { t } = useTranslation();
  
  const handleEmailPress = () => {
    Linking.openURL('mailto:contact@ligain.com');
  };

  const handleGithubPress = () => {
    Linking.openURL('https://github.com/linsesh/ligain');
  };

  const handleLinkedInPress = () => {
    Linking.openURL('https://fr.linkedin.com/in/beno%C3%AEt-linsey-fazi-74681a116');
  };

  const handleBuyMeACoffeePress = () => {
    Linking.openURL('https://buymeacoffee.com/linsesh');
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      <View style={styles.contentContainer}>
        <Text style={[styles.title, { color: colors.text }]}>{t('about.whatIsLigain')}</Text>
        <Text style={[styles.text, { color: colors.text }]}>
          {t('about.description')}
        </Text>
        <Text style={[styles.text, { color: colors.text }]}>
          {t('about.alphaNote')}
        </Text>
        <Text style={[styles.text, { color: colors.text }]}>
          {t('about.contactInfo')}{' '}
          <Text style={[styles.link, { color: colors.link }]} onPress={handleEmailPress}>contact@ligain.com</Text>
        </Text>

        <View style={styles.supportSection}>
          <Text style={[styles.supportTitle, { color: colors.text }]}>{t('about.supportSection')}</Text>
          <Text style={[styles.text, { color: colors.text }]}>
            {t('about.supportDescription')}
          </Text>
          <TouchableOpacity onPress={handleBuyMeACoffeePress} style={styles.buyMeACoffeeButton}>
            <FontAwesome name="coffee" size={20} color="#FFFFFF" />
            <Text style={styles.buyMeACoffeeText}>{t('about.buyMeACoffee')}</Text>
          </TouchableOpacity>
        </View>


        <View style={styles.socialContainer}>
          <TouchableOpacity onPress={handleGithubPress} style={styles.socialButton}>
            <FontAwesome name="github" size={24} color={colors.text} />
          </TouchableOpacity>
          <TouchableOpacity onPress={handleLinkedInPress} style={styles.socialButton}>
            <FontAwesome name="linkedin" size={24} color={colors.text} />
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  contentContainer: {
    padding: 20,
    width: '100%',
    maxWidth: 600,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  text: {
    fontSize: 16,
    lineHeight: 24,
    marginBottom: 16,
  },
  link: {
    textDecorationLine: 'underline',
  },
  socialContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    marginTop: 24,
    gap: 20,
  },
  socialButton: {
    padding: 10,
  },
  supportSection: {
    marginTop: 24,
    marginBottom: 24,
    padding: 16,
    backgroundColor: colors.card,
    borderRadius: 12,
  },
  supportTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  buyMeACoffeeButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#FFDD00',
    paddingVertical: 12,
    paddingHorizontal: 20,
    borderRadius: 8,
    marginTop: 16,
  },
  buyMeACoffeeText: {
    color: '#000000',
    fontSize: 16,
    fontWeight: '600',
    marginLeft: 8,
  },
});
