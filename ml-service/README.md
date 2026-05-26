# PalpitAI ML Service

Etapa 2 calcula métricas agregadas de seleções e features de partidas para serem consumidas pelo backend Go na Etapa 3.

## Fonte histórica local

Os CSVs em `ml-service/data` são usados como fonte histórica estática e não são persistidos integralmente no PostgreSQL. Os scripts carregam esses arquivos uma vez por execução, calculam as métricas em memória e salvam somente os agregados em `team_metrics`, `team_metric_snapshots` e `match_features`.

O arquivo `data/team_aliases.json` centraliza os aliases usados para casar nomes dos CSVs, nomes traduzidos do backend e times existentes no banco. Ele foi montado a partir do mapper em `backend/internal/utils/mapper.go` e de `data/former_names.csv`.

## Configuração

```bash
cd ml-service
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

Variáveis:

```bash
export DATABASE_URL='postgresql://...'
```

## Popular times e aliases

```bash
python ml-service/app/scripts/seed_teams.py
```

O seed lê `data/team_aliases.json`, cria os times em `teams` usando `db_name` e recria os aliases em `team_aliases`.

## Calcular métricas de seleções

```bash
python ml-service/app/scripts/calculate_team_metrics.py --metric-date=2026-06-01
```

O script carrega `data/results.csv`, `data/goalscorers.csv` e `data/shootouts.csv`, normaliza os nomes com `data/team_aliases.json`, filtra jogos até `metric-date`, calcula métricas apenas para os times-alvo do mapper e faz upsert em `team_metrics`. Partidas contra seleções fora do mapper continuam entrando no histórico com IDs locais em memória, sem serem salvas como times no banco. Também grava snapshots agregados em `team_metric_snapshots` com componentes como profundidade de artilharia, dependência de pênaltis e aproveitamento em disputas por pênaltis.

## Calcular features de partidas

```bash
python ml-service/app/scripts/calculate_match_features.py --from-date=2026-06-01 --to-date=2026-07-31
```

O script busca partidas alvo em `world_cup_matches`, resolve os nomes com `data/team_aliases.json`, usa as métricas mais recentes antes da data do jogo, busca o ranking FIFA anterior à data da partida em `data/ranking_fifa_historical.csv` e salva em `match_features`.

## Job diário

```bash
python ml-service/app/scripts/run_daily_metrics_job.py --feature-to-date=2026-07-31
```

Por padrão, o job usa a data atual como `metric-date`, inclui partidas finalizadas de `world_cup_matches` no histórico em memória, recalcula `team_metrics` e recalcula `match_features` a partir do dia seguinte até `feature-to-date`. Para reprocessar uma data específica:

```bash
python ml-service/app/scripts/run_daily_metrics_job.py --metric-date=2026-06-20 --feature-to-date=2026-07-31
```

## Tabelas preenchidas

- `team_metrics`
- `team_metric_snapshots`
- `match_features`

## Próximos passos

Na Etapa 3, o backend pode consumir `match_features` para alimentar modelos estatísticos ou ML real. Esta etapa não chama LLM, não gera explicações e não produz previsões finais.

## Etapa 3: ML real para previsão de resultados

Objetivo: treinar e servir um modelo supervisionado para prever `HOME_WIN`, `DRAW` ou `AWAY_WIN` com probabilidades calibradas. Esta etapa não chama LLM, não gera explicações textuais e não usa APIs externas.

### Dataset

O treino usa `match_features` como entrada. O label `target_result` é montado a partir de placares finais em `world_cup_matches` via `match_id` ou, se existir, `historical_matches` com colunas compatíveis.

Features usadas:

- `elo_diff`, `fifa_rank_diff`
- métricas de força, ataque, defesa, forma recente e histórico de Copa para casa/fora
- médias de gols marcados/sofridos para casa/fora
- `neutral`

`home_score` e `away_score` nunca entram como features.

### Treinar e avaliar

Antes do primeiro treino, gere features e labels históricos a partir dos CSVs locais:

```bash
python ml-service/app/scripts/build_historical_training_features.py \
  --from-date 2002-01-01 \
  --to-date 2025-12-31 \
  --tournament-contains "world cup"
```

Esse comando salva features pré-jogo em `match_features` e placares finais em `historical_matches`. As métricas de cada partida usam somente jogos anteriores à data do jogo.

```bash
python ml-service/app/scripts/train_model.py \
  --model-name palpitai-result-model \
  --version v1.0.0 \
  --train-until 2018-12-31 \
  --test-from 2019-01-01 \
  --test-until 2022-12-31 \
  --algorithm logistic_regression
