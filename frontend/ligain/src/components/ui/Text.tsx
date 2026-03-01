import React from 'react';
import { Text as RNText, TextProps } from 'react-native';

type Props = TextProps & { className?: string };

export function Text({ className, ...props }: Props) {
  return (
    <RNText
      className={`font-sans${className ? ` ${className}` : ''}`}
      {...props}
    />
  );
}
