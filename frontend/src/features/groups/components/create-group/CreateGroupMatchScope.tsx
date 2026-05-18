import { Pressable, StyleSheet, Text, View } from 'react-native';
import { CreateGroupTeamSelector } from './CreateGroupTeamSelector';

type MatchScope = 'Todos os jogos' | 'Selecionar seleções';

type CreateGroupMatchScopeProps = {
  matchScope: MatchScope;
  matchScopes: readonly MatchScope[];
  onChangeMatchScope: (value: MatchScope) => void;
  filteredTeams: string[];
  isOpen: boolean;
  selectedTeams: string[];
  searchText: string;
  onChangeSearchText: (value: string) => void;
  onToggleOpen: () => void;
  onToggleTeam: (team: string) => void;
};

export function CreateGroupMatchScope({
  matchScope,
  matchScopes,
  onChangeMatchScope,
  filteredTeams,
  isOpen,
  selectedTeams,
  searchText,
  onChangeSearchText,
  onToggleOpen,
  onToggleTeam,
}: CreateGroupMatchScopeProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.label}>Jogos do bolão</Text>
      <View style={styles.segmentedControl}>
        {matchScopes.map((item) => {
          const isSelected = item === matchScope;

          return (
            <Pressable
              key={item}
              onPress={() => onChangeMatchScope(item)}
              style={[styles.segmentButton, isSelected && styles.segmentButtonSelected]}>
              <Text
                style={[styles.segmentButtonText, isSelected && styles.segmentButtonTextSelected]}>
                {item}
              </Text>
            </Pressable>
          );
        })}
      </View>
      {matchScope === 'Selecionar seleções' ? (
        <View style={styles.fieldGroup}>
          <Text style={styles.label}>Seleções</Text>
          <CreateGroupTeamSelector
            filteredTeams={filteredTeams}
            isOpen={isOpen}
            onChangeSearchText={onChangeSearchText}
            onToggleOpen={onToggleOpen}
            onToggleTeam={onToggleTeam}
            searchText={searchText}
            selectedTeams={selectedTeams}
          />
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    gap: 8,
  },
  fieldGroup: {
    gap: 8,
  },
  label: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '700',
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
});
