export type RealtimeEventName =
  | 'match.updated'
  | 'match.finished'
  | 'match.goal'
  | 'ranking.updated'
  | string;

export type RealtimeEvent = {
  name: RealtimeEventName;
  payload: Record<string, unknown>;
  room?: string;
};

export type RealtimeOptions = {
  groupID?: string;
  onEvent: (event: RealtimeEvent) => void;
};
