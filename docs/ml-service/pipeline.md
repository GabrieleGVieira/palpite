# ML Service — Visão Geral do Pipeline

O pipeline de ML da PalpitAI é dividido em etapas sequenciais. Cada etapa produz dados que alimentam a próxima, culminando em previsões de resultado e placar que o backend Go consome do banco de dados.

---

## Visão geral das etapas

```mermaid
flowchart TD
    subgraph "Fontes de dados"
        CSV["CSVs históricos\n(results, goalscorers,\nshootouts, FIFA ranking)"]
        DB_MATCHES["world_cup_matches\n(banco)"]
        ALIASES["team_aliases.json\n(mapeamento de nomes)"]
    end

    subgraph "Etapa 1 — Seed"
        SEED["seed_teams.py\nCria teams + team_aliases no banco"]
    end

    subgraph "Etapa 2 — Métricas"
        METRICS["calculate_team_metrics.py\nCalcula ELO, ataque, defesa,\nforma, histórico de Copa"]
        FEATURES["calculate_match_features.py\nMonta 17 features por partida"]
        HIST["build_historical_training_features.py\nBackfill histórico para treino"]
    end

    subgraph "Etapa 3A — Modelo de resultado"
        TRAIN_R["train_model.py\nTreina classificador\n(HOME_WIN | DRAW | AWAY_WIN)"]
        PRED_R["predict_upcoming_matches.py\nGera probabilidades de resultado"]
    end

    subgraph "Etapa 3B — Modelo de gols"
        TRAIN_G["train_goals_model.py\nTreina dois PoissonRegressors\n(xG casa + xG visitante)"]
        PRED_G["predict_match_goals.py\nGera xG, placar mais provável,\nover 2.5, ambas marcam"]
    end

    subgraph "Etapa 3C — Calibração conjunta"
        CALIB["calibrate_score_result_predictions.py\nRepondera matriz de placares\nusando probabilidades de resultado"]
    end

    subgraph "Etapa 4 — Explicações"
        BACKEND["Backend Go\n(worker generate_prediction_explanations)"]
        GEMINI["Gemini API\nGera texto explicativo em PT"]
    end

    CSV --> SEED
    ALIASES --> SEED
    SEED --> METRICS
    CSV --> METRICS
    DB_MATCHES --> METRICS
    METRICS --> FEATURES
    DB_MATCHES --> FEATURES
    CSV --> HIST
    HIST --> TRAIN_R
    HIST --> TRAIN_G
    FEATURES --> PRED_R
    FEATURES --> PRED_G
    TRAIN_R --> PRED_R
    TRAIN_G --> PRED_G
    PRED_R --> CALIB
    PRED_G --> CALIB
    CALIB --> BACKEND
    BACKEND --> GEMINI
```

---

## Tabelas produzidas por etapa

| Etapa | Tabelas preenchidas |
| --- | --- |
| Seed | `teams`, `team_aliases` |
| Métricas | `team_metrics`, `team_metric_snapshots` |
| Features | `match_features`, `historical_matches` |
| Modelo resultado | `ml_models`, `prediction_runs`, `match_predictions` |
| Modelo gols | `goal_models`, `match_goal_predictions`, `match_score_probabilities` |
| Calibração | Atualiza `match_goal_predictions` e `match_score_probabilities` |
| Explicações | `prediction_explanations` |

---

## 1. Etapa 2 — Cálculo de métricas e features

```mermaid
flowchart TD
    A["results.csv\ngoalscorers.csv\nshootouts.csv"] --> B["Normaliza nomes\nvia team_aliases.json"]
    DB_WC["world_cup_matches finalizadas\n(banco)"] --> B

    B --> C["Filtra partidas\naté metric_date"]
    C --> D["calculate_team_metrics()\npor seleção"]
    D --> UPSERT_M["Upsert team_metrics\n(por time e data)"]
    D --> UPSERT_S["Insert team_metric_snapshots\n(componentes detalhados)"]

    UPSERT_M --> FE["calculate_match_features()"]
    DB_WC2["world_cup_matches futuras\n(banco)"] --> FE
    FIFA["ranking_fifa_historical.csv"] --> FE
    FE --> F17["17 features por partida"]
    F17 --> UPSERT_F["Upsert match_features"]
```

---

## 2. Etapa 3A — Treino e predição de resultado

