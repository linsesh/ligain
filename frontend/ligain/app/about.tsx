import { View, ScrollView, TouchableOpacity, Linking } from 'react-native';
import { Text } from '../src/components/ui/Text';
import { FontAwesome, Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useTranslation } from '../src/hooks/useTranslation';
import { SectionCard } from '../src/components/ui/SectionCard';
import { colors } from '../src/constants/colors';

export default function AboutScreen() {
  const { t } = useTranslation();
  const router = useRouter();

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
    <View className="flex-1" style={{ backgroundColor: 'transparent' }}>
      <TouchableOpacity onPress={() => router.back()} className="self-start p-4">
        <Ionicons name="arrow-back" size={24} className="text-foreground" color={colors.text} />
      </TouchableOpacity>

      <ScrollView
        className="flex-1"
        contentContainerClassName="px-4 pb-8 gap-4"
        showsVerticalScrollIndicator={false}
      >
        {/* Hero block */}
        <View className="items-center py-8">
          <Text className="text-6xl font-extrabold text-primary">LIGAIN</Text>
        </View>

        {/* What is Ligain? */}
        <SectionCard title={t('about.whatIsLigain')}>
          <Text className="text-base text-foreground leading-6 mb-3">
            {t('about.description')}
          </Text>
          <Text className="text-base text-foreground-secondary leading-6">
            {t('about.alphaNote')}
          </Text>
        </SectionCard>

        {/* Contact */}
        <SectionCard title="Contact">
          <Text className="text-base text-foreground leading-6">
            {t('about.contactInfo')}{' '}
            <Text className="text-link underline" onPress={handleEmailPress}>
              contact@ligain.com
            </Text>
          </Text>
        </SectionCard>

        {/* Support the Project */}
        <SectionCard title={t('about.supportSection')}>
          <Text className="text-base text-foreground leading-6 mb-4">
            {t('about.supportDescription')}
          </Text>
          <TouchableOpacity
            onPress={handleBuyMeACoffeePress}
            className="flex-row items-center justify-center bg-primary rounded-full py-3 px-6"
          >
            <FontAwesome name="coffee" size={18} color="#ffffff" />
            <Text className="text-white text-base font-semibold ml-2">
              {t('about.buyMeACoffee')}
            </Text>
          </TouchableOpacity>
        </SectionCard>

        {/* Social row */}
        <View className="flex-row justify-center gap-6 mt-4">
          <TouchableOpacity
            onPress={handleGithubPress}
            className="bg-surface rounded-full p-3"
          >
            <FontAwesome name="github" size={24} color={colors.text} />
          </TouchableOpacity>
          <TouchableOpacity
            onPress={handleLinkedInPress}
            className="bg-surface rounded-full p-3"
          >
            <FontAwesome name="linkedin" size={24} color={colors.text} />
          </TouchableOpacity>
        </View>
      </ScrollView>
    </View>
  );
}
