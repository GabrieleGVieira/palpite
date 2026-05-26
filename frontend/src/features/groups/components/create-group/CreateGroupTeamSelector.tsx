import { Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';

type CreateGroupTeamSelectorProps = {
  filteredTeams: string[];
  isOpen: boolean;
  selectedTeams: string[];
  searchText: string;
  onChangeSearchText: (value: string) => void;
  onToggleOpen: () => void;
  onToggleTeam: (team: string) => void;
};

export function CreateGroupTeamSelector({
  filteredTeams,
  isOpen,
  selectedTeams,
  searchText,
  onChangeSearchText,
  onToggleOpen,
  onToggleTeam,
}: CreateGroupTeamSelectorProps) {
  return (
    <View style={styles.selectBox}>
      <Pressable onPress={onToggleOpen} style={styles.dropdownButton}>
        <Text style={styles.dropdownButtonText}>
          {selectedTeams.length > 0
            ? `${selectedTeams.length} selecionada${selectedTeams.length > 1 ? 's' : ''}`
            : 'Escolha uma ou mais seleções'}
        </Text>
        <Text style={styles.dropdownIcon}>{isOpen ? '▲' : '▼'}</Text>
      </Pressable>

      {selectedTeams.length > 0 ? (
        <Text style={styles.selectedTeamsText}>{selectedTeams.join(', ')}</Text>
      ) : null}

      {isOpen ? (
        <View style={styles.dropdownList}>
          <TextInput
            autoCapitalize="words"
            onChangeText={onChangeSearchText}
            placeholder="Pesquisar seleção"
            placeholderTextColor="#7c8898"
            style={styles.searchInput}
            value={searchText}
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
                  onPress={() => onToggleTeam(team)}
                  style={styles.dropdownItem}>
                  <View style={[styles.checkbox, isSelected && styles.checkboxSelected]}>
                    {isSelected ? <Text style={styles.checkboxMark}>✓</Text> : null}
                  </View>
                  <Text style={styles.dropdownItemText}>{team}</Text>
                </Pressable>
              );
            })}

            {filteredTeams.length === 0 ? (
              <Text style={styles.emptySearchText}>Nenhuma seleção encontrada.</Text>
            ) : null}
          </ScrollView>
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
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
});
