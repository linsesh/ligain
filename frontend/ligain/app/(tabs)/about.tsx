import { Text, View, StyleSheet, Linking, TouchableOpacity } from 'react-native';
import { FontAwesome } from '@expo/vector-icons';
import { colors } from '../../src/constants/colors';

export default function AboutScreen() {
  const handleEmailPress = () => {
    Linking.openURL('mailto:benoitlinsey27@gmail.com');
  };

  const handleGithubPress = () => {
    Linking.openURL('https://github.com/linsesh/ligain');
  };

  const handleLinkedInPress = () => {
    Linking.openURL('https://fr.linkedin.com/in/beno%C3%AEt-linsey-fazi-74681a116');
  };

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      <View style={styles.contentContainer}>
        <Text style={[styles.title, { color: colors.text }]}>What is Ligain?</Text>
        <Text style={[styles.text, { color: colors.text }]}>
          Ligain is a platform for betting on football matches between friends.
        </Text>
        <Text style={[styles.text, { color: colors.text }]}>
          It's currently in alpha and only available for Ligue 1 matches.
        </Text>
        <Text style={[styles.text, { color: colors.text }]}>
          The application is being made by a single developer, so it may not be perfect.
        </Text>
        <Text style={[styles.text, { color: colors.text }]}>
          If you have any feedback, please contact me at{' '}
          <Text style={[styles.link, { color: colors.link }]} onPress={handleEmailPress}>benoitlinsey27@gmail.com</Text>
        </Text>

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
});
