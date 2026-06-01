# Metrics Engine

## Métricas calculadas

- `avg_goals_scored`: gols marcados dividido por partidas jogadas até a data de referência.
- `avg_goals_conceded`: gols sofridos dividido por partidas jogadas até a data de referência.
- `win_rate`, `draw_rate`, `loss_rate`: proporção de vitórias, empates e derrotas.
- `recent_form_score`: forma recente normalizada de 0 a 100.
- `attack_score`: força ofensiva normalizada de 0 a 100.
- `defense_score`: força defensiva normalizada de 0 a 100.
- `world_cup_history_score`: desempenho histórico em jogos de Copa do Mundo, com maior peso para anos recentes.
- `knockout_score` e `group_stage_score`: calculados somente quando houver campo `stage` confiável; o `knockout_score` também considera aproveitamento em disputas por pênaltis.
- `elo_score`: rating Elo cronológico.

## Elo

Elo inicial é 1500. A expectativa usa:

```text
E_a = 1 / (1 + 10 ^ ((R_b - R_a) / 400))
```

Atualização:

```text
R'_a = R_a + K * G * (S_a - E_a)
```

Onde `S_a` é 1 para vitória, 0.5 para empate e 0 para derrota. `K` é 20 por padrão, 30 para torneios continentais e 40 para Copa do Mundo. `G` é um multiplicador moderado por diferença de gols.

## Recent Form

Usa os últimos `N` jogos antes ou na data de referência. Vitória vale 3, empate 1 e derrota 0. Jogos mais recentes recebem peso maior:

```text
weight_i = i / total_matches
score = weighted_points / max_weighted_points * 100
```

## Attack Score

Combina média histórica de gols marcados, média de gols nos jogos recentes e desempenho em Copa:

```text
attack = 0.30 * historical_goals_score
       + 0.30 * recent_goals_score
       + 0.15 * world_cup_history_score
       + 0.15 * scorer_depth_score
       + 0.10 * penalty_independence_score
```

Médias de gols são normalizadas de 0 a 3 gols por jogo para a escala 0 a 100.

## Defense Score

Combina baixa média de gols sofridos, baixa média recente de gols sofridos e clean sheets:

```text
defense = 0.45 * conceded_score
        + 0.35 * recent_conceded_score
        + 0.20 * clean_sheet_score
```

`conceded_score` é invertido: sofrer 0 gol tende a 100, sofrer 3 ou mais tende a 0.

## Restrições temporais

Todas as métricas são calculadas até uma data de referência. Para features de partidas, o script usa apenas `team_metrics.metric_date < match_date` e rankings FIFA de `ml-service/data/ranking_fifa_historical.csv` com `date < match_date`.

## Limitações conhecidas

- O CSV histórico não traz uma fase estruturada para todos os jogos; por isso `knockout_score` e `group_stage_score` podem ficar nulos.
- Rankings FIFA dependem da cobertura do CSV local.
- Normalização de nomes depende de `teams`, `team_aliases` e `ml-service/data/team_aliases.json`; times não mapeados são reportados e não criados automaticamente.
- O motor de métricas não gera probabilidade de resultado nem explicações; ele alimenta os modelos de resultado, gols e o worker de explicações.

## Integração com ML e PalpitAI

A Etapa 3 já treina modelos supervisionados usando `match_features` como entrada e resultados históricos como labels. O resultado é salvo em `match_predictions`, `match_goal_predictions` e `match_score_probabilities`. A Etapa 4, no backend Go, usa essas tabelas junto de `match_features` para gerar explicações em linguagem natural sem recalcular probabilidades.
