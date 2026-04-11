import React from 'react';
import { View, Image } from 'react-native';
import { getTeamLogo, isPngLogo } from '../../utils/teamLogos';

interface TeamLogoProps {
  teamName: string;
  size?: number;
}

export function TeamLogo({ teamName, size = 56 }: TeamLogoProps) {
  const Logo = getTeamLogo(teamName);
  if (!Logo) return <View style={{ width: size, height: size }} />;
  return isPngLogo(Logo) ? (
    <Image source={Logo} style={{ width: size, height: size }} resizeMode="contain" />
  ) : (
    <Logo width={size} height={size} />
  );
}
