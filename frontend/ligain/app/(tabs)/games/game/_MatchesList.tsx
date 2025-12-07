import React, { useState, useEffect, useRef } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl, KeyboardAvoidingView, Platform, Animated, Dimensions } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useMatches } from '../../../../hooks/useMatches';
import { useBetSubmission } from '../../../../hooks/useBetSubmission';
import { useBetSynchronization } from '../../../../hooks/useBetSynchronization';
import { useMatchNotifications } from '../../../../src/hooks/useMatchNotifications';
import { useAuth } from '../../../../src/contexts/AuthContext';
import { useGames } from '../../../../src/contexts/GamesContext';
import { useTranslation } from 'react-i18next';
import { formatTime, formatDate } from '../../../../src/utils/dateUtils';
import { colors } from '../../../../src/constants/colors';
import { sharedStyles } from '../../../../src/constants/sharedStyles';
import StatusTag from '../../../../src/components/StatusTag';
import { BetSyncModal, SyncResult } from '../../../../src/components/BetSyncModal';
import ShareableMatchResult from '../../../../src/components/ShareableMatchResult';
import { captureAndShareWithOptions, formatDateForShare } from '../../../../src/utils/shareUtils';
import ViewShot from 'react-native-view-shot';
import { getTeamLogo } from '../../../../src/utils/teamLogos';

interface TempScores {
  [key: string]: {
    home?: number;
    away?: number;
  };
}

function TeamInput({ teamName, value, onChange, canModify, isAway = false, onFocus, onBlur, isFavorite }: {
  teamName: string;
  value: string;
  onChange: (value: string) => void;
  canModify: boolean;
  isAway?: boolean;
  onFocus?: () => void;
  onBlur?: () => void;
  isFavorite?: boolean;
}) {
  const teamLogo = getTeamLogo(teamName);
  const TeamLogo = teamLogo;
  
  return (
    <View style={[styles.teamContainer, isAway && styles.awayTeamContainer]}>
      <View style={[styles.logoAndNameContainer, isAway && styles.awayLogoAndNameContainer]}>
        {TeamLogo && (
          <TeamLogo width={40} height={40} />
        )}
        <Text style={[styles.teamName, isAway && styles.awayTeamName]}>
          {teamName}
          {isFavorite && <Text style={{ color: colors.primary }}>⭐</Text>}
        </Text>
      </View>
      <View style={[styles.inputContainer, isAway && styles.awayInputContainer]}>
        <TextInput
          style={[styles.betInput, !canModify && styles.disabledInput]}
          value={value}
          onChangeText={onChange}
          editable={canModify}
          keyboardType="numeric"
          onFocus={onFocus}
          onBlur={onBlur}
          maxLength={2}
        />
      </View>
    </View>
  );
}

