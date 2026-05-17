import type { RealtimeEvent } from '../services/realtime';

export function notificationMessageFromEvent(event: RealtimeEvent, fallbackGroupName?: string) {
  if (event.name === 'ranking.updated') {
    const message = stringPayload(event, 'message');
    if (message) {
      return message;
    }

    const groupName = stringPayload(event, 'group_name') || fallbackGroupName;
    return groupName ? `Ranking do grupo ${groupName} atualizado` : 'Ranking atualizado';
  }

  if (event.name === 'match.finished') {
    const message = stringPayload(event, 'message');
    if (message) {
      return message;
    }

    const homeTeam = stringPayload(event, 'home_team');
    const awayTeam = stringPayload(event, 'away_team');
    const homeScore = numberPayload(event, 'home_score');
    const awayScore = numberPayload(event, 'away_score');

    if (homeTeam && awayTeam && homeScore !== null && awayScore !== null) {
      return `${homeTeam} ${homeScore}x${awayScore} ${awayTeam} - resultado final lançado`;
    }

    return 'Resultado final lançado';
  }

  return null;
}

function stringPayload(event: RealtimeEvent, key: string) {
  const value = event.payload[key];
  return typeof value === 'string' ? value : null;
}

function numberPayload(event: RealtimeEvent, key: string) {
  const value = event.payload[key];
  return typeof value === 'number' ? value : null;
}
