import { useCallback } from 'react';

import type { RealtimeEvent } from '../../realtime/types';
import { notificationMessageFromEvent } from '../../realtime/notifications';
import { useHomeData } from './useHomeData';
import { useJoinGroupForm } from './useJoinGroupForm';
import { useRealtimeEvents } from '../../realtime/useRealtimeEvents';
import { useTemporaryNotification } from '../../../shared/hooks/useTemporaryNotification';

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
