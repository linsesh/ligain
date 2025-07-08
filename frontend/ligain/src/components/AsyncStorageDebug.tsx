import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Alert, ScrollView } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { getItem, setItem, removeItem, clear, isUsingMemoryFallback } from '../utils/storage';
import { colors } from '../constants/colors';

export const AsyncStorageDebug: React.FC = () => {
  const [debugInfo, setDebugInfo] = useState<any>({});
  const [isLoading, setIsLoading] = useState(false);

  const testAsyncStorage = async () => {
    setIsLoading(true);
    const info: any = {};

    try {
      // Test 1: Check if AsyncStorage is available
      info.asyncStorageAvailable = !!AsyncStorage;
      info.asyncStorageType = typeof AsyncStorage;
      info.asyncStorageMethods = AsyncStorage ? Object.keys(AsyncStorage) : [];

      // Test 2: Check if setItem is a function
      info.setItemIsFunction = typeof AsyncStorage?.setItem === 'function';

      // Test 3: Try direct AsyncStorage operations
      try {
        await AsyncStorage.setItem('__debug_test__', 'test_value');
        const result = await AsyncStorage.getItem('__debug_test__');
        await AsyncStorage.removeItem('__debug_test__');
        info.directAsyncStorageWorks = result === 'test_value';
      } catch (error) {
        info.directAsyncStorageWorks = false;
        info.directAsyncStorageError = error instanceof Error ? error.message : String(error);
      }

      // Test 4: Test our safe wrapper
      try {
        await setItem('__safe_test__', 'safe_test_value');
        const result = await getItem('__safe_test__');
        await removeItem('__safe_test__');
        info.safeWrapperWorks = result === 'safe_test_value';
      } catch (error) {
        info.safeWrapperWorks = false;
        info.safeWrapperError = error instanceof Error ? error.message : String(error);
      }

      // Test 5: Check fallback status
      info.usingMemoryFallback = isUsingMemoryFallback();

      // Test 6: Check all keys
      try {
        const keys = await AsyncStorage.getAllKeys();
        info.totalKeys = keys.length;
        info.keys = keys.slice(0, 10); // Show first 10 keys
      } catch (error) {
        info.totalKeys = 'Error';
        info.keysError = error instanceof Error ? error.message : String(error);
      }

    } catch (error) {
      info.generalError = error instanceof Error ? error.message : String(error);
    }

    setDebugInfo(info);
    setIsLoading(false);
  };

  const clearAllStorage = async () => {
    try {
      await clear();
      Alert.alert('Success', 'All storage cleared');
      testAsyncStorage(); // Refresh debug info
    } catch (error) {
      Alert.alert('Error', 'Failed to clear storage');
    }
  };

  useEffect(() => {
    testAsyncStorage();
  }, []);

  return (
    <ScrollView style={styles.container}>
      <Text style={[styles.title, { color: colors.text }]}>
        AsyncStorage Debug Info
      </Text>

      <TouchableOpacity
        style={[styles.button, { backgroundColor: colors.link }]}
        onPress={testAsyncStorage}
        disabled={isLoading}
      >
        <Text style={styles.buttonText}>
          {isLoading ? 'Testing...' : 'Refresh Debug Info'}
        </Text>
      </TouchableOpacity>

      <TouchableOpacity
        style={[styles.button, { backgroundColor: colors.border }]}
        onPress={clearAllStorage}
      >
        <Text style={styles.buttonText}>Clear All Storage</Text>
      </TouchableOpacity>

      <View style={styles.debugContainer}>
        {Object.entries(debugInfo).map(([key, value]) => (
          <View key={key} style={styles.debugRow}>
            <Text style={[styles.debugKey, { color: colors.text }]}>
              {key}:
            </Text>
            <Text style={[styles.debugValue, { color: colors.textSecondary }]}>
              {typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)}
            </Text>
          </View>
        ))}
      </View>

      <View style={styles.statusContainer}>
        <Text style={[styles.statusTitle, { color: colors.text }]}>
          Status Summary:
        </Text>
        <Text style={[styles.statusText, { color: debugInfo.asyncStorageAvailable ? '#4CAF50' : '#F44336' }]}>
          AsyncStorage Available: {debugInfo.asyncStorageAvailable ? '✅ Yes' : '❌ No'}
        </Text>
        <Text style={[styles.statusText, { color: debugInfo.directAsyncStorageWorks ? '#4CAF50' : '#F44336' }]}>
          Direct Operations: {debugInfo.directAsyncStorageWorks ? '✅ Working' : '❌ Failed'}
        </Text>
        <Text style={[styles.statusText, { color: debugInfo.safeWrapperWorks ? '#4CAF50' : '#F44336' }]}>
          Safe Wrapper: {debugInfo.safeWrapperWorks ? '✅ Working' : '❌ Failed'}
        </Text>
        <Text style={[styles.statusText, { color: debugInfo.usingMemoryFallback ? '#FF9800' : '#4CAF50' }]}>
          Using Fallback: {debugInfo.usingMemoryFallback ? '⚠️ Yes' : '✅ No'}
        </Text>
      </View>
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
  },
  title: {
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 20,
    textAlign: 'center',
  },
  button: {
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    marginVertical: 8,
    alignItems: 'center',
  },
  buttonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  debugContainer: {
    marginTop: 20,
    padding: 16,
    backgroundColor: '#f5f5f5',
    borderRadius: 8,
  },
  debugRow: {
    marginBottom: 8,
  },
  debugKey: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 2,
  },
  debugValue: {
    fontSize: 12,
    fontFamily: 'monospace',
  },
  statusContainer: {
    marginTop: 20,
    padding: 16,
    backgroundColor: '#f5f5f5',
    borderRadius: 8,
  },
  statusTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  statusText: {
    fontSize: 14,
    marginBottom: 6,
  },
}); 