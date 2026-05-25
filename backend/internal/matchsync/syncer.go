package matchsync

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

type Syncer struct {
	// baseURL e o endpoint base da API football-data.
	baseURL string
	// competitionCode identifica a competicao que sera sincronizada.
	competitionCode string
	// db executa consultas e comandos no banco da aplicacao.
	db datastore
	// httpClient faz as requisicoes ao provedor externo.
	httpClient *http.Client
	// inFlight impede duas sincronizacoes simultaneas no mesmo processo.
	inFlight sync.Mutex
	// lastRequestAt guarda a hora da ultima chamada para aplicar rate limit.
	lastRequestAt time.Time
	// logger registra falhas, eventos e resumos de sincronizacao.
	logger *slog.Logger
	// publisher envia eventos realtime gerados pela sincronizacao.
	publisher Publisher
	// rateMu protege lastRequestAt contra acesso concorrente.
	rateMu sync.Mutex
	// season restringe a consulta a uma temporada especifica quando configurada.
	season string
	// token autentica as chamadas na API football-data.
	token string
}

func New(cfg config.Config, db datastore, logger *slog.Logger) (*Syncer, bool) {
	// 1. Sem token do football-data nao ha como sincronizar, entao o worker fica desabilitado.
	if strings.TrimSpace(cfg.FootballDataToken) == "" {
		return nil, false
	}

	// 2. Garante um logger valido mesmo quando o chamador passa nil.
	if logger == nil {
		logger = slog.Default()
	}

	// 3. Normaliza a URL base e injeta dependencias padrao do sincronizador.
	return &Syncer{
		baseURL:         strings.TrimRight(cfg.FootballDataAPIBaseURL, "/"),
		competitionCode: cfg.FootballDataCompetitionCode,
		db:              db,
		httpClient:      http.DefaultClient,
		logger:          logger,
		publisher:       LogPublisher{logger: logger},
		season:          cfg.FootballDataSeason,
		token:           cfg.FootballDataToken,
	}, true
}

func (syncer *Syncer) SetPublisher(publisher Publisher) {
	// 1. Permite substituir o publisher padrao, normalmente em testes ou integracao realtime.
	if publisher != nil {
		// 2. Ignora nil para preservar o publisher atual e evitar falhas em publicacoes futuras.
		syncer.publisher = publisher
	}
}

func (syncer *Syncer) SyncOnce(ctx context.Context, kind syncKind) (domain.SyncSummary, error) {
	// 1. Tenta adquirir o lock sem bloquear; se ja existe uma sincronizacao, esta chamada sai sem erro.
	if !syncer.inFlight.TryLock() {
		return domain.SyncSummary{}, nil
	}
	defer syncer.inFlight.Unlock()

	// 2. Decide se vale a pena sincronizar esse tipo agora, evitando chamadas externas desnecessarias.
	shouldSync, err := syncer.shouldSync(ctx, kind)
	if err != nil {
		return domain.SyncSummary{}, err
	}
	if !shouldSync {
		return domain.SyncSummary{}, nil
	}

	// 3. Busca as partidas no provedor externo usando o recorte solicitado.
	matches, err := syncer.fetchMatches(ctx, kind)
	if err != nil {
		return domain.SyncSummary{}, err
	}

	// 4. Inicializa o resumo com a quantidade recebida do provedor.
	summary := domain.SyncSummary{SyncedMatches: len(matches)}
	for _, match := range matches {
		// 5. Normaliza os dados externos para o formato esperado pelo dominio.
		match = domain.NormalizeProviderMatch(match)
		if err := domain.ValidateProviderMatch(match); err != nil {
			// 6. Ignora partidas invalidas sem abortar o lote inteiro.
			syncer.logger.Warn("provider match ignored", "error", err)
			continue
		}

		// 7. Sincroniza a partida individualmente e acumula os contadores retornados.
		matchSummary, err := syncer.syncMatch(ctx, match)
		if err != nil {
			return domain.SyncSummary{}, err
		}

		summary.ChangedMatches += matchSummary.ChangedMatches
		summary.CreatedEvents += matchSummary.CreatedEvents
		summary.ScoredPredictions += matchSummary.ScoredPredictions
		summary.UpdatedLiveMatches += matchSummary.UpdatedLiveMatches
	}

	return summary, nil
}

func (syncer *Syncer) shouldSync(ctx context.Context, kind syncKind) (bool, error) {
	// 1. Sincronizacoes de hoje e futuras sempre podem rodar quando o ticker dispara.
	if kind != syncLive {
		return true, nil
	}

	// 2. Para live, tambem consulta sempre: o provedor e a fonte mais confiavel para descobrir que um jogo mudou para ao vivo.
	return true, nil
}

func (syncer *Syncer) syncMatch(ctx context.Context, match domain.ProviderMatch) (domain.SyncSummary, error) {
	// 1. Busca o snapshot anterior para comparar status e placar antes do upsert.
	snapshot, err := repositories.MatchSnapshotByProviderMatch(ctx, syncer.db, match)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return domain.SyncSummary{}, err
	}

	// 2. Insere ou atualiza a partida com os dados atuais do provedor.
	changedRows, matchID, err := repositories.UpsertProviderMatch(ctx, syncer.db, match)
	if err != nil {
		return domain.SyncSummary{}, err
	}

	// 3. Registra se a partida mudou e publica evento quando houve alteracao persistida.
	summary := domain.SyncSummary{ChangedMatches: changedRows}
	if changedRows > 0 {
		syncer.publishMatchChanged(ctx, matchID, snapshot, match)
	}

	// 4. Insere eventos de gol ainda nao conhecidos e publica cada gol novo.
	createdEvents, err := syncer.syncGoals(ctx, matchID, match)
	if err != nil {
		return domain.SyncSummary{}, err
	}
	summary.CreatedEvents = createdEvents

	// 5. Sem placar completo ou sem status relevante, nao ha pontuacao de palpites para calcular.
	if match.HomeScore == nil || match.AwayScore == nil || (match.Status != "live" && match.Status != "finished") {
		return summary, nil
	}

	// 6. Atualiza a pontuacao dos palpites dessa partida conforme o placar atual/final.
	scoredPredictions, err := repositories.ScoreProviderMatchPredictions(ctx, syncer.db, matchID, *match.HomeScore, *match.AwayScore)
	if err != nil {
		return domain.SyncSummary{}, err
	}

	// 7. Se a pontuacao final mudou por causa de uma partida encerrada, notifica rankings afetados.
	if match.Status == "finished" && scoredPredictions > 0 && changedRows > 0 {
		if err := syncer.publishRankingChanged(ctx, matchID, match); err != nil {
			return domain.SyncSummary{}, err
		}
	}

	// 8. Completa o resumo com palpites recalculados e marca partida live atualizada quando aplicavel.
	summary.ScoredPredictions = scoredPredictions
	if match.Status == "live" {
		summary.UpdatedLiveMatches = 1
	}

	return summary, nil
}

func (syncer *Syncer) syncGoals(ctx context.Context, matchID string, match domain.ProviderMatch) (int, error) {
	// 1. Conta apenas eventos de gol criados nesta execucao.
	created := 0
	for _, goal := range match.Goals {
		// 2. Tenta inserir o gol usando chave externa para evitar duplicidade.
		wasCreated, err := repositories.InsertGoalEvent(ctx, syncer.db, matchID, goal)
		if err != nil {
			return 0, err
		}

		// 3. Quando o gol ja existia, segue para o proximo sem publicar evento duplicado.
		if !wasCreated {
			continue
		}

		// 4. Para gol novo, incrementa o contador e publica o evento na sala da partida.
		created++
		syncer.publishRealtimeEvent(ctx, domain.Event{
			Name: "match.goal",
			Payload: map[string]any{
				"away_score":  goal.AwayScore,
				"away_team":   match.AwayTeam,
				"home_score":  goal.HomeScore,
				"home_team":   match.HomeTeam,
				"match_id":    matchID,
				"minute":      goal.Minute,
				"player_name": goal.PlayerName,
				"team_name":   goal.TeamName,
				"type":        goal.Type,
			},
			Room: "match:" + matchID,
		})
	}

	return created, nil
}