```

O split é temporal. O script treina o modelo, calibra probabilidades quando há dados suficientes, avalia em teste, salva o artefato em `ml-service/models/` e registra o modelo em `ml_models` com `metrics_json`.

Métricas salvas: accuracy, balanced accuracy, log loss, Brier score multiclass, classification report, matriz de confusão, tamanhos de treino/teste e distribuição de classes.

### Gerar previsões futuras

```bash
python ml-service/app/scripts/predict_upcoming_matches.py \
  --model-name palpitai-result-model \
  --version v1.0.0 \
  --from-date 2026-06-01 \
  --to-date 2026-07-31
```

O script carrega o artefato registrado, busca `match_features`, usa a mesma ordem de features do treino, valida probabilidades, classifica confiança e salva em `match_predictions`. Cada execução cria um registro em `prediction_runs`.

### API opcional

O endpoint `POST /predict` existe para inferência síncrona se `ML_MODEL_ARTIFACT` apontar para um `.joblib`. O fluxo principal da etapa continua sendo batch, com o Go Backend lendo as previsões no banco.

### Tabelas preenchidas

- `ml_models`
- `prediction_runs`
- `match_predictions`

### Limitações conhecidas

- O score sugerido é heurístico nesta etapa.
- XGBoost não foi incluído para manter a stack simples; `HistGradientBoostingClassifier` fica disponível como alternativa inicial.
- A qualidade do modelo depende das features históricas representarem somente informação disponível antes da partida.

### Próximos passos

Etapa 4: adicionar IA explicativa consumindo previsões já geradas, sem alterar o pipeline supervisionado.

## Etapa 3B — Modelo de gols

Objetivo: prever `expected_home_goals`, `expected_away_goals` e os placares mais prováveis com probabilidade de cada placar. Esta etapa também não chama LLM, não gera explicações textuais e não usa APIs externas.

Diferença entre modelos:

- Etapa 3A: classifica resultado (`HOME_WIN`, `DRAW`, `AWAY_WIN`).
- Etapa 3B: estima gols esperados e placares prováveis.

Antes do primeiro treino, rode o backfill histórico da Etapa 3A se ainda não houver labels em `historical_matches`:

```bash
python ml-service/app/scripts/build_historical_training_features.py \
  --from-date 2002-01-01 \
  --to-date 2025-12-31 \
  --tournament-contains "world cup"
```

Treino:

```bash
python ml-service/app/scripts/train_goals_model.py \
  --model-name palpitai-goals-model \
  --version v1.0.0 \
  --train-until 2018-12-31 \
  --test-from 2019-01-01 \
  --test-until 2022-12-31 \
  --algorithm poisson_regression
```

O script treina dois modelos `PoissonRegressor`: um para gols do time A/casa e outro para gols do time B/visitante. O artefato fica em `ml-service/models/` e o registro do modelo é salvo em `goal_models`.

Predição:

```bash
python ml-service/app/scripts/predict_match_goals.py \
  --model-name palpitai-goals-model \
  --version v1.0.0 \
  --from-date 2026-06-01 \
  --to-date 2026-07-31 \
  --top-scores 10
```

O script salva:

- `match_goal_predictions`: expected goals, placar mais provável, over 1.5/2.5 e ambas marcam.
- `match_score_probabilities`: top placares ordenados por probabilidade.

Limitações conhecidas:

- A matriz de placares é truncada em 0x0 até 6x6, então existe probabilidade residual para placares acima de 6 gols.
- O modelo de gols não força coerência com o modelo de resultado da Etapa 3A.
- Uma etapa futura pode combinar 3A e 3B para calibrar placar e resultado conjuntamente.

## Etapa 3C — Calibração conjunta resultado + placar

Depois de rodar a Etapa 3A e a Etapa 3B, repondere os placares usando as probabilidades de resultado:

```bash
python ml-service/app/scripts/calibrate_score_result_predictions.py \
  --result-model-name palpitai-result-model \
  --result-version v1.0.0 \
  --goal-model-name palpitai-goals-model \
  --goal-version v1.0.0 \
  --from-date 2026-06-01 \
  --to-date 2026-07-31 \
  --top-scores 10
```

Esse script recalcula a matriz 0x0 até 6x6 do modelo de gols e repondera cada bucket:

- placares com vitória do time A somam a probabilidade `HOME_WIN` da Etapa 3A;
- empates somam a probabilidade `DRAW`;
- placares com vitória do time B somam a probabilidade `AWAY_WIN`.

Ele atualiza `match_goal_predictions` e `match_score_probabilities`, preenchendo `result_model_id`, `result_probabilities`, `calibration_method`, `score_probability_mass` e `calibrated_at`.
