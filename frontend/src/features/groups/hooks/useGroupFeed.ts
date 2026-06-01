import { useCallback, useEffect, useState } from 'react';

import {
  deleteFeedReaction,
  listGroupFeed,
  reactToFeedEvent,
  type FeedReactionType,
  type GroupFeedEvent,
} from '../services/groups';

const pageSize = 20;

export function useGroupFeed(groupID: string) {
  const [events, setEvents] = useState<GroupFeedEvent[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [page, setPage] = useState(1);

  const loadFeed = useCallback(
    async (nextPage = 1, mode: 'initial' | 'refresh' | 'more' = 'initial') => {
      if (mode === 'more') {
        setIsLoadingMore(true);
      } else if (mode === 'refresh') {
        setIsRefreshing(true);
      } else {
        setIsLoading(true);
      }
      setError(null);

      try {
        const response = await listGroupFeed(groupID, nextPage, pageSize);
        setEvents((current) =>
          nextPage === 1 ? response.events : mergeEvents(current, response.events),
        );
        setHasMore(response.hasMore);
        setPage(response.page);
      } catch (loadError) {
        setError(
          loadError instanceof Error
            ? loadError.message
            : 'Não foi possível carregar as atividades.',
        );
      } finally {
        setIsLoading(false);
        setIsLoadingMore(false);
        setIsRefreshing(false);
      }
    },
    [groupID],
  );

  useEffect(() => {
    void loadFeed();
  }, [loadFeed]);

  async function loadMore() {
    if (!hasMore || isLoadingMore) {
      return;
    }
    await loadFeed(page + 1, 'more');
  }

  async function refresh() {
    await loadFeed(1, 'refresh');
  }

  async function toggleReaction(eventID: string, reactionType: FeedReactionType) {
    const previous = events;
    const event = events.find((item) => item.id === eventID);
    const shouldRemove =
      event?.reactions.some(
        (reaction) => reaction.reactionType === reactionType && reaction.reactedByMe,
      ) ?? false;
    setEvents((current) => applyOptimisticReaction(current, eventID, reactionType, shouldRemove));

    try {
      if (shouldRemove) {
        await deleteFeedReaction(groupID, eventID, reactionType);
      } else {
        await reactToFeedEvent(groupID, eventID, reactionType);
      }
    } catch (reactionError) {
      setEvents(previous);
      setError(
        reactionError instanceof Error
          ? reactionError.message
          : 'Não foi possível atualizar a reação.',
      );
    }
  }

  return {
    error,
    events,
    hasMore,
    isLoading,
    isLoadingMore,
    isRefreshing,
    loadMore,
    refresh,
    toggleReaction,
  };
}

function mergeEvents(current: GroupFeedEvent[], next: GroupFeedEvent[]) {
  const existing = new Set(current.map((event) => event.id));
  return [...current, ...next.filter((event) => !existing.has(event.id))];
}

function applyOptimisticReaction(
  events: GroupFeedEvent[],
  eventID: string,
  reactionType: FeedReactionType,
  shouldRemove: boolean,
) {
  return events.map((event) => {
    if (event.id !== eventID) {
      return event;
    }

    const reactions = event.reactions.map((reaction) => {
      let count = reaction.count;
      let reactedByMe = reaction.reactedByMe;

      if (reaction.reactionType === reactionType) {
        if (shouldRemove) {
          count = Math.max(0, count - 1);
          reactedByMe = false;
        } else if (!reaction.reactedByMe) {
          count += 1;
          reactedByMe = true;
        }
      }

      return { ...reaction, count, reactedByMe };
    });

    if (!shouldRemove && !reactions.some((reaction) => reaction.reactionType === reactionType)) {
      reactions.push({ count: 1, reactedByMe: true, reactionType });
    }

    return {
      ...event,
      reactions: reactions.filter((reaction) => reaction.count > 0 || reaction.reactedByMe),
    };
  });
}
