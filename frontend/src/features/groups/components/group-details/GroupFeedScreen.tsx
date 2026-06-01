import { Image, Pressable, StyleSheet, Text, View } from 'react-native';

import { EmptyBox } from '../../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../../shared/components/LoadingIndicator';
import { colors } from '../../../../shared/theme';
import type { FeedReactionType, GroupFeedEvent } from '../../types';

const reactionButtons: { emoji: string; type: FeedReactionType }[] = [
  { emoji: '👏', type: 'clap' },
  { emoji: '🔥', type: 'fire' },
  { emoji: '😂', type: 'laugh' },
  { emoji: '😮', type: 'surprised' },
  { emoji: '🎯', type: 'target' },
];

type Props = {
  error: string | null;
  events: GroupFeedEvent[];
  hasMore: boolean;
  isLoading: boolean;
  isLoadingMore: boolean;
  onLoadMore: () => void;
  onToggleReaction: (eventID: string, reactionType: FeedReactionType) => void;
};

export function GroupFeedScreen({
  error,
  events,
  hasMore,
  isLoading,
  isLoadingMore,
  onLoadMore,
  onToggleReaction,
}: Props) {
  if (isLoading) {
    return <LoadingIndicator text="Carregando atividades..." />;
  }

  if (error) {
    return <Text style={styles.errorText}>{error}</Text>;
  }

  if (events.length === 0) {
    return <EmptyBox title="Sem atividades" text="Os acontecimentos do bolão aparecerão aqui." />;
  }

  return (
    <View style={styles.feedList}>
      {events.map((event) => (
        <View key={event.id} style={styles.card}>
          <View style={styles.header}>
            <Avatar event={event} />
            <View style={styles.content}>
              <Text style={styles.description}>{eventDescription(event)}</Text>
              <Text style={styles.timeText}>{relativeTime(event.createdAt)}</Text>
            </View>
          </View>

          <View style={styles.reactions}>
            {reactionButtons.map((reaction) => {
              const summary = event.reactions.find((item) => item.reactionType === reaction.type);
              return (
                <Pressable
                  key={reaction.type}
                  onPress={() => onToggleReaction(event.id, reaction.type)}
                  style={[
                    styles.reactionButton,
                    summary?.reactedByMe && styles.reactionButtonActive,
                  ]}>
                  <Text style={styles.reactionEmoji}>{reaction.emoji}</Text>
                  <Text
                    style={[
                      styles.reactionCount,
                      summary?.reactedByMe && styles.reactionCountActive,
                    ]}>
                    {summary?.count ?? 0}
                  </Text>
                </Pressable>
              );
            })}
          </View>
        </View>
      ))}

      {hasMore ? (
        <Pressable disabled={isLoadingMore} onPress={onLoadMore} style={styles.loadMoreButton}>
          <Text style={styles.loadMoreText}>
            {isLoadingMore ? 'Carregando...' : 'Carregar mais'}
          </Text>
        </Pressable>
      ) : null}
    </View>
  );
}

function Avatar({ event }: { event: GroupFeedEvent }) {
  const avatarURL = event.actor?.avatarUrl || event.actor?.avatar_url;
  if (avatarURL) {
    return <Image source={{ uri: avatarURL }} style={styles.avatar} />;
  }

  return (
    <View style={styles.avatarFallback}>
      <Text style={styles.avatarText}>{eventIcon(event.type)}</Text>
    </View>
  );
}

function eventDescription(event: GroupFeedEvent) {
  const actorName = event.actor?.name || 'Palpiteiro';
  const metadata = event.metadata ?? {};
  const homeTeam = asString(metadata.homeTeam);
  const awayTeam = asString(metadata.awayTeam);
  const score = asString(metadata.score);

  switch (event.type) {
    case 'member_joined':
      return `👤 ${actorName} entrou no grupo.`;
    case 'leader_changed':
      return `🏆 ${actorName} assumiu a liderança do bolão.`;
    case 'exact_score':
      return `🎯 ${actorName} acertou o placar exato de ${homeTeam} ${formatScore(score)} ${awayTeam}.`;
    case 'match_finished':
      return `⚽ ${homeTeam} x ${awayTeam} foi finalizado.`;
    case 'top3_reached':
      return `📈 ${actorName} entrou no Top 3.`;
    default:
      return 'Nova atividade no grupo.';
  }
}

function eventIcon(type: GroupFeedEvent['type']) {
  switch (type) {
    case 'member_joined':
      return '👤';
    case 'leader_changed':
      return '🏆';
    case 'exact_score':
      return '🎯';
    case 'match_finished':
      return '⚽';
    case 'top3_reached':
      return '📈';
    default:
      return '•';
  }
}

function asString(value: unknown) {
  return typeof value === 'string' ? value : '';
}

function formatScore(score: string) {
  return score.replace('x', ' x ');
}

function relativeTime(value: string) {
  const date = new Date(value);
  const diffMs = Date.now() - date.getTime();
  const minutes = Math.floor(diffMs / 60000);
  if (minutes < 1) {
    return 'agora';
  }
  if (minutes < 60) {
    return `há ${minutes} minuto${minutes === 1 ? '' : 's'}`;
  }
  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `há ${hours} hora${hours === 1 ? '' : 's'}`;
  }
  if (hours < 48) {
    return 'ontem';
  }

  return new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short' }).format(date);
}

const styles = StyleSheet.create({
  feedList: {
    gap: 12,
  },
  errorText: {
    color: colors.danger,
    fontSize: 13,
    lineHeight: 18,
  },
  card: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: 14,
    padding: 14,
  },
  header: {
    flexDirection: 'row',
    gap: 12,
  },
  avatar: {
    borderRadius: 22,
    height: 44,
    width: 44,
  },
  avatarFallback: {
    alignItems: 'center',
    backgroundColor: '#e3efe0',
    borderRadius: 22,
    height: 44,
    justifyContent: 'center',
    width: 44,
  },
  avatarText: {
    fontSize: 20,
  },
  content: {
    flex: 1,
    gap: 4,
  },
  description: {
    color: colors.primaryText,
    fontSize: 15,
    fontWeight: '700',
    lineHeight: 21,
  },
  timeText: {
    color: colors.mutedText,
    fontSize: 12,
  },
  reactions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  reactionButton: {
    alignItems: 'center',
    backgroundColor: '#f2f6ef',
    borderColor: colors.border,
    borderRadius: 999,
    borderWidth: 1,
    flexDirection: 'row',
    gap: 4,
    minHeight: 34,
    paddingHorizontal: 10,
  },
  reactionButtonActive: {
    backgroundColor: '#dff0e3',
    borderColor: colors.primary,
  },
  reactionEmoji: {
    fontSize: 16,
  },
  reactionCount: {
    color: colors.mutedText,
    fontSize: 12,
    fontWeight: '800',
  },
  reactionCountActive: {
    color: colors.primary,
  },
  loadMoreButton: {
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 8,
    minHeight: 46,
    justifyContent: 'center',
  },
  loadMoreText: {
    color: colors.surface,
    fontSize: 14,
    fontWeight: '800',
  },
});
