import type { GroupMatch } from '../services/groups';

export function formatDate(value: string) {
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(new Date(value));
}

export function formatUserID(userID: string) {
  if (userID.length <= 12) {
    return userID;
  }

  return `${userID.slice(0, 8)}...${userID.slice(-4)}`;
}

export function formatMatchStatus(status: GroupMatch['status']) {
  const statusLabels: Record<string, string> = {
    cancelled: 'Cancelado',
    finished: 'Encerrado',
    in_play: 'Ao vivo',
    live: 'Ao vivo',
    paused: 'Pausado',
    postponed: 'Adiado',
    scheduled: 'Agendado',
    suspended: 'Suspenso',
    timed: 'Agendado',
  };

  return statusLabels[status.trim().toLowerCase()] ?? status;
}

export function formatMatchStage(stage: GroupMatch['stage']) {
  const stageLabels: Record<string, string> = {
    FINAL: 'Final',
    GROUP_STAGE: 'Fase de grupos',
    LAST_16: 'Oitavas de final',
    LAST_32: 'Mata-mata inicial',
    QUARTER_FINALS: 'Quartas de final',
    SEMI_FINALS: 'Semi-finais',
  };

  return stageLabels[stage] ?? stage;
}
