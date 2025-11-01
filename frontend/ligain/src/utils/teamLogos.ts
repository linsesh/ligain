// Team logo mapping utility
// Maps team names to their corresponding logo assets

import React from 'react';

// Import SVG files directly
import AngersLogoSvg from '../../assets/images/logo_angers.svg';
import AuxerreLogoSvg from '../../assets/images/logo_auxerre.svg';
import BrestLogoSvg from '../../assets/images/logo_brest.svg';
import LeHavreLogoSvg from '../../assets/images/logo_le_havre.svg';
import LensLogoSvg from '../../assets/images/logo_lens.svg';
import LilleLogoSvg from '../../assets/images/logo_lille.svg';
import LorientLogoSvg from '../../assets/images/logo_lorient.svg';
import MetzLogoSvg from '../../assets/images/logo_metz.svg';
import MonacoLogoSvg from '../../assets/images/logo_monaco.svg';
import NantesLogoSvg from '../../assets/images/logo_nantes.svg';
import NiceLogoSvg from '../../assets/images/logo_nice.svg';
import OLLogoSvg from '../../assets/images/logo_ol.svg';
import OMLogoSvg from '../../assets/images/logo_om.svg';
import PFCLogoSvg from '../../assets/images/logo_pfc.svg';
import PSGLogoSvg from '../../assets/images/logo_psg.svg';
import RennesLogoSvg from '../../assets/images/logo_rennes.svg';
import StrasbourgLogoSvg from '../../assets/images/logo_strasbourg.svg';
import ToulouseLogoSvg from '../../assets/images/logo_toulouse.svg';

export const TEAM_LOGOS: { [key: string]: React.ComponentType<any> } = {
  // Ligue 1 teams - Original names from backend
  'Angers SCO': AngersLogoSvg,
  'Auxerre': AuxerreLogoSvg,
  'Brest': BrestLogoSvg,
  'Le Havre': LeHavreLogoSvg,
  'Lens': LensLogoSvg,
  'LOSC Lille': LilleLogoSvg,
  'Lorient': LorientLogoSvg,
  'Metz': MetzLogoSvg,
  'Monaco': MonacoLogoSvg,
  'Nantes': NantesLogoSvg,
  'Nice': NiceLogoSvg,
  'Olympique Lyonnais': OLLogoSvg,
  'Olympique Marseille': OMLogoSvg,
  'Paris': PFCLogoSvg,
  'Paris Saint Germain': PSGLogoSvg,
  'Rennes': RennesLogoSvg,
  'Strasbourg': StrasbourgLogoSvg,
  'Toulouse': ToulouseLogoSvg,
  
  // Display names (shortened versions)
  'Angers': AngersLogoSvg,
  'Lille': LilleLogoSvg,
  'Lyon': OLLogoSvg,
  'Marseille': OMLogoSvg,
  'Paris FC': PFCLogoSvg,
  'PSG': PSGLogoSvg,
};

/**
 * Get the logo component for a team by name
 * @param teamName - The name of the team
 * @returns The logo component or null if not found
 */
export const getTeamLogo = (teamName: string): React.ComponentType<any> | null => {
  return TEAM_LOGOS[teamName] || null;
};

/**
 * Check if a team has a logo available
 * @param teamName - The name of the team
 * @returns True if logo is available, false otherwise
 */
export const hasTeamLogo = (teamName: string): boolean => {
  return teamName in TEAM_LOGOS;
};