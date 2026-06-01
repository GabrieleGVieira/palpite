# PalpitAI Goals Prediction Engine

## Objetivo

A Etapa 3B estima gols esperados e placares prováveis sem LLM. O output principal é:

- `expected_home_goals`
- `expected_away_goals`
- placar mais provável
- top placares com probabilidades
- over 1.5, over 2.5 e ambas marcam

## Expected goals

O modelo treinado padrão usa dois `PoissonRegressor` do scikit-learn:

- um modelo para `home_score`
- um modelo para `away_score`

Os placares finais entram apenas como labels. `home_score` e `away_score` nunca entram como features.

## Features

As features vêm de `match_features` e seguem ordem fixa em `app/goals/feature_columns.py`. A inferência sempre reordena as colunas para a mesma ordem salva no artefato.

## Split temporal

O treino usa split temporal:

- treino: jogos até `--train-until`
- teste: `--test-from` até `--test-until`

Não há split aleatório como validação principal, evitando vazamento de dados futuros.

## Placares prováveis

Com `expected_home_goals` e `expected_away_goals`, o motor calcula:

```text
P(home_score = h) = poisson.pmf(h, expected_home_goals)
P(away_score = a) = poisson.pmf(a, expected_away_goals)
P(h x a) = P(home = h) * P(away = a)
```

A matriz padrão vai de 0x0 até 6x6. Os top N placares são ordenados por probabilidade decrescente e salvos em `match_score_probabilities`.

## Over e ambas marcam

- `over_1_5_probability`: soma de placares com total de gols maior que 1.5.
- `over_2_5_probability`: soma de placares com total de gols maior que 2.5.
- `both_teams_score_probability`: soma de placares em que ambos os times marcam pelo menos 1 gol.

Essas probabilidades são calculadas sobre a matriz truncada 0x0 até 6x6. A probabilidade residual representa placares fora da matriz.

## Tabelas

- `goal_models`: registry dos modelos de gols.
- `match_goal_predictions`: expected goals e agregados por partida.
- `match_score_probabilities`: top placares por previsão.

## Integração com 3A

A Etapa 3A prevê resultado em classes (`HOME_WIN`, `DRAW`, `AWAY_WIN`). A Etapa 3B prevê gols e placares. A calibração conjunta 3C combina os dois outputs para ajustar coerência entre resultado, probabilidades e placar sugerido.

## Calibração conjunta 3A + 3B

A calibração conjunta é um pós-processamento batch. Para cada partida com previsão de resultado e previsão de gols:

1. Recalcula a matriz Poisson 0x0 até 6x6 usando `expected_home_goals` e `expected_away_goals`.
2. Agrupa os placares em três buckets: `HOME_WIN`, `DRAW`, `AWAY_WIN`.
3. Repondera cada placar para que a soma do bucket seja igual à probabilidade da Etapa 3A.
4. Recalcula placar mais provável, over 1.5, over 2.5, ambas marcam e top placares.

Isso preserva a estrutura relativa dos placares dentro de cada bucket do modelo de gols, mas força coerência com a previsão de resultado.

Comando:

```bash
python ml-service/app/scripts/calibrate_score_result_predictions.py \
  --result-model-name palpite-result-model \
  --result-version v1.0.0 \
  --goal-model-name palpite-goals-model \
  --goal-version v1.0.0 \
  --from-date 2026-06-01 \
  --to-date 2026-07-31 \
  --top-scores 10
```

## Limitações

- A independência entre gols dos dois times é uma aproximação.
- A matriz 0x0 até 6x6 não soma exatamente 1 quando há massa de probabilidade acima de 6 gols.
- O modelo depende da qualidade temporal de `match_features`; rankings e métricas precisam ser anteriores ao jogo.
- A calibração conjunta repondera a matriz truncada; probabilidades acima de 6 gols ficam fora desse ajuste.