function MatchCard({ matchResult, tempScores, expandedMatches, onBetChange, onToggleBetSection, onFocus, onBlur, onDone, onRef, gameId }: {
  matchResult: any;
  tempScores: TempScores;
  expandedMatches: { [key: string]: boolean };
  onBetChange: (matchId: string, team: 'home' | 'away', value: string) => void;
  onToggleBetSection: (matchId: string) => void;
  onFocus: () => void;
  onBlur: () => void;
  onDone?: () => void;
  onRef?: (ref: View | null) => void;
  gameId: string;
}) {
  const { player } = useAuth();
  const { games } = useGames();
  const { t } = useTranslation();
  const [isSharing, setIsSharing] = useState(false);
  const shareableRef = useRef<ViewShot>(null);
  const now = new Date();
  const isFuture = !matchResult.match.isFinished() && !matchResult.match.isInProgress();
  const userBet = player && matchResult.bets ? matchResult.bets[player.id] : undefined;
  const canModify = isFuture && (userBet?.isModifiable(now) !== false);
  
  // Get game name from context
  const game = games.find(g => g.gameId === gameId);
  const gameName = game?.name || 'Ligain Game';

  const handleShareMatch = async () => {
    if (!matchResult.match.isFinished() || isSharing) return;
    
    setIsSharing(true);
    try {
      await captureAndShareWithOptions(shareableRef, {
        title: t('share.shareTitle'),
        message: t('share.shareTitle'),
      });
    } catch (error) {
      console.error('Error sharing match:', error);
      Alert.alert(t('share.shareFailed'), t('share.shareFailed'));
    } finally {
      setIsSharing(false);
    }
  };

  // Tag logic
  let tagText: string | null = null;
  let tagVariant: 'warning' | 'success' | 'negative' | 'finished' | 'primary' | null = null;
  let hasTag = false;
  if (matchResult.match.isInProgress()) {
    tagText = t('games.inProgressTag');
    tagVariant = 'primary';
    hasTag = true;
  } else if (matchResult.match.isFinished() && player) {
    if (matchResult.scores && matchResult.scores[player.id]) {
      const points = matchResult.scores[player.id].points;
      if (typeof points === 'number' && points > 0 && points < 1000) {
        tagText = `+${points} points`;
        tagVariant = 'success';
        hasTag = true;
      } else if (typeof points === 'number' && points < 0) {
        tagText = `${points} points (${t('games.negativePointsTag')})`;
        tagVariant = 'negative';
        hasTag = true;
      } else {
        tagText = `0 points`;
        tagVariant = 'finished';
        hasTag = true;
      }
    } else {
      // No score entry means no bet was placed, show 0 points
      tagText = `0 points`;
      tagVariant = 'finished';
      hasTag = true;
    }
  }

  // Determine card style: always greyed out if finished
  const cardStyle = [
    styles.matchCard,
    matchResult.match.isFinished() ? styles.finishedMatchCard : null,
    hasTag ? styles.matchCardWithTag : null
  ].filter(Boolean);

  return (
    <View 
      key={matchResult.match.id()} 
      style={cardStyle}
      ref={onRef}
    >
      {/* Status Tag and Share Button */}
      {matchResult.match.isFinished() && (
        <View style={styles.topLeftContainer}>
          <TouchableOpacity
            style={sharedStyles.shareButton}
            onPress={handleShareMatch}
            disabled={isSharing}
          >
            <Ionicons 
              name={isSharing ? "hourglass-outline" : "share-outline"} 
              size={20} 
              color={isSharing ? colors.textSecondary : "#000000"} 
            />
          </TouchableOpacity>
        </View>
      )}
      <View style={styles.topRightContainer}>
        {tagText && typeof tagVariant === 'string' && (
          <StatusTag text={tagText} variant={tagVariant} style={styles.statusTag} />
        )}
      </View>
      <View style={styles.bettingContainer}>
        {/* Only show the current user's bet for future matches */}
        {isFuture ? (
          <>
            <TeamInput
              teamName={matchResult.match.getHomeTeam()}
              value={tempScores[matchResult.match.id()]?.home?.toString() || ''}
              onChange={(value) => onBetChange(matchResult.match.id(), 'home', value)}
              canModify={canModify}
              onFocus={() => onFocus()}
              onBlur={() => onBlur()}
              isFavorite={matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getHomeTeam()}
            />
            <Text style={styles.vsText}>{t('common.vs')}</Text>
            <TeamInput
              teamName={matchResult.match.getAwayTeam()}
              value={tempScores[matchResult.match.id()]?.away?.toString() || ''}
              onChange={(value) => onBetChange(matchResult.match.id(), 'away', value)}
              canModify={canModify}
              isAway
              onFocus={() => onFocus()}
              onBlur={() => onBlur()}
              isFavorite={matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getAwayTeam()}
            />
          </>
        ) : (
          <>
            <TeamInput
              teamName={matchResult.match.getHomeTeam()}
              value={matchResult.match.getHomeGoals().toString()}
              onChange={() => {}}
              canModify={false}
              isFavorite={matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getHomeTeam()}
            />
            <Text style={styles.vsText}>{t('common.vs')}</Text>
            <TeamInput
              teamName={matchResult.match.getAwayTeam()}
              value={matchResult.match.getAwayGoals().toString()}
              onChange={() => {}}
              canModify={false}
              isAway
              isFavorite={matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getAwayTeam()}
            />
          </>
        )}
      </View>
      {/* Odds display */}
      <View style={styles.oddsContainer}>
        <View style={styles.oddsRow}>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>1</Text>
            <Text style={styles.oddsValue}>
              {matchResult.match.getHomeTeamOdds() === 0 ? '-' : matchResult.match.getHomeTeamOdds().toFixed(2)}
            </Text>
            {matchResult.match.hasClearFavorite() && (
              matchResult.match.getFavoriteTeam() === matchResult.match.getHomeTeam()
                ? <Text style={styles.outsiderMark}>x1</Text>
                : <Text style={styles.outsiderMark}>×2</Text>
            )}
          </View>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>N</Text>
            <Text style={styles.oddsValue}>
              {matchResult.match.getDrawOdds() === 0 ? '-' : matchResult.match.getDrawOdds().toFixed(2)}
            </Text>
            {matchResult.match.hasClearFavorite() && (
              <Text style={styles.outsiderMark}>×1.5</Text>
            )}
          </View>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>2</Text>
            <Text style={styles.oddsValue}>
              {matchResult.match.getAwayTeamOdds() === 0 ? '-' : matchResult.match.getAwayTeamOdds().toFixed(2)}
            </Text>
            {matchResult.match.hasClearFavorite() && (
              matchResult.match.getFavoriteTeam() === matchResult.match.getAwayTeam()
                ? <Text style={styles.outsiderMark}>x1</Text>
                : <Text style={styles.outsiderMark}>×2</Text>
            )}
          </View>
        </View>
      </View>
      {/* Keep the rest of the card (other players' bets/scores) only for past/in-progress matches */}
      {(!isFuture && (matchResult.match.isFinished() || matchResult.match.isInProgress())) && (
        <>
          <TouchableOpacity 
            style={styles.toggleButton} 
            onPress={() => onToggleBetSection(matchResult.match.id())}
          >
            <Text style={styles.toggleButtonText}>
              {t('games.playersBets')}
            </Text>
            <Ionicons 
              name={expandedMatches[matchResult.match.id()] ? "chevron-up" : "chevron-down"} 
              size={24} 
              color="#333" 
            />
          </TouchableOpacity>
          {expandedMatches[matchResult.match.id()] && (
            <View style={styles.betResultContainer}>
              {matchResult.scores ? (
                <View style={styles.scoresContainer}>
                  {/* Custom order: current user first, then others */}
                  {(() => {
                    const entries = Object.entries(matchResult.scores);
                    let userRow = null;
                    let otherRows: React.ReactNode[] = [];
                    entries.forEach(([playerId, scoreData]) => {
                      const s = scoreData as { playerName: string; points: number };
                      if (player && playerId === player.id) {
                        userRow = (
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {s.playerName} ({t('common.me')}): {s.points} points </Text>
                            {matchResult.bets?.[playerId] ? (
                              <Text style={styles.betResultText}> ({matchResult.bets[playerId].predictedHomeGoals} - {matchResult.bets[playerId].predictedAwayGoals}) </Text>
                            ) : (
                              <Text style={[styles.betResultText, { fontStyle: 'italic', color: '#999' }]}> ({t('games.negativePointsTag')}) </Text>
                            )}
                          </View>
                        );
                      } else {
                        otherRows.push(
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={styles.scoreText}> {s.playerName}: {s.points} points </Text>
                            {matchResult.bets?.[playerId] ? (
                              <Text style={styles.betResultText}> ({matchResult.bets[playerId].predictedHomeGoals} - {matchResult.bets[playerId].predictedAwayGoals}) </Text>
                            ) : (
                              <Text style={[styles.betResultText, { fontStyle: 'italic', color: '#999' }]}> ({t('games.negativePointsTag')}) </Text>
                            )}
                          </View>
                        );
                      }
                    });
                    if (!userRow && player) {
                      userRow = (
                        <View key={player.id} style={styles.playerScoreRow}>
                          <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {player.name} ({t('common.me')}): <Text style={{ fontStyle: 'italic', color: '#999' }}>{t('games.negativePointsTag')}</Text> </Text>
                        </View>
                      );
                    }
                    return [userRow, ...otherRows];
                  })()}
                </View>
              ) : matchResult.bets && (
                <View style={styles.scoresContainer}>
                  {(() => {
                    const entries = Object.entries(matchResult.bets);
                    let userRow = null;
                    let otherRows: React.ReactNode[] = [];
                    entries.forEach(([playerId, betData]) => {
                      const b = betData as { playerName: string; predictedHomeGoals: number; predictedAwayGoals: number };
                      if (player && playerId === player.id) {
                        userRow = (
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {b.playerName} ({t('common.me')}): </Text>
                            <Text style={styles.betResultText}> ({b.predictedHomeGoals} - {b.predictedAwayGoals}) </Text>
                          </View>
                        );
                      } else {
                        otherRows.push(
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={styles.scoreText}> {b.playerName}: </Text>
                            <Text style={styles.betResultText}> ({b.predictedHomeGoals} - {b.predictedAwayGoals}) </Text>
                          </View>
                        );
                      }
                    });
                    if (!userRow && player) {
                      userRow = (
                        <View key={player.id} style={styles.playerScoreRow}>
                          <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {player.name} ({t('common.me')}): <Text style={{ fontStyle: 'italic', color: '#999' }}>{t('games.negativePointsTag')}</Text> </Text>
                        </View>
                      );
                    }
                    return [userRow, ...otherRows];
                  })()}
                </View>
              )}
            </View>
          )}
        </>
      )}
      
      {/* Hidden shareable component for image generation */}
      {matchResult.match.isFinished() && (
        <View style={{ position: 'absolute', left: -9999, top: -9999 }}>
          <ViewShot ref={shareableRef}>
            <ShareableMatchResult
              homeTeam={matchResult.match.getHomeTeam()}
              awayTeam={matchResult.match.getAwayTeam()}
              homeScore={matchResult.match.getHomeGoals()}
              awayScore={matchResult.match.getAwayGoals()}
              myHomeScore={player ? matchResult.bets?.[player.id]?.predictedHomeGoals : undefined}
              myAwayScore={player ? matchResult.bets?.[player.id]?.predictedAwayGoals : undefined}
              date={formatDateForShare(matchResult.match.getDate())}
              players={Object.entries(matchResult.scores || {}).map(([playerId, scoreData]: [string, any]) => ({
                name: scoreData.playerName,
                points: scoreData.points,
                bet: matchResult.bets?.[playerId] ? 
                  `${matchResult.bets[playerId].predictedHomeGoals}-${matchResult.bets[playerId].predictedAwayGoals}` : 
                  undefined
              }))}
              gameName={gameName}
            />
          </ViewShot>
        </View>
      )}
    </View>
  );
}

