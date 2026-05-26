package repositories

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

func HasLiveOrSoonMatches(ctx context.Context, db Querier) (bool, error) {
	// 1. Verifica se existe partida ao vivo ou partida prestes a comecar.
	var exists bool
	err := db.QueryRow(ctx, `
		select exists (
			select 1
			from world_cup_matches
			where status = 'live'
				or (
					status not in ('finished', 'cancelled', 'postponed')
					and kickoff_at between now() - interval '3 hours' and now() + interval '30 minutes'
				)
		)
	`).Scan(&exists)

	// 2. Retorna o booleano calculado pelo banco junto com qualquer erro da consulta.
	return exists, err
}

func MatchSnapshotByProviderMatch(ctx context.Context, db Querier, match domain.ProviderMatch) (domain.MatchSnapshot, error) {
	// 1. Procura a partida pela identidade do provedor; nomes sao fallback para partidas manuais.
	var snapshot domain.MatchSnapshot
	err := db.QueryRow(ctx, `
		select id, status, home_score, away_score
		from world_cup_matches
		where (
				$1::text is not null
				and external_id = $1
			)
			or (
				home_team = $2
				and away_team = $3
				and kickoff_at = $4
			)
	`, nullableString(match.ExternalID), match.HomeTeam, match.AwayTeam, match.KickoffAt).Scan(
		&snapshot.ID,
		&snapshot.Status,
		&snapshot.HomeScore,
		&snapshot.AwayScore,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// 2. Converte o erro especifico do pgx para o erro padronizado dos repositorios.
		return domain.MatchSnapshot{}, ErrNotFound
	}

	// 3. Retorna o snapshot encontrado ou propaga erro de leitura.
	return snapshot, err
}

func UpsertProviderMatch(ctx context.Context, db Querier, match domain.ProviderMatch) (int, string, error) {
	// 1. Prepara variaveis que receberao o ID da partida e se houve insert/update real.
	var matchID string
	var changed bool
	// 2. Insere a partida ou atualiza campos que mudaram no provedor.
	err := db.QueryRow(ctx, `
		with updated_by_external_id as (
			update world_cup_matches
			set
				home_team = $2,
				away_team = $3,
				stage = $4,
				status = $6,
				home_score = $7,
				away_score = $8,
				finished_at = case
					when $6 = 'finished' then coalesce(finished_at, now())
					else null
				end,
				last_synced_at = now()
			where $1::text is not null
				and external_id = $1
				and (
					home_team is distinct from $2
					or away_team is distinct from $3
					or stage is distinct from $4
					or status is distinct from $6
					or home_score is distinct from $7
					or away_score is distinct from $8
				)
			returning id, true as changed
		),
		existing_by_external_id as (
			select id, false as changed
			from world_cup_matches
			where $1::text is not null
				and external_id = $1
				and not exists (select 1 from updated_by_external_id)
		),
		upserted_by_match_key as (
			insert into world_cup_matches (
				external_id,
				home_team,
				away_team,
				stage,
				kickoff_at,
				status,
				home_score,
				away_score,
				finished_at,
				last_synced_at
			)
			select
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				case when $6 = 'finished' then now() else null end,
				now()
			where not exists (select 1 from updated_by_external_id)
				and not exists (select 1 from existing_by_external_id)
			on conflict (home_team, away_team, kickoff_at)
			do update set
				external_id = coalesce(excluded.external_id, world_cup_matches.external_id),
				stage = excluded.stage,
				status = excluded.status,
				home_score = excluded.home_score,
				away_score = excluded.away_score,
				finished_at = case
					when excluded.status = 'finished' then coalesce(world_cup_matches.finished_at, now())
					else null
				end,
				last_synced_at = now()
			where world_cup_matches.external_id is distinct from coalesce(excluded.external_id, world_cup_matches.external_id)
				or world_cup_matches.stage is distinct from excluded.stage
				or world_cup_matches.status is distinct from excluded.status
				or world_cup_matches.home_score is distinct from excluded.home_score
				or world_cup_matches.away_score is distinct from excluded.away_score
			returning id, true as changed
		)
		select id, changed
		from (
			-- 5. Se houve insert/update, retorna essa linha marcada como alterada.
			select id, changed from updated_by_external_id
			union all
			select id, changed from upserted_by_match_key
			union all
			-- 6. Se nada mudou, recupera o ID da partida existente marcada como nao alterada.
			select id, changed from existing_by_external_id
			union all
			select id, false as changed
			from world_cup_matches
			where home_team = $2 and away_team = $3 and kickoff_at = $5
		) rows
		order by changed desc
		limit 1
	`, nullableString(match.ExternalID), match.HomeTeam, match.AwayTeam, match.Stage, match.KickoffAt, match.Status, match.HomeScore, match.AwayScore).Scan(&matchID, &changed)
	if err != nil {
		return 0, "", err
	}

	// 7. Traduz o booleano changed para contador de linhas alteradas usado no resumo.
	if changed {
		return 1, matchID, nil
	}

	// 8. Retorna o ID mesmo quando nao houve alteracao, pois os gols e palpites ainda usam esse ID.
	return 0, matchID, nil
}

