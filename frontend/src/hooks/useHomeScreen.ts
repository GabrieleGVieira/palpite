import { useCallback } from 'react';

import type { RealtimeEvent } from '../services/realtime';
import { notificationMessageFromEvent } from '../utils/realtimeNotifications';
import { useHomeData } from './useHomeData';
import { useJoinGroupForm } from './useJoinGroupForm';
import { useRealtimeEvents } from './useRealtimeEvents';
import { useTemporaryNotification } from './useTemporaryNotification';

export function useHomeScreen() {
  const homeData = useHomeData();
  const { refreshHome } = homeData;
  const joinGroupForm = useJoinGroupForm(refreshHome);
  const { notificationMessage, showNotification } = useTemporaryNotification();

  const handleRealtimeEvent = useCallback(
    (event: RealtimeEvent) => {
      if (event.name === 'ranking.updated' || event.name === 'match.finished') {
        showNotification(notificationMessageFromEvent(event));
        void refreshHome();
      }
    },
    [refreshHome, showNotification],
  );

  useRealtimeEvents({ onEvent: handleRealtimeEvent });

  return {
    ...homeData,
    ...joinGroupForm,
    notificationMessage,
  };
}
