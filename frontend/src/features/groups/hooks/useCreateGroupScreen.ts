import { useMemo, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { createGroup, type CreateGroupPayload } from '../services/groups';

const matchScopes = ['Todos os jogos', 'Selecionar seleções'] as const;
const worldCupTeams = [
  'Brasil',
  'África do Sul',
  'Alemanha',
  'Arábia Saudita',
  'Argélia',
  'Argentina',
  'Austrália',
  'Áustria',
  'Bélgica',
  'Bósnia e Herzegovina',
  'Cabo Verde',
  'Canadá',
  'Colombia',
  'Coreia do Sul',
  'Costa do Marfim',
  'Croácia',
  'Curação',
  'Egito',
  'Equador',
  'Espanha',
  'Estados Unidos',
  'França',
  'Gana',
  'Haiti',
  'Holanda',
  'Inglaterra',
  'Irã',
  'Iraque',
  'Japão',
  'Jordânia',
  'Marrocos',
  'México',
  'Noruega',
  'Nova Zelândia',
  'Panamá',
  'Paraguai',
  'Portugal',
  'Qatar',
  'Rep. Democrática do Congo',
  'República Tcheca',
  'Escócia',
  'Senegal',
  'Suecia',
  'Suíça',
  'Tunísia',
  'Turquia',
  'Uruguai',
  'Uzbequistão',
];

type UseCreateGroupScreenResult = {
  blockPendingPredictions: boolean;
  createGroupLabel: string;
  description: string;
  formError: string | null;
  hasUnlimitedParticipants: boolean;
  isPaid: boolean;
  isPrivate: boolean;
  isSubmitting: boolean;
  matchScope: (typeof matchScopes)[number];
  matchScopes: readonly (typeof matchScopes)[number][];
  participantLimit: string;
  paymentAmount: string;
  selectedTeams: string[];
  teamSearch: string;
  toggleTeamDropdown: () => void;
  isTeamDropdownOpen: boolean;
  filteredTeams: string[];
  groupName: string;
  onChangeDescription: (value: string) => void;
  onChangeGroupName: (value: string) => void;
  onChangeMatchScope: (value: (typeof matchScopes)[number]) => void;
  onChangeParticipantLimit: (value: string) => void;
  onChangeTeamSearch: (value: string) => void;
  onCreateGroup: () => Promise<void>;
  toggleTeam: (team: string) => void;
  setHasUnlimitedParticipants: (value: boolean) => void;
  setBlockPendingPredictions: (value: boolean) => void;
  setIsPaid: (value: boolean) => void;
  setIsPrivate: (value: boolean) => void;
  setPaymentAmount: (value: string) => void;
};

export function useCreateGroupScreen(
  onGroupCreated: () => void,
  onBack: () => void,
): UseCreateGroupScreenResult {
  const [groupName, setGroupName] = useState('');
  const [description, setDescription] = useState('');
  const [matchScope, setMatchScope] = useState<(typeof matchScopes)[number]>(matchScopes[0]);
  const [selectedTeams, setSelectedTeams] = useState<string[]>([]);
  const [isTeamDropdownOpen, setIsTeamDropdownOpen] = useState(false);
  const [teamSearch, setTeamSearch] = useState('');
  const [participantLimit, setParticipantLimit] = useState('20');
  const [hasUnlimitedParticipants, setHasUnlimitedParticipants] = useState(false);
  const [isPrivate, setIsPrivate] = useState(true);
  const [isPaid, setIsPaid] = useState(false);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [blockPendingPredictions, setBlockPendingPredictions] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const createGroupMutation = useMutation({
    mutationFn: createGroup,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['groups'] });
    },
  });

  const filteredTeams = useMemo(
    () =>
      worldCupTeams.filter((team) => team.toLowerCase().includes(teamSearch.trim().toLowerCase())),
    [teamSearch],
  );

  async function onCreateGroup() {
    setFormError(null);

    if (!groupName.trim()) {
      setFormError('Informe o nome do grupo.');
      return;
    }

    if (matchScope === 'Selecionar seleções' && selectedTeams.length === 0) {
      setFormError('Selecione pelo menos uma seleção para o bolão.');
      return;
    }

    if (!hasUnlimitedParticipants) {
      const parsedLimit = Number(participantLimit);

      if (!Number.isInteger(parsedLimit) || parsedLimit < 2) {
        setFormError('O limite precisa ser um número maior que 1.');
        return;
      }
    }

    const parsedPaymentAmount = Number(paymentAmount.replace(',', '.'));
    if (isPaid && (!Number.isFinite(parsedPaymentAmount) || parsedPaymentAmount < 0)) {
      setFormError('Informe um valor de participação válido.');
      return;
    }

    const payload: CreateGroupPayload = {
      block_pending_predictions: blockPendingPredictions,
      description,
      has_unlimited_participants: hasUnlimitedParticipants,
      is_paid: isPaid,
      is_private: isPrivate,
      match_scope: matchScope === 'Todos os jogos' ? 'all' : 'selected',
      name: groupName,
      participant_limit: hasUnlimitedParticipants ? null : Number(participantLimit),
      payment_amount: isPaid ? parsedPaymentAmount : 0,
      selected_teams: matchScope === 'Selecionar seleções' ? selectedTeams : [],
    };

    try {
      await createGroupMutation.mutateAsync(payload);
      onGroupCreated();
    } catch (error) {
      setFormError(errorMessage(error, 'Não foi possível criar o grupo.'));
    }
  }

  function toggleTeam(team: string) {
    setSelectedTeams((currentTeams) =>
      currentTeams.includes(team)
        ? currentTeams.filter((selectedTeam) => selectedTeam !== team)
        : [...currentTeams, team],
    );
  }

  function toggleTeamDropdown() {
    setIsTeamDropdownOpen((current) => !current);
  }

  return {
    blockPendingPredictions,
    createGroupLabel: 'Criar grupo',
    description,
    formError,
    hasUnlimitedParticipants,
    isPaid,
    isPrivate,
    isSubmitting: createGroupMutation.isPending,
    matchScope,
    matchScopes,
    participantLimit,
    paymentAmount,
    selectedTeams,
    teamSearch,
    toggleTeamDropdown,
    isTeamDropdownOpen,
    filteredTeams,
    groupName,
    onChangeDescription: setDescription,
    onChangeGroupName: setGroupName,
    onChangeMatchScope: setMatchScope,
    onChangeParticipantLimit: setParticipantLimit,
    onChangeTeamSearch: setTeamSearch,
    onCreateGroup,
    toggleTeam,
    setHasUnlimitedParticipants,
    setBlockPendingPredictions,
    setIsPaid,
    setIsPrivate,
    setPaymentAmount,
  };
}

function errorMessage(error: unknown, fallback: string) {
  if (error == null) {
    return fallback;
  }

  if (typeof error === 'string') {
    return error.trim() || fallback;
  }

  if (typeof error === 'object' && 'message' in error) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === 'string' && message.trim()) {
      return message;
    }
  }

  return fallback;
}