func InsertGoalEvent(ctx context.Context, db Querier, matchID string, goal domain.ProviderGoal) (bool, error) {
	// 1. Guarda o gol completo como JSON para auditoria e possivel uso futuro.
	payload, err := json.Marshal(goal)
	if err != nil {
		return false, err
	}

	// 2. Insere o evento de gol usando external_key como chave idempotente.
	commandTag, err := db.Exec(ctx, `
		insert into match_events (
			match_id,
			external_key,
			event_type,
			team_name,
			player_name,
			assist_name,
			minute,
			injury_time,
			home_score,
			away_score,
			payload
		)
		values ($1, $2, 'goal', $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		on conflict (external_key) do nothing
	`, matchID, goal.EventKey, goal.TeamName, goal.PlayerName, goal.AssistName, goal.Minute, goal.InjuryTime, goal.HomeScore, goal.AwayScore, string(payload))
	if err != nil {
		return false, err
	}

	// 3. Retorna true apenas quando uma nova linha foi criada.
	return commandTag.RowsAffected() > 0, nil
}

func ScoreProviderMatchPredictions(ctx context.Context, db Querier, matchID string, homeScore int, awayScore int) (int, error) {
	// 1. Recalcula a pontuacao dos palpites da partida com base no placar recebido.
	commandTag, err := db.Exec(ctx, `
		update predictions p
		set
			points = case
				-- 1. Placar exato recebe pontuação máxima.
				when p.home_score = $2 and p.away_score = $3 then 10

				-- 2. Acertou o vencedor/empate E TAMBÉM o número de gols de um dos times.
				-- (Exemplo: Jogo 2x1, Palpite 2x0. Ganha 5 pelo vencedor + bônus de gols = 7 pontos)
				-- Se no seu bolão isso acumula, o ideal é criar uma faixa intermediária, por exemplo, 7 pontos:
				when sign(p.home_score - p.away_score) = sign($2 - $3) 
					and (p.home_score = $2 or p.away_score = $3) then 7

				-- 3. Mesmo vencedor ou empate (sem acertar nenhum gol exato).
				when sign(p.home_score - p.away_score) = sign($2 - $3) then 5

				-- 4. Errou o vencedor, mas acertou a quantidade de gols de um dos times.
				-- (Exemplo: Jogo 2x1, Palpite 0x1. Errou quem venceu, mas cravou os gols do visitante).
				when p.home_score = $2 or p.away_score = $3 then 3

				-- 5. Resultado incorreto não recebe pontos.
				else 0
			end,
			scored_at = now(),
			updated_at = now()
		from world_cup_matches m
		where p.match_id = m.id
			and m.id = $1
			-- 5. Atualiza apenas linhas cuja pontuacao realmente mudaria.
			and p.points is distinct from case
				when p.home_score = $2 and p.away_score = $3 then 10
				when sign(p.home_score - p.away_score) = sign($2 - $3) then 5
				else 0
			end
	`, matchID, homeScore, awayScore)
	if err != nil {
		return 0, err
	}

	// 6. Retorna quantos palpites tiveram pontuacao alterada.
	return int(commandTag.RowsAffected()), nil
}

func AffectedGroupsByMatch(ctx context.Context, db Querier, matchID string) ([]domain.AffectedGroup, error) {
	// 1. Busca grupos que possuem ao menos um palpite para a partida.
	rows, err := db.Query(ctx, `
		select distinct g.id::text, g.name
		from groups g
		join predictions p on p.group_id = g.id
		where p.match_id = $1
		order by g.name asc
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 2. Percorre o cursor e converte cada linha em AffectedGroup.
	groups := []domain.AffectedGroup{}
	for rows.Next() {
		var group domain.AffectedGroup
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	// 3. Retorna os grupos acumulados e qualquer erro ocorrido durante a iteracao.
	return groups, rows.Err()
}

func nullableString(value string) *string {
	// 1. Converte string vazia para nil para persistir NULL no banco.
	if value == "" {
		return nil
	}

	// 2. Para valores preenchidos, retorna ponteiro para permitir uso em parametros nullable.
	return &value
}