```mermaid
flowchart TD
    MF["match_features"] --> DS["Dataset builder\n(join com historical_matches\npara labels)"]
    DS --> LABEL["Label builder\nhome > away → HOME_WIN\nhome < away → AWAY_WIN\nhome = away → DRAW"]
    LABEL --> SPLIT["Split temporal\n(sem vazamento de dados futuros)"]
    SPLIT --> TRAIN_SET["Train: até train_until"]
    SPLIT --> VAL_SET["Validation: val_from→val_until"]
    SPLIT --> TEST_SET["Test: test_from→test_until"]

    TRAIN_SET --> MODEL["Pipeline:\nSimpleImputer(median)\n→ StandardScaler\n→ LogisticRegression\n(class_weight=balanced)"]
    MODEL --> CALIB_M["CalibratedClassifierCV\n(sigmoid, cv=prefit\nse validation existe)"]
    CALIB_M --> EVAL["Avalia em TEST\naccuracy, log_loss, brier_score\nclassification_report"]
    EVAL --> ARTIFACT["Salva .joblib em models/"]
    ARTIFACT --> DB_REG["Registra em ml_models\n(name, version, metrics_json)"]

    ARTIFACT --> PRED["predict_upcoming_matches.py"]
    MF2["match_features (futuras)"] --> PRED
    PRED --> PROBA["model.predict_proba()\n[HOME_WIN, DRAW, AWAY_WIN]"]
    PROBA --> CONF["classify_confidence(max_prob)\n≥0.60 high | ≥0.45 medium | <0.45 low"]
    CONF --> SAVE_P["Upsert match_predictions\n(probabilidades + label + confidence)"]
```

---

## 3. Etapa 3B — Treino e predição de gols

```mermaid
flowchart TD
    MF["match_features"] --> DS2["Dataset builder\n(labels: home_score, away_score\nde historical_matches)"]
    DS2 --> SPLIT2["Split temporal"]
    SPLIT2 --> TRN["Train"]
    SPLIT2 --> TST["Test"]

    TRN --> HOME_M["PoissonRegressor α=0.001\n→ xG casa"]
    TRN --> AWAY_M["PoissonRegressor α=0.001\n→ xG visitante"]

    HOME_M --> ARTI2["Artefato:\n{ home_model, away_model }"]
    AWAY_M --> ARTI2
    ARTI2 --> EVAL2["MAE, RMSE (casa + visitante)\nMédia gols previstos vs reais"]
    EVAL2 --> DBREG2["Registra em goal_models"]

    ARTI2 --> PREDICT_G["predict_match_goals.py"]
    PREDICT_G --> XG["xG_casa, xG_visitante\n(clampado em [0.1, 5.0])"]
    XG --> MATRIX["Matriz de placares Poisson\n0x0 até 6x6\n(49 combinações)"]
    MATRIX --> DERIV["Probabilidades derivadas\n• over 1.5 / over 2.5\n• ambas marcam\n• placar mais provável"]
    DERIV --> TOP10["Top 10 placares\nordenados por probabilidade"]
    TOP10 --> SAVE_G["Upsert match_goal_predictions\n+ match_score_probabilities"]
```

---

## 4. Etapa 3C — Calibração conjunta

Combina as probabilidades do modelo de resultado com a matriz de placares do modelo de gols.

```mermaid
flowchart TD
    R_PROBS["match_predictions\n(HOME_WIN%, DRAW%, AWAY_WIN%)"] --> CALIB
    MATRIX["Matriz Poisson 7x7\n(do modelo de gols)"] --> CALIB

    CALIB["Calibrador ensemble"] --> BUCKET["Agrupa placares em 3 buckets:\n• HOME_WIN: home > away\n• DRAW: home = away\n• AWAY_WIN: home < away"]

    BUCKET --> WEIGHT["Para cada placar:\ncal_prob = poisson_prob × (result_prob / bucket_mass)"]

    WEIGHT --> RECALC["Recalcula:\n• top 10 placares calibrados\n• xG esperados\n• over 2.5 / ambas marcam"]

    RECALC --> UPDATE["Atualiza match_goal_predictions:\ncalibration_method, calibrated_at\nresult_model_id, result_probabilities\n+ match_score_probabilities"]
```

**Efeito da calibração:** um modelo de gols pode prever 1x1 como placar mais provável, mas se o modelo de resultado diz 62% HOME_WIN, a calibração redistribui probabilidade para placares com vitória do time da casa.

---

## 5. Job diário

O job diário recalcula métricas e features com a data de hoje, mantendo o banco atualizado à medida que novas partidas são disputadas.

```mermaid
flowchart TD
    START["run_daily_metrics_job.py\n--feature-to-date=2026-07-31"] --> DATE["metric_date = hoje\n(ou --metric-date especificado)"]

    DATE --> STEP1["1. calculate_and_save_team_metrics(metric_date)"]
    STEP1 --> A1["Carrega teams + aliases do banco"]
    A1 --> A2["Carrega CSVs + partidas da Copa finalizadas"]
    A2 --> A3["Combina fontes, filtra até metric_date"]
    A3 --> A4["Calcula métricas por seleção"]
    A4 --> A5["Upsert team_metrics\nInsert team_metric_snapshots"]

    A5 --> STEP2["2. calculate_and_save_match_features(\n  from_date = metric_date + 1d\n  to_date = feature_to_date)"]
    STEP2 --> B1["Carrega partidas futuras do banco"]
    B1 --> B2["Para cada partida:\n  busca métricas mais recentes\n  busca ranking FIFA anterior"]
    B2 --> B3["Monta 17 features"]
    B3 --> B4["Upsert match_features"]

    B4 --> REPORT["Imprime resumo:\ntimes processados, features salvas\ntimes sem mapeamento"]
```
