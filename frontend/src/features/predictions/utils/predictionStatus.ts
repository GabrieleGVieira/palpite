export function isScheduledStatus(status?: string): boolean {
  const normalizedStatus = status?.trim().toLowerCase();

  return normalizedStatus === 'scheduled' || normalizedStatus === 'timed';
}
