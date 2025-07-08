import AsyncStorage from '@react-native-async-storage/async-storage';
import { Platform } from 'react-native';

// In-memory fallback storage for when AsyncStorage is not available
class MemoryStorage {
  private storage: Map<string, string> = new Map();

  async getItem(key: string): Promise<string | null> {
    return this.storage.get(key) || null;
  }

  async setItem(key: string, value: string): Promise<void> {
    this.storage.set(key, value);
  }

  async removeItem(key: string): Promise<void> {
    this.storage.delete(key);
  }

  async multiRemove(keys: string[]): Promise<void> {
    keys.forEach(key => this.storage.delete(key));
  }

  async getAllKeys(): Promise<readonly string[]> {
    return Array.from(this.storage.keys());
  }

  async clear(): Promise<void> {
    this.storage.clear();
  }
}

// Safe wrapper for AsyncStorage with fallback to memory storage
class SafeStorage {
  private static instance: SafeStorage;
  private isInitialized = false;
  private useMemoryFallback = false;
  private memoryStorage = new MemoryStorage();

  static getInstance(): SafeStorage {
    if (!SafeStorage.instance) {
      SafeStorage.instance = new SafeStorage();
    }
    return SafeStorage.instance;
  }

  async initialize(): Promise<void> {
    if (this.isInitialized) return;
    
    try {
      // Test AsyncStorage availability
      if (AsyncStorage && typeof AsyncStorage.setItem === 'function') {
        await AsyncStorage.setItem('__test__', 'test');
        const result = await AsyncStorage.getItem('__test__');
        await AsyncStorage.removeItem('__test__');
        
        if (result === 'test') {
          this.isInitialized = true;
          this.useMemoryFallback = false;
          console.log('✅ AsyncStorage initialized successfully');
          return;
        }
      }
    } catch (error) {
      console.warn('⚠️ AsyncStorage not available, using memory fallback:', error);
    }
    
    // Fallback to memory storage
    this.useMemoryFallback = true;
    this.isInitialized = true;
    console.log('⚠️ Using memory storage fallback');
  }

  private async getStorage() {
    await this.initialize();
    return this.useMemoryFallback ? this.memoryStorage : AsyncStorage;
  }

  async getItem(key: string): Promise<string | null> {
    try {
      const storage = await this.getStorage();
      return await storage.getItem(key);
    } catch (error) {
      console.error(`Error getting item ${key}:`, error);
      // Try memory fallback if AsyncStorage fails
      if (!this.useMemoryFallback) {
        console.log('🔄 Falling back to memory storage for getItem');
        this.useMemoryFallback = true;
        return await this.memoryStorage.getItem(key);
      }
      return null;
    }
  }

  async setItem(key: string, value: string): Promise<void> {
    try {
      const storage = await this.getStorage();
      await storage.setItem(key, value);
    } catch (error) {
      console.error(`Error setting item ${key}:`, error);
      // Try memory fallback if AsyncStorage fails
      if (!this.useMemoryFallback) {
        console.log('🔄 Falling back to memory storage for setItem');
        this.useMemoryFallback = true;
        await this.memoryStorage.setItem(key, value);
      } else {
        throw error;
      }
    }
  }

  async removeItem(key: string): Promise<void> {
    try {
      const storage = await this.getStorage();
      await storage.removeItem(key);
    } catch (error) {
      console.error(`Error removing item ${key}:`, error);
      // Try memory fallback if AsyncStorage fails
      if (!this.useMemoryFallback) {
        console.log('🔄 Falling back to memory storage for removeItem');
        this.useMemoryFallback = true;
        await this.memoryStorage.removeItem(key);
      } else {
        throw error;
      }
    }
  }

  async multiRemove(keys: string[]): Promise<void> {
    try {
      const storage = await this.getStorage();
      await storage.multiRemove(keys);
    } catch (error) {
      console.error('Error removing multiple items:', error);
      // Try memory fallback if AsyncStorage fails
      if (!this.useMemoryFallback) {
        console.log('🔄 Falling back to memory storage for multiRemove');
        this.useMemoryFallback = true;
        await this.memoryStorage.multiRemove(keys);
      } else {
        throw error;
      }
    }
  }

  async getAllKeys(): Promise<readonly string[]> {
    try {
      const storage = await this.getStorage();
      return await storage.getAllKeys();
    } catch (error) {
      console.error('Error getting all keys:', error);
      // Try memory fallback if AsyncStorage fails
      if (!this.useMemoryFallback) {
        console.log('🔄 Falling back to memory storage for getAllKeys');
        this.useMemoryFallback = true;
        return await this.memoryStorage.getAllKeys();
      }
      return [];
    }
  }

  async clear(): Promise<void> {
    try {
      const storage = await this.getStorage();
      await storage.clear();
    } catch (error) {
      console.error('Error clearing storage:', error);
      // Try memory fallback if AsyncStorage fails
      if (!this.useMemoryFallback) {
        console.log('🔄 Falling back to memory storage for clear');
        this.useMemoryFallback = true;
        await this.memoryStorage.clear();
      } else {
        throw error;
      }
    }
  }

  // Method to check if we're using memory fallback
  isUsingMemoryFallback(): boolean {
    return this.useMemoryFallback;
  }
}

// Export singleton instance
export const safeStorage = SafeStorage.getInstance();

// Export convenience functions
export const getItem = (key: string) => safeStorage.getItem(key);
export const setItem = (key: string, value: string) => safeStorage.setItem(key, value);
export const removeItem = (key: string) => safeStorage.removeItem(key);
export const multiRemove = (keys: string[]) => safeStorage.multiRemove(keys);
export const getAllKeys = () => safeStorage.getAllKeys();
export const clear = () => safeStorage.clear();
export const isUsingMemoryFallback = () => safeStorage.isUsingMemoryFallback(); 