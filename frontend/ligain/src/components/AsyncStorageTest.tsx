import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Alert } from 'react-native';
import { getItem, setItem, removeItem, clear } from '../utils/storage';
import { colors } from '../constants/colors';

export const AsyncStorageTest: React.FC = () => {
  const [testValue, setTestValue] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const testKey = '__async_storage_test__';

  const runTest = async () => {
    setIsLoading(true);
    try {
      // Test 1: Set item
      await setItem(testKey, 'test_value_123');
      console.log('✅ Set item successful');

      // Test 2: Get item
      const retrievedValue = await getItem(testKey);
      setTestValue(retrievedValue);
      console.log('✅ Get item successful:', retrievedValue);

      // Test 3: Remove item
      await removeItem(testKey);
      console.log('✅ Remove item successful');

      // Test 4: Verify removal
      const afterRemoval = await getItem(testKey);
      console.log('✅ Verify removal successful:', afterRemoval);

      Alert.alert(
        'AsyncStorage Test',
        'All tests passed! AsyncStorage is working correctly.',
        [{ text: 'OK' }]
      );
    } catch (error) {
      console.error('❌ AsyncStorage test failed:', error);
      Alert.alert(
        'AsyncStorage Test Failed',
        `Error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        [{ text: 'OK' }]
      );
    } finally {
      setIsLoading(false);
    }
  };

  const clearAll = async () => {
    try {
      await clear();
      setTestValue(null);
      Alert.alert('Success', 'All storage cleared');
    } catch (error) {
      Alert.alert('Error', 'Failed to clear storage');
    }
  };

  useEffect(() => {
    // Check if there's a stored test value on component mount
    getItem(testKey).then(value => {
      setTestValue(value);
    });
  }, []);

  return (
    <View style={styles.container}>
      <Text style={[styles.title, { color: colors.text }]}>
        AsyncStorage Test
      </Text>
      
      <Text style={[styles.status, { color: colors.textSecondary }]}>
        Current test value: {testValue || 'None'}
      </Text>

      <TouchableOpacity
        style={[styles.button, { backgroundColor: colors.link }]}
        onPress={runTest}
        disabled={isLoading}
      >
        <Text style={styles.buttonText}>
          {isLoading ? 'Running Test...' : 'Run AsyncStorage Test'}
        </Text>
      </TouchableOpacity>

      <TouchableOpacity
        style={[styles.button, { backgroundColor: colors.border }]}
        onPress={clearAll}
      >
        <Text style={styles.buttonText}>Clear All Storage</Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    padding: 20,
    alignItems: 'center',
  },
  title: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  status: {
    fontSize: 14,
    marginBottom: 20,
    textAlign: 'center',
  },
  button: {
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    marginVertical: 8,
    minWidth: 200,
    alignItems: 'center',
  },
  buttonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
}); 