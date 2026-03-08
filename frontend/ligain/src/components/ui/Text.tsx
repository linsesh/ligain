import React from 'react';
import { Text as RNText, TextProps } from 'react-native';

type Props = TextProps & { className?: string };

export function Text({ className, ...props }: Props) {
  const hasCustomFont = className && /font-hk-/.test(className);
  return (
    <RNText
      className={hasCustomFont ? className : `font-sans${className ? ` ${className}` : ''}`}
      {...props}
    />
  );
}
