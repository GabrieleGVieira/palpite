import { StatusBar } from 'expo-status-bar';
import { useState } from 'react';
import {
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Switch,
  Text,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { createGroup } from '../services/groups';

type CreateGroupScreenProps = {
  onBack: () => void;
};

const matchScopes = ['Todos os jogos', 'Selecionar selecoes'];

const worldCupTeams = [
  'Brasil',
  'Africa do Sul',
  'Alemanha',
  'Argelia',
  'Argentina',
  'Australia',
  'Austria',
  'Belgica',
  'Bosnia-Herzegovina',
  'Cabo Verde',
  'Canada',
  'Colombia',
  'Coreia do Sul',
  'Costa do Marfim',
  'Croacia',
  'Curacao',
  'DR Congo',
  'Egito',
  'Equador',
  'Espanha',
  'Estados Unidos',
  'Franca',
  'Gana',
  'Haiti',
  'Holanda',
  'Inglaterra',
  'Ira',
  'Iraque',
  'Japao',
  'Jordania',
  'Marrocos',
  'Mexico',
  'Noruega',
  'Nova Zelandia',
  'Panama',
  'Paraguai',
  'Portugal',
  'Qatar',
  'Republica Tcheca',
  'Arabia Saudita',
  'Escocia',
  'Senegal',
  'Suecia',
  'Suica',
  'Tunisia',
  'Turquia',
  'Uruguai',
  'Uzbequistao',
];

export function CreateGroupScreen({ onBack }: CreateGroupScreenProps) {
  const [groupName, setGroupName] = useState('');
  const [description, setDescription] = useState('');
  const [matchScope, setMatchScope] = useState(matchScopes[0]);
  const [selectedTeams, setSelectedTeams] = useState<string[]>([]);
  const [isTeamDropdownOpen, setIsTeamDropdownOpen] = useState(false);
  const [teamSearch, setTeamSearch] = useState('');
  const [participantLimit, setParticipantLimit] = useState('20');
  const [hasUnlimitedParticipants, setHasUnlimitedParticipants] = useState(false);
  const [isPrivate, setIsPrivate] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  async function handleCreateGroup() {
    setFormError(null);
    setSuccessMessage(null);

    if (!groupName.trim()) {
      setFormError('Informe o nome do grupo.');
      return;
    }

    if (matchScope === 'Selecionar selecoes' && selectedTeams.length === 0) {
      setFormError('Selecione pelo menos uma selecao para o bolao.');
      return;
    }

    if (!hasUnlimitedParticipants) {
      const parsedLimit = Number(participantLimit);

      if (!Number.isInteger(parsedLimit) || parsedLimit < 2) {
        setFormError('O limite precisa ser um numero maior que 1.');
        return;
      }
    }

    setIsSubmitting(true);

    try {
      const group = await createGroup({
        description,
        has_unlimited_participants: hasUnlimitedParticipants,
        is_private: isPrivate,
        match_scope: matchScope === 'Todos os jogos' ? 'all' : 'selected',
        name: groupName,
        participant_limit: hasUnlimitedParticipants ? null : Number(participantLimit),
        selected_teams: matchScope === 'Selecionar selecoes' ? selectedTeams : [],
      });

      setSuccessMessage(`Grupo criado. Codigo de convite: ${group.invite_code}`);
    } catch (error) {
      setFormError(error instanceof Error ? error.message : 'Nao foi possivel criar o grupo.');
    } finally {
      setIsSubmitting(false);
    }
  }

  function toggleTeam(team: string) {
    setSelectedTeams((currentTeams) => {
      if (currentTeams.includes(team)) {
        return currentTeams.filter((selectedTeam) => selectedTeam !== team);
      }

      return [...currentTeams, team];
    });
  }

  const filteredTeams = worldCupTeams.filter((team) =>
    team.toLowerCase().includes(teamSearch.trim().toLowerCase()),
  );

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.keyboardView}>
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}>
          <View style={styles.pitchMarkTop} />
          <View style={styles.pitchCircle} />

          <Pressable accessibilityLabel="Voltar" onPress={onBack} style={styles.backButton}>
            <Text style={styles.backButtonText}>‹</Text>
          </Pressable>

          <View style={styles.header}>
            <Text style={styles.title}>Criar grupo</Text>
            <Text style={styles.subtitle}>
              Monte um bolao da Copa do Mundo, defina quais jogos entram e convide sua turma para
              disputar o ranking.
            </Text>
          </View>

          <View style={styles.form}>
            <View style={styles.fieldGroup}>
              <Text style={styles.label}>Nome do grupo</Text>
              <TextInput
                autoCapitalize="words"
                onChangeText={setGroupName}
                placeholder="Ex: Familia na Copa"
                placeholderTextColor="#7c8898"
                style={styles.input}
                value={groupName}
              />
            </View>

            <View style={styles.fieldGroup}>
              <Text style={styles.label}>Descricao</Text>
              <TextInput
                multiline
                onChangeText={setDescription}
                placeholder="Regras, combinados ou premio simbolico"
                placeholderTextColor="#7c8898"
                style={[styles.input, styles.textArea]}
                textAlignVertical="top"
                value={description}
              />
            </View>

            <View style={styles.fieldGroup}>
              <Text style={styles.label}>Jogos do bolao</Text>
              <View style={styles.segmentedControl}>
                {matchScopes.map((item) => {
                  const isSelected = item === matchScope;

                  return (
                    <Pressable
                      key={item}
                      onPress={() => setMatchScope(item)}
                      style={[styles.segmentButton, isSelected && styles.segmentButtonSelected]}>
                      <Text
                        style={[
                          styles.segmentButtonText,
                          isSelected && styles.segmentButtonTextSelected,
                        ]}>
                        {item}
                      </Text>
                    </Pressable>
                  );
                })}
              </View>
            </View>

            {matchScope === 'Selecionar selecoes' ? (
              <View style={styles.fieldGroup}>
                <Text style={styles.label}>Selecoes</Text>
                <View style={styles.selectBox}>
                  <Pressable
                    onPress={() => setIsTeamDropdownOpen((isOpen) => !isOpen)}
                    style={styles.dropdownButton}>
                    <Text style={styles.dropdownButtonText}>
                      {selectedTeams.length > 0
                        ? `${selectedTeams.length} selecionada${selectedTeams.length > 1 ? 's' : ''}`
                        : 'Escolha uma ou mais selecoes'}
                    </Text>
                    <Text style={styles.dropdownIcon}>{isTeamDropdownOpen ? '⌃' : '⌄'}</Text>
                  </Pressable>

                  {selectedTeams.length > 0 ? (
                    <Text style={styles.selectedTeamsText}>{selectedTeams.join(', ')}</Text>
                  ) : null}

                  {isTeamDropdownOpen ? (
                    <View style={styles.dropdownList}>
                      <TextInput
                        autoCapitalize="words"
                        onChangeText={setTeamSearch}
                        placeholder="Pesquisar selecao"
                        placeholderTextColor="#7c8898"
                        style={styles.searchInput}
                        value={teamSearch}
                      />

                      <ScrollView
                        keyboardShouldPersistTaps="handled"
                        nestedScrollEnabled
                        style={styles.dropdownScroll}>
                        {filteredTeams.map((team) => {
                          const isSelected = selectedTeams.includes(team);

                          return (
                            <Pressable
                              key={team}
                              onPress={() => toggleTeam(team)}
                              style={styles.dropdownItem}>
                              <View
                                style={[styles.checkbox, isSelected && styles.checkboxSelected]}>
                                {isSelected ? <Text style={styles.checkboxMark}>✓</Text> : null}
                              </View>
                              <Text style={styles.dropdownItemText}>{team}</Text>
                            </Pressable>
                          );
                        })}

                        {filteredTeams.length === 0 ? (
                          <Text style={styles.emptySearchText}>Nenhuma selecao encontrada.</Text>
                        ) : null}
                      </ScrollView>
                    </View>
                  ) : null}
                </View>
              </View>
            ) : null}

            <View style={styles.row}>
              <View style={[styles.fieldGroup, styles.limitField]}>
                <Text style={styles.label}>Participantes</Text>
                <TextInput
                  editable={!hasUnlimitedParticipants}
                  keyboardType="number-pad"
                  onChangeText={setParticipantLimit}
                  placeholder={hasUnlimitedParticipants ? 'Ilimitado' : '20'}
                  placeholderTextColor="#7c8898"
                  style={[styles.input, hasUnlimitedParticipants && styles.inputDisabled]}
                  value={hasUnlimitedParticipants ? '' : participantLimit}
                />
              </View>

              <View style={styles.privateBox}>
                <View>
                  <Text style={styles.privateTitle}>Ilimitado</Text>
                  <Text style={styles.privateSubtitle}>Sem teto de entrada</Text>
                </View>
                <Switch
                  onValueChange={setHasUnlimitedParticipants}
                  thumbColor="#ffffff"
                  trackColor={{ false: '#c9d7c3', true: '#1f7a4a' }}
                  value={hasUnlimitedParticipants}
                />
              </View>
            </View>

            <View style={styles.row}>
              <View style={styles.privateBox}>
                <View>
                  <Text style={styles.privateTitle}>Privado</Text>
                  <Text style={styles.privateSubtitle}>Entrada por convite</Text>
                </View>
                <Switch
                  onValueChange={setIsPrivate}
                  thumbColor="#ffffff"
                  trackColor={{ false: '#c9d7c3', true: '#1f7a4a' }}
                  value={isPrivate}
                />
              </View>
            </View>

            <View style={styles.rulesBox}>
              <Text style={styles.rulesTitle}>Regras iniciais</Text>
              <Text style={styles.rulesText}>3 pontos por placar exato</Text>
              <Text style={styles.rulesText}>1 ponto por vencedor ou empate correto</Text>
              <Text style={styles.rulesText}>Palpites fecham no inicio de cada jogo</Text>
            </View>

            {formError ? <Text style={styles.errorText}>{formError}</Text> : null}
            {successMessage ? <Text style={styles.successText}>{successMessage}</Text> : null}

            <Pressable
              disabled={isSubmitting}
              onPress={handleCreateGroup}
              style={[styles.primaryButton, isSubmitting && styles.primaryButtonDisabled]}>
              <Text style={styles.primaryButtonText}>
                {isSubmitting ? 'Criando grupo...' : 'Criar grupo'}
              </Text>
            </Pressable>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f8ef',
  },
  keyboardView: {
    flex: 1,
  },
  scrollContent: {
    backgroundColor: '#f5f8ef',
    flexGrow: 1,
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  pitchMarkTop: {
    borderColor: 'rgba(255, 255, 255, 0.68)',
    borderRadius: 8,
    borderWidth: 2,
    height: 116,
    left: 24,
    position: 'absolute',
    right: 24,
    top: -42,
  },
  pitchCircle: {
    borderColor: 'rgba(32, 111, 67, 0.12)',
    borderRadius: 140,
    borderWidth: 2,
    height: 280,
    position: 'absolute',
    right: -128,
    top: 104,
    width: 280,
  },
  backButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 22,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 6, width: 0 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    width: 44,
  },
  backButtonText: {
    color: '#1f7a4a',
    fontSize: 34,
    fontWeight: '600',
    lineHeight: 38,
  },
  header: {
    paddingTop: 32,
  },
  title: {
    color: '#123d2a',
    fontSize: 36,
    fontWeight: '800',
    letterSpacing: 0,
  },
  subtitle: {
    color: '#486654',
    fontSize: 16,
    lineHeight: 24,
    marginTop: 12,
    maxWidth: 360,
  },
  form: {
    gap: 16,
    marginTop: 36,
  },
  fieldGroup: {
    gap: 8,
  },
  label: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '700',
  },
  input: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 16,
    minHeight: 54,
    paddingHorizontal: 16,
  },
  inputDisabled: {
    backgroundColor: '#edf3e8',
    color: '#7c8898',
  },
  textArea: {
    minHeight: 104,
    paddingTop: 14,
  },
  segmentedControl: {
    backgroundColor: '#e7efdf',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    padding: 4,
  },
  segmentButton: {
    alignItems: 'center',
    borderRadius: 6,
    flex: 1,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: 6,
  },
  segmentButtonSelected: {
    backgroundColor: '#ffffff',
  },
  segmentButtonText: {
    color: '#486654',
    fontSize: 12,
    fontWeight: '800',
    textAlign: 'center',
  },
  segmentButtonTextSelected: {
    color: '#1f7a4a',
  },
  selectBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    overflow: 'hidden',
  },
  dropdownButton: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 54,
    paddingHorizontal: 16,
  },
  dropdownButtonText: {
    color: '#183f2d',
    flex: 1,
    fontSize: 15,
    fontWeight: '700',
  },
  dropdownIcon: {
    color: '#1f7a4a',
    fontSize: 22,
    fontWeight: '800',
    marginLeft: 12,
  },
  selectedTeamsText: {
    borderTopColor: '#edf3e8',
    borderTopWidth: 1,
    color: '#486654',
    fontSize: 13,
    lineHeight: 18,
    paddingHorizontal: 16,
    paddingVertical: 10,
  },
  dropdownList: {
    borderTopColor: '#cfe0c9',
    borderTopWidth: 1,
  },
  searchInput: {
    backgroundColor: '#f5f8ef',
    borderColor: '#d9e7d4',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 15,
    margin: 12,
    minHeight: 46,
    paddingHorizontal: 12,
  },
  dropdownScroll: {
    maxHeight: 280,
  },
  dropdownItem: {
    alignItems: 'center',
    flexDirection: 'row',
    minHeight: 46,
    paddingHorizontal: 16,
  },
  checkbox: {
    alignItems: 'center',
    borderColor: '#9bb99c',
    borderRadius: 4,
    borderWidth: 1,
    height: 20,
    justifyContent: 'center',
    marginRight: 10,
    width: 20,
  },
  checkboxSelected: {
    backgroundColor: '#1f7a4a',
    borderColor: '#1f7a4a',
  },
  checkboxMark: {
    color: '#ffffff',
    fontSize: 13,
    fontWeight: '800',
  },
  dropdownItemText: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '700',
  },
  emptySearchText: {
    color: '#486654',
    fontSize: 14,
    paddingHorizontal: 16,
    paddingVertical: 18,
  },
  row: {
    flexDirection: 'row',
    gap: 12,
  },
  limitField: {
    flex: 1,
  },
  privateBox: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    flex: 1.2,
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 82,
    paddingHorizontal: 14,
  },
  privateTitle: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '800',
  },
  privateSubtitle: {
    color: '#486654',
    fontSize: 12,
    marginTop: 4,
  },
  rulesBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 16,
  },
  rulesTitle: {
    color: '#123d2a',
    fontSize: 16,
    fontWeight: '800',
    marginBottom: 8,
  },
  rulesText: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 22,
  },
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
  successText: {
    color: '#1f7a4a',
    fontSize: 13,
    lineHeight: 18,
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 54,
  },
  primaryButtonDisabled: {
    opacity: 0.72,
  },
  primaryButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '800',
  },
});
