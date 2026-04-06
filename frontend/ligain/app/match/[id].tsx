import { useLocalSearchParams, useRouter } from 'expo-router';
import { View, TouchableOpacity } from 'react-native';
import { Text } from '../../src/components/ui/Text';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../../src/constants/colors';

export default function MatchDetailScreen() {
  const { id, gameId } = useLocalSearchParams<{ id: string; gameId: string }>();
  const router = useRouter();

  return (
    <View style={{ flex: 1, backgroundColor: colors.background }}>
      <TouchableOpacity onPress={() => router.back()} style={{ padding: 16 }}>
        <Ionicons name="arrow-back" size={24} color={colors.text} />
      </TouchableOpacity>
      <View style={{ flex: 1, alignItems: 'center', justifyContent: 'center' }}>
        <Text>Match {id}</Text>
      </View>
    </View>
  );
}