interface MatchesListProps {
  gameId: string;
  initialMatchday?: number;
}

export default function MatchesList({ gameId, initialMatchday }: MatchesListProps) {
  const { incomingMatches, pastMatches, loading: matchesLoading, error: matchesError, refresh } = useMatches(gameId);
  // Automatically manage match notifications (schedule/cancel based on bets)
  useMatchNotifications(gameId);
  const [tempScores, setTempScores] = useState<TempScores>({});
  const [expandedMatches, setExpandedMatches] = useState<{ [key: string]: boolean }>({});
  const [refreshing, setRefreshing] = useState(false);
  const [editingMatchId, setEditingMatchId] = useState<string | null>(null);
  const [currentMatchday, setCurrentMatchday] = useState<number | null>(initialMatchday ?? null);
  const { player } = useAuth();
  const { t } = useTranslation();
  const scrollViewRef = React.useRef<ScrollView>(null);
  const [keyboardHeight, setKeyboardHeight] = useState(0);
  const matchCardRefs = React.useRef<{ [key: string]: View | null }>({});

  // Bet synchronization state
  const { syncOpportunity, loading: syncLoading, refetch: refetchSync } = useBetSynchronization(gameId);
  const [showSyncModal, setShowSyncModal] = useState(false);
  const [syncModalShownForGame, setSyncModalShownForGame] = useState<string | null>(null);
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncResult, setSyncResult] = useState<SyncResult | null>(null);
  const [syncModalMode, setSyncModalMode] = useState<'initial' | 'success' | 'partialSuccess' | 'failure'>('initial');
  const [failedBetsToRetry, setFailedBetsToRetry] = useState<SyncResult['failed']>([]);

  // Combine incoming and past matches
  const matches = [...Object.values(incomingMatches), ...Object.values(pastMatches)];

  // Group matches by matchday
  const matchesByMatchday = matches.reduce((acc, matchResult) => {
    const matchday = matchResult.match.getMatchday();
    if (!acc[matchday]) {
      acc[matchday] = [];
    }
    acc[matchday].push(matchResult);
    return acc;
  }, {} as { [key: number]: any[] });

  // Sort matchdays
  const sortedMatchdays = Object.keys(matchesByMatchday)
    .map(Number)
    .sort((a, b) => a - b);

  // Set initial matchday if not set
  useEffect(() => {
    if (sortedMatchdays.length > 0 && currentMatchday === null) {
      if (initialMatchday && sortedMatchdays.includes(initialMatchday)) {
        setCurrentMatchday(initialMatchday);
      } else {
        setCurrentMatchday(sortedMatchdays[0]);
      }
    }
    // Only run on mount or when sortedMatchdays changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sortedMatchdays]);

  // Listen to keyboard events to track keyboard height
  useEffect(() => {
    const keyboardDidShowListener = Keyboard.addListener('keyboardDidShow', (e) => {
      setKeyboardHeight(e.endCoordinates.height);
    });
    const keyboardDidHideListener = Keyboard.addListener('keyboardDidHide', () => {
      setKeyboardHeight(0);
    });

    return () => {
      keyboardDidShowListener?.remove();
      keyboardDidHideListener?.remove();
    };
  }, []);

  // Group matches by date and time within the current matchday
  const getMatchesByTime = (matchday: number) => {
    const matchdayMatches = matchesByMatchday[matchday] || [];
    // Sort by date and time (ascending)
    const sortedMatches = matchdayMatches.sort((a, b) => 
      a.match.getDate().getTime() - b.match.getDate().getTime()
    );
    // Group by date and time
    const groupedByDateTime = sortedMatches.reduce((acc, matchResult) => {
      const date = matchResult.match.getDate();
      const dateKey = formatDate(date);
      const timeKey = formatTime(date);
      const dateTimeKey = `${dateKey} - ${timeKey}`;
      
      if (!acc[dateTimeKey]) {
        acc[dateTimeKey] = [];
      }
      acc[dateTimeKey].push(matchResult);
      return acc;
    }, {} as { [key: string]: any[] });
    return groupedByDateTime;
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  // Initialize tempScores with existing bets
  useEffect(() => {
    const initialTempScores: TempScores = {};
    [...Object.values(incomingMatches), ...Object.values(pastMatches)].forEach((matchResult: any) => {
      if (player && matchResult.bets && matchResult.bets[player.id]) {
        const userBet = matchResult.bets[player.id];
        initialTempScores[matchResult.match.id()] = {
          home: userBet.predictedHomeGoals,
          away: userBet.predictedAwayGoals
        };
      }
    });
    setTempScores(initialTempScores);
  }, [incomingMatches, pastMatches, player]);

  // Helper to reset tempScores for a match to the last server value
  const resetTempScoreToServer = (matchId: string) => {
    const matchResult = matches.find((m: any) => m.match.id() === matchId);
    if (!matchResult || !player) return;
    if (matchResult.bets && matchResult.bets[player.id]) {
      const userBet = matchResult.bets[player.id];
      setTempScores(prev => ({
        ...prev,
        [matchId]: {
          home: userBet.predictedHomeGoals,
          away: userBet.predictedAwayGoals
        }
      }));
    } else {
      // No bet on server, clear temp score
      setTempScores(prev => {
        const newScores = { ...prev };
        delete newScores[matchId];
        return newScores;
      });
    }
  };

  // Use the new onFail callback in useBetSubmission
  const { submitBet, isSubmitting, error: submitError } = useBetSubmission(gameId, resetTempScoreToServer);

  useEffect(() => {
    if (submitError) {
      Alert.alert(
        t('games.betNotSaved'),
        t('games.betNotSavedMessage'),
        [{ text: t('common.ok') }]
      );
    }
  }, [submitError, t]);

  // Show sync modal when opportunity exists and not shown for this game
  useEffect(() => {
    if (syncOpportunity && syncModalShownForGame !== gameId) {
      setShowSyncModal(true);
      setSyncModalShownForGame(gameId);
    }
  }, [syncOpportunity, gameId, syncModalShownForGame]);

  // Reset sync modal state when game changes
  useEffect(() => {
    setSyncModalShownForGame(null);
    setShowSyncModal(false);
    setIsSyncing(false);
    setSyncResult(null);
    setSyncModalMode('initial');
    setFailedBetsToRetry([]);
  }, [gameId]);

  const toggleBetSection = (matchId: string) => {
    setExpandedMatches(prev => ({
      ...prev,
      [matchId]: !prev[matchId]
    }));
  };

  const handleBetChange = async (matchId: string, team: 'home' | 'away', value: string) => {
    const matchResult = matches.find((m: any) => m.match.id() === matchId);
    if (!matchResult) return;
    const currentTempScores = tempScores[matchId] || {};
    const newTempScores = {
      ...currentTempScores,
      [team]: value && value.trim() !== '' ? parseInt(value) : undefined
    };
    setTempScores(prev => ({
      ...prev,
      [matchId]: newTempScores
    }));
    if (newTempScores.home !== undefined && newTempScores.away !== undefined) {
      const homeGoals = Number(newTempScores.home);
      const awayGoals = Number(newTempScores.away);
      if (!isNaN(homeGoals) && !isNaN(awayGoals) && homeGoals >= 0 && awayGoals >= 0) {
        await submitBet(matchId, homeGoals, awayGoals);
        await refresh();
      }
    }
  };

  const handleFocus = (matchId: string) => {
    setEditingMatchId(matchId);
    
    // Get all matches for the current matchday
    const currentMatchdayMatches = currentMatchday ? getMatchesByTime(currentMatchday) : {};
    const allMatches = Object.values(currentMatchdayMatches).flat() as any[];
    
    // Check if this is the last match
    const lastMatch = allMatches[allMatches.length - 1];
    
    // Only scroll if this is the last match
    if (lastMatch && lastMatch.match.id() === matchId && scrollViewRef.current) {
      setTimeout(() => {
        scrollViewRef.current?.scrollToEnd({ animated: true });
      }, 100);
    }
  };

  const handleBlur = () => {
    setEditingMatchId(null);
  };

  const handleDone = () => {
    setEditingMatchId(null);
    Keyboard.dismiss();
  };

  // Sync handlers
  const handleSyncSynchronize = async () => {
    if (!syncOpportunity) return;
    
    setIsSyncing(true);
    const results: SyncResult = { successful: [], failed: [] };
    
    try {
      // Submit bets for each match to sync individually
      for (const matchToSync of syncOpportunity.matchesToSync) {
        try {
          await submitBet(
            matchToSync.matchId,
            matchToSync.predictedHomeGoals,
            matchToSync.predictedAwayGoals
          );
          results.successful.push(matchToSync);
        } catch (error) {
          console.error(`Failed to sync bet for match ${matchToSync.matchId}:`, error);
          results.failed.push({ match: matchToSync, error });
        }
      }
      
      // Always refresh to show current state
      await refresh();
      
      // Set the results and determine modal mode
      setSyncResult(results);
      
      if (results.failed.length === 0) {
        // All successful - close modal immediately (current behavior)
        setShowSyncModal(false);
        setSyncResult(null);
        setSyncModalMode('initial');
        setFailedBetsToRetry([]);
        return;
      } else if (results.successful.length > 0) {
        // Partial success
        setSyncModalMode('partialSuccess');
        setFailedBetsToRetry(results.failed);
      } else {
        // All failed
        setSyncModalMode('failure');
      }
    } catch (error) {
      console.error('Error during sync process:', error);
      setSyncModalMode('failure');
      setSyncResult({ successful: [], failed: [] });
    } finally {
      setIsSyncing(false);
    }
  };

  const handleSyncNotNow = () => {
    setShowSyncModal(false);
    setSyncResult(null);
    setSyncModalMode('initial');
    setFailedBetsToRetry([]);
  };




  const handleRetryFailed = async () => {
    if (failedBetsToRetry.length === 0) return;
    
    setIsSyncing(true);
    const retryResults: SyncResult = { successful: [], failed: [] };
    
    try {
      // Retry only the failed bets
      for (const failedBet of failedBetsToRetry) {
        try {
          await submitBet(
            failedBet.match.matchId,
            failedBet.match.predictedHomeGoals,
            failedBet.match.predictedAwayGoals
          );
          retryResults.successful.push(failedBet.match);
        } catch (error) {
          console.error(`Retry failed for match ${failedBet.match.matchId}:`, error);
          retryResults.failed.push(failedBet);
        }
      }
      
      // Refresh to show updated state
      await refresh();
      
      // Update results with retry results
      const updatedResults: SyncResult = {
        successful: [...(syncResult?.successful || []), ...retryResults.successful],
        failed: retryResults.failed
      };
      
      setSyncResult(updatedResults);
      
      if (retryResults.failed.length === 0) {
        setShowSyncModal(false);
        setSyncResult(null);
        setSyncModalMode('initial');
        setFailedBetsToRetry([]);
        return;
      } else {
        setSyncModalMode('partialSuccess');
        setFailedBetsToRetry(retryResults.failed);
      }
    } catch (error) {
      console.error('Error during retry:', error);
      setSyncModalMode('failure');
    } finally {
      setIsSyncing(false);
    }
  };

  const navigateMatchday = (direction: 'prev' | 'next') => {
    if (!currentMatchday) return;
    const currentIndex = sortedMatchdays.indexOf(currentMatchday);
    if (direction === 'prev' && currentIndex > 0) {
      setCurrentMatchday(sortedMatchdays[currentIndex - 1]);
    } else if (direction === 'next' && currentIndex < sortedMatchdays.length - 1) {
      setCurrentMatchday(sortedMatchdays[currentIndex + 1]);
    }
  };

  // Find the next closest unbet match in the current matchday
  const getNextUnbetMatchId = () => {
    if (!currentMatchday) return null;
    const matchesInDay = matchesByMatchday[currentMatchday] || [];
    const now = new Date();
    // Only consider matches in the future or in progress
    const unbetMatches = matchesInDay.filter(m => {
      const matchDate = m.match.getDate();
      const isFutureOrInProgress = !m.match.isFinished() || matchDate > now;
      const hasBet = player && m.bets && m.bets[player.id];
      return isFutureOrInProgress && !hasBet;
    });
    if (unbetMatches.length === 0) return null;
    // Find the closest in time to now
    unbetMatches.sort((a, b) => a.match.getDate().getTime() - b.match.getDate().getTime());
    return unbetMatches[0].match.id();
  };

  const nextUnbetMatchId = getNextUnbetMatchId();

  // Scroll to the next unbet match
  const scrollToNextMatch = () => {
    if (!nextUnbetMatchId || !scrollViewRef.current) return;
    // Find the y offset of the next match card
    // For simplicity, scroll to top (could be improved with refs per card)
    scrollViewRef.current.scrollTo({ y: 0, animated: true });
  };

  if (matchesLoading && !refreshing) {
    return (
      <View style={[styles.container, { backgroundColor: colors.loadingBackground }]}> 
        <ActivityIndicator testID="loading-indicator" size="large" color={colors.primary} />
      </View>
    );
  }

  const currentMatchdayMatches = currentMatchday ? getMatchesByTime(currentMatchday) : {};
  const sortedDateTimeKeys = Object.keys(currentMatchdayMatches).sort((a, b) => {
    // Extract date and time from the key format "Monday, January 15 - 20:00"
    const dateTimeA = a.split(' - ');
    const dateTimeB = b.split(' - ');
    
    if (dateTimeA.length !== 2 || dateTimeB.length !== 2) {
      return a.localeCompare(b); // Fallback to string comparison
    }
    
    const dateA = dateTimeA[0];
    const timeA = dateTimeA[1];
    const dateB = dateTimeB[0];
    const timeB = dateTimeB[1];
    
    // First compare by date
    if (dateA !== dateB) {
      // Parse dates for comparison (assuming format like "Monday, January 15")
      const dateAObj = new Date(dateA);
      const dateBObj = new Date(dateB);
      return dateAObj.getTime() - dateBObj.getTime();
    }
    
    // If same date, compare by time
    const timeANum = parseInt(timeA.replace(':', ''));
    const timeBNum = parseInt(timeB.replace(':', ''));
    return timeANum - timeBNum;
  });

  // Show legend if not editing a match or if the editing match has a clear favorite
  const editingMatch = editingMatchId ? matches.find(m => m.match.id() === editingMatchId) : null;
  const shouldShowLegend = !editingMatchId || (editingMatch && editingMatch.match.hasClearFavorite());

  return (
    <KeyboardAvoidingView 
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <ScrollView 
        ref={scrollViewRef}
        style={styles.scrollView}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
            progressBackgroundColor="#25292e"
            progressViewOffset={20}
          />
        }
      >
        {/* Matchday Navigation */}
        <View style={styles.matchdayNavigation}>
          <TouchableOpacity 
            style={[
              styles.navButton, 
              sortedMatchdays.indexOf(currentMatchday!) <= 0 && styles.navButtonDisabled
            ]}
            onPress={() => navigateMatchday('prev')}
            disabled={sortedMatchdays.indexOf(currentMatchday!) <= 0}
            testID="prev-matchday-button"
          >
            <Ionicons name="chevron-back" size={24} color="#fff" />
          </TouchableOpacity>
          <View style={styles.matchdayInfo}>
            <Text style={styles.matchdayTitle}>{t('games.matchday')} {currentMatchday}</Text>
            {currentMatchday && matchesByMatchday[currentMatchday] && matchesByMatchday[currentMatchday].length > 0 && (
              <Text style={styles.matchdayDate}>
                {formatDate(matchesByMatchday[currentMatchday][0].match.getDate())}
              </Text>
            )}
          </View>
          <TouchableOpacity 
            style={[
              styles.navButton, 
              sortedMatchdays.indexOf(currentMatchday!) >= sortedMatchdays.length - 1 && styles.navButtonDisabled
            ]}
            onPress={() => navigateMatchday('next')}
            disabled={sortedMatchdays.indexOf(currentMatchday!) >= sortedMatchdays.length - 1}
            testID="next-matchday-button"
          >
            <Ionicons name="chevron-forward" size={24} color="#fff" />
          </TouchableOpacity>
        </View>
        {/* Matches for current matchday */}
        <View style={styles.matchesContainer}>
          {sortedDateTimeKeys.map((dateTimeKey: string) => {
            const matchesAtTime = currentMatchdayMatches[dateTimeKey];
            const dateTimeParts = dateTimeKey.split(' - ');
            const dateDisplay = dateTimeParts[0] || '';
            const timeDisplay = dateTimeParts[1] || '';
            return (
              <View key={dateTimeKey} style={styles.timeGroup}>
                <View style={styles.timeHeaderContainer}>
                  <Text style={styles.timeHeader}>{timeDisplay}</Text>
                  <Text style={styles.dayHeader}>{dateDisplay}</Text>
                </View>
                {matchesAtTime.map((matchResult: any) => (
                  <MatchCard
                    key={matchResult.match.id()}
                    matchResult={matchResult}
                    tempScores={tempScores}
                    expandedMatches={expandedMatches}
                    onBetChange={handleBetChange}
                    onToggleBetSection={toggleBetSection}
                    onFocus={() => handleFocus(matchResult.match.id())}
                    onBlur={handleBlur}
                    onDone={editingMatchId === matchResult.match.id() ? handleDone : undefined}
                    onRef={(ref) => {
                      matchCardRefs.current[matchResult.match.id()] = ref;
                    }}
                    gameId={gameId}
                  />
                ))}
              </View>
            );
          })}
        </View>
        {/* Legend for odds indicators */}
        {shouldShowLegend && (
          <View style={styles.legendContainer}>
            <Text style={styles.legendTitle}>{t('games.oddsLegend')}</Text>
            <View style={styles.legendItem}>
              <Text style={styles.legendStar}>⭐</Text>
              <Text style={styles.legendText}>{t('games.clearFavorite')}</Text>
            </View>
            <View style={styles.legendItem}>
              <Text style={styles.legendMark}>x1</Text>
              <Text style={styles.legendText}>{t('games.clearFavoriteBonus')}</Text>
            </View>
            <View style={styles.legendItem}>
              <Text style={styles.legendMark}>×1.5</Text>
              <Text style={styles.legendText}>{t('games.drawBonus')}</Text>
            </View>
            <View style={styles.legendItem}>
              <Text style={styles.legendMark}>×2</Text>
              <Text style={styles.legendText}>{t('games.outsiderWinBonus')}</Text>
            </View>
          </View>
        )}
        
        
        {/* Add padding at the bottom to ensure last match is visible above keyboard */}
        <View style={{ height: 100 }} />
      </ScrollView>
      
      {/* Bet Synchronization Modal */}
      <BetSyncModal
        visible={showSyncModal}
        syncOpportunity={syncOpportunity}
        onSynchronize={handleSyncSynchronize}
        onNotNow={handleSyncNotNow}
        onRetryFailed={handleRetryFailed}
        loading={isSyncing}
        syncResult={syncResult}
        mode={syncModalMode}
      />
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
  },
  scrollView: {
    flex: 1,
  },
  matchdayNavigation: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#333',
    borderBottomWidth: 1,
    borderBottomColor: '#444',
  },
  navButton: {
    padding: 8,
  },
  navButtonDisabled: {
    opacity: 0.5,
  },
  matchdayInfo: {
    flex: 1,
    alignItems: 'center',
  },
  matchdayTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  matchdayDate: {
    color: '#ccc',
    fontSize: 14,
    marginTop: 4,
  },
  matchesContainer: {
    marginTop: 16,
  },
  timeGroup: {
    marginHorizontal: 16,
    marginBottom: 12,
  },
  timeHeaderContainer: {
    marginBottom: 8,
  },
  timeHeader: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  dayHeader: {
    fontSize: 14,
    color: '#ccc',
    marginTop: 2,
  },
  matchCard: {
    backgroundColor: '#f5f5f5',
    padding: 16,
    borderRadius: 8,
    marginBottom: 12,
    position: 'relative',
  },
  matchCardWithTag: {
    paddingTop: 48,
  },
  finishedMatchCard: {
    backgroundColor: '#d3d3d3',
    opacity: 0.6,
  },
  bettingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 8,
  },
  teamContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
    justifyContent: 'space-between',
    gap: 8,
  },
  awayTeamContainer: {
    flexDirection: 'row-reverse',
  },
  inputContainer: {
    position: 'absolute',
    right: 0,
    width: 40,
    marginTop: 25,
  },
  awayInputContainer: {
    position: 'absolute',
    left: 0,
    right: 'auto',
  },
  logoAndNameContainer: {
    flexDirection: 'column',
    alignItems: 'center',
    gap: 4,
    position: 'absolute',
    left: 0,
    right: 50, // Leave space for the input (40px width + 10px margin)
    justifyContent: 'center',
    height: '100%',
    marginTop: 25,
  },
  awayLogoAndNameContainer: {
    flexDirection: 'column',
    alignItems: 'center',
    gap: 4,
    position: 'absolute',
    left: 50, // Leave space for the input (40px width + 10px margin)
    right: 0,
    justifyContent: 'center',
    height: '100%',
  },
  // teamLogo style removed as we're using SVG components now
  teamName: {
    fontSize: 14,
    color: '#333',
    textAlign: 'center',
    minWidth: 60,
  },
  awayTeamName: {
    textAlign: 'center',
  },
  vsText: {
    fontSize: 14,
    color: '#666',
    marginHorizontal: 8,
    marginTop: 25,
  },
  betInput: {
    borderWidth: 1,
    borderColor: '#ccc',
    borderRadius: 4,
    padding: 8,
    width: 40,
    textAlign: 'center',
    backgroundColor: '#fff',
  },
  disabledInput: {
    backgroundColor: '#f0f0f0',
    color: '#999',
  },
  resultRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 8,
  },
  resultTeam: {
    fontSize: 14,
    color: '#333',
    flex: 1,
    textAlign: 'center',
  },
  resultScore: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
    marginHorizontal: 8,
  },
  topRightContainer: {
    position: 'absolute',
    top: 8,
    right: 8,
    flexDirection: 'row',
    alignItems: 'center',
    zIndex: 1,
  },
  topLeftContainer: {
    position: 'absolute',
    top: 8,
    left: 14,
    flexDirection: 'row',
    alignItems: 'center',
    zIndex: 1,
  },
  statusTag: {
    marginRight: 8,
  },
  statusTagText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: 'bold',
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
    backgroundColor: colors.primary,
  },
  negativeTag: {
    backgroundColor: '#ff6b6b',
  },
  toggleButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 8,
    paddingHorizontal: 16,
    backgroundColor: '#e8e8e8',
    borderRadius: 4,
    marginTop: 8,
    marginBottom: 4,
  },
  toggleButtonText: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#333',
  },
  betResultContainer: {
    marginTop: 8,
  },
  scoresContainer: {
    backgroundColor: '#f9f9f9',
    padding: 12,
    borderRadius: 4,
  },
  playerScoreRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 4,
  },
  scoreText: {
    fontSize: 14,
    color: '#333',
  },
  betResultText: {
    fontSize: 14,
    color: '#666',
    fontStyle: 'italic',
  },
  legendContainer: {
    marginTop: 24,
    marginHorizontal: 16,
    padding: 16,
    backgroundColor: '#333',
    borderRadius: 8,
  },
  legendTitle: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  legendStar: {
    fontSize: 18,
    marginRight: 8,
    color: colors.primary,
    width: 32,
    textAlign: 'center',
  },
  legendMark: {
    fontSize: 16,
    marginRight: 8,
    color: colors.primary,
    width: 32,
    textAlign: 'center',
  },
  legendText: {
    color: '#fff',
    fontSize: 14,
    textAlign: 'left',
  },
  doneButton: {
    backgroundColor: colors.primary,
    paddingVertical: 10,
    paddingHorizontal: 20,
    borderRadius: 5,
    alignSelf: 'center',
    marginTop: 10,
  },
  doneButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  oddsContainer: {
    marginTop: 40,
    paddingVertical: 8,
    paddingHorizontal: 12,
    backgroundColor: '#e0e0e0',
    borderRadius: 4,
  },
  oddsRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
  },
  oddsItem: {
    alignItems: 'center',
  },
  oddsLabel: {
    fontSize: 12,
    color: '#666',
    marginBottom: 4,
  },
  oddsValue: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
  },
  favoriteStar: {
    fontSize: 16,
    color: '#000',
  },
  outsiderMark: {
    fontSize: 14,
    color: '#000',
  },
  drawMark: {
    fontSize: 14,
    color: '#000',
  },
  nextMatchButton: {
    backgroundColor: colors.secondary,
    paddingVertical: 14,
    paddingHorizontal: 32,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginVertical: 16,
    marginHorizontal: 32,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 6,
    elevation: 4,
  },
  nextMatchButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
}); 