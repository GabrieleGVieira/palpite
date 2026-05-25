package matchsync

import (
	"context"
	"fmt"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

func (syncer *Syncer) publishMatchChanged(ctx context.Context, matchID string, previous domain.MatchSnapshot, match domain.ProviderMatch) {
	// 1. Monta um payload unico com dados atuais, dados anteriores e mensagem amigavel.
	payload := map[string]any{
		"away_score":       match.AwayScore,
		"away_team":        match.AwayTeam,
		"external_id":      match.ExternalID,
		"final_away_score": match.AwayScore,
		"final_home_score": match.HomeScore,
		"home_score":       match.HomeScore,
		"home_team":        match.HomeTeam,
		"kickoff_at":       match.KickoffAt,
		"match_id":         matchID,
		"message":          resultMessage(match.HomeTeam, match.AwayTeam, match.HomeScore, match.AwayScore),
		"previous_score":   scorePair(previous.HomeScore, previous.AwayScore),
		"previous_status":  previous.Status,
		"status":           match.Status,
	}

	// 2. Publica a atualizacao geral para clientes inscritos na sala de partidas.
	syncer.publishRealtimeEvent(ctx, domain.Event{
		Name:    "match.updated",
		Payload: payload,
		Room:    "matches",
	})

	// 3. Quando a partida termina, publica tambem um evento semantico de encerramento.
	if match.Status == "finished" {
		syncer.publishRealtimeEvent(ctx, domain.Event{
			Name:    "match.finished",
			Payload: payload,
			Room:    "matches",
		})
	}
}

func (syncer *Syncer) publishRankingChanged(ctx context.Context, matchID string, match domain.ProviderMatch) error {
	// 1. Busca quais grupos tem palpites para a partida, pois so eles precisam de atualizacao.
	groups, err := repositories.AffectedGroupsByMatch(ctx, syncer.db, matchID)
	if err != nil {
		return err
	}

	// 2. Para cada grupo afetado, monta um payload com placar, partida e identificacao do grupo.
	for _, group := range groups {
		payload := map[string]any{
			"away_score": match.AwayScore,
			"away_team":  match.AwayTeam,
			"group_id":   group.ID,
			"group_name": group.Name,
			"home_score": match.HomeScore,
			"home_team":  match.HomeTeam,
			"match_id":   matchID,
			"message":    "Ranking do grupo " + group.Name + " atualizado",
		}

		// 3. Publica em uma sala geral de rankings para telas agregadas.
		syncer.publishRealtimeEvent(ctx, domain.Event{
			Name:    "ranking.updated",
			Payload: payload,
			Room:    "rankings",
		})
		// 4. Publica tambem na sala especifica do grupo para atualizar detalhes em tempo real.
		syncer.publishRealtimeEvent(ctx, domain.Event{
			Name:    "ranking.updated",
			Payload: payload,
			Room:    "group:" + group.ID,
		})
	}

	return nil
}

func (syncer *Syncer) publishRealtimeEvent(ctx context.Context, event domain.Event) {
	syncer.logger.Info(
		"realtime event sent to frontend",
		"name", event.Name,
		"room", event.Room,
		"match_id", event.Payload["match_id"],
		"group_id", event.Payload["group_id"],
		"status", event.Payload["status"],
	)
	syncer.publisher.Publish(ctx, event)
}

func scorePair(homeScore *int, awayScore *int) map[string]*int {
	// 1. Agrupa ponteiros de placar mantendo nil quando o provedor ainda nao informou o valor.
	return map[string]*int{
		"away": awayScore,
		"home": homeScore,
	}
}

func resultMessage(homeTeam string, awayTeam string, homeScore *int, awayScore *int) string {
	// 1. Se algum placar esta ausente, retorna mensagem generica sem tentar desreferenciar ponteiros nil.
	if homeScore == nil || awayScore == nil {
		return homeTeam + " x " + awayTeam + " - resultado final lancado"
	}

	// 2. Quando ambos os placares existem, inclui o resultado final formatado na mensagem.
	return fmt.Sprintf("%s %dx%d %s - resultado final lancado", homeTeam, *homeScore, *awayScore, awayTeam)
}
