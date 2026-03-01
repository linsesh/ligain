import React from 'react';
import { View } from 'react-native';
import { Text } from './Text';

type SectionCardProps = {
  title?: string;
  children: React.ReactNode;
  className?: string;
};

export function SectionCard({ title, children, className }: SectionCardProps) {
  return (
    <View className={`rounded-2xl bg-background p-4 ${className ?? ''}`}>
      {title && (
        <Text className="text-xs font-semibold text-foreground-secondary tracking-widest uppercase mb-3">
          {title}
        </Text>
      )}
      {children}
    </View>
  );
}
