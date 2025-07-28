import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

export interface StatusTagProps {
  text: string;
  variant: 'warning' | 'success' | 'primary' | 'finished' | 'negative';
  style?: any;
  textStyle?: any;
}

/**
 * Shared StatusTag component for displaying status information
 * Used across the app for consistent status display
 */
export default function StatusTag({ text, variant, style, textStyle }: StatusTagProps) {
  const baseStyle = [styles.statusTag];
  let variantStyle = null;
  
  switch (variant) {
    case 'success':
      variantStyle = styles.successTag;
      break;
    case 'warning':
      variantStyle = styles.inProgressTag;
      break;
    case 'finished':
      variantStyle = styles.finishedTag;
      break;
    case 'primary':
      variantStyle = styles.primaryTag;
      break;
    case 'negative':
      variantStyle = styles.negativeTag;
      break;
  }
  
  return (
    <View style={[...baseStyle, variantStyle, style]}>
      <Text style={[styles.statusTagText, textStyle]}>{text}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  statusTag: {
    paddingHorizontal: 10,
    paddingVertical: 5,
    borderRadius: 8,
    alignSelf: 'flex-start',
  },
  statusTagText: {
    fontSize: 12,
    fontWeight: 'bold',
    color: '#fff',
  },
  successTag: {
    backgroundColor: '#4CAF50',
  },
  inProgressTag: {
    backgroundColor: '#FFC107',
  },
  finishedTag: {
    backgroundColor: '#9E9E9E',
  },
  primaryTag: {
    backgroundColor: '#4CAF50',
  },
  negativeTag: {
    backgroundColor: '#ff6b6b',
  },
}); 