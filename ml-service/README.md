# PalpitAI ML Service

Pipeline de machine learning e API de inferência da PalpitAI. Calcula métricas de seleções, treina modelos supervisionados para prever resultado e placar de partidas da Copa do Mundo e serve análises para o backend Go.

## O que é

O ML Service é responsável pela inteligência preditiva da PalpitAI. Opera em quatro etapas sequenciais: cálculo de métricas históricas de seleções, treinamento de modelos de resultado (classificação) e de gols (regressão Poisson), calibração conjunta dos modelos, e geração de previsões futuras salvas no banco. O backend Go consome essas previsões diretamente do PostgreSQL e as expõe no app como análises complementares.

As explicações em linguagem natural não são geradas pelo serviço Python. Elas são produzidas por um worker do backend Go depois que `match_predictions`, `match_goal_predictions` e `match_score_probabilities` já existem no banco.

## Tecnologias

- **Python 3.11+**
- **scikit-learn** — modelos de classificação (Logistic Regression, HistGradientBoosting) e regressão Poisson
- **pandas + numpy** — manipulação e feature engineering
- **joblib** — serialização de artefatos de modelo
- **FastAPI + uvicorn** — API de inferência síncrona
- **psycopg** — driver PostgreSQL
- **pytest** — testes

## Fontes de dados

| Fonte                              | Uso                                                      |
| ---------------------------------- | -------------------------------------------------------- |
| `data/results.csv`                 | Histórico de resultados de partidas internacionais       |
| `data/goalscorers.csv`             | Artilheiros por partida                                  |
| `data/shootouts.csv`               | Disputas de pênaltis                                     |
| `data/ranking_fifa_historical.csv` | Ranking FIFA histórico por data                          |
| `data/team_aliases.json`           | Mapeamento de nomes entre CSVs, backend e banco          |
| `world_cup_matches` (PostgreSQL)   | Partidas da Copa 2026 com datas e times                  |
| `team_metrics` (PostgreSQL)        | Métricas calculadas usadas como features                 |
| `match_features` (PostgreSQL)      | Features pré-jogo para treino e inferência               |

## Configuração

```bash
cd ml-service
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

```env
DATABASE_URL=postgresql://usuario:senha@host:porta/banco
```

## Como rodar

### API de inferência

```bash
uvicorn app.api.main:app --reload
```

`POST /predict` realiza inferência síncrona quando `ML_MODEL_ARTIFACT` aponta para um `.joblib`.

### Pipeline batch (ordem de execução)

**1. Popular times e aliases:**

```bash
python app/scripts/seed_teams.py
```

**2. Calcular métricas de seleções:**

```bash
python app/scripts/calculate_team_metrics.py --metric-date=2026-06-01
```

**3. Calcular features das partidas:**

```bash
python app/scripts/calculate_match_features.py --from-date=2026-06-01 --to-date=2026-07-31
```

**4. Gerar features históricas para treino:**

```bash
python app/scripts/build_historical_training_features.py \
  --from-date 2002-01-01 \
  --to-date 2025-12-31 \
  --tournament-contains "world cup"
```

**5. Treinar modelo de resultado:**

```bash
python app/scripts/train_model.py \
  --model-name palpite-result-model \
  --version v1.0.0 \
  --train-until 2018-12-31 \
  --test-from 2019-01-01 \
  --test-until 2022-12-31 \
  --algorithm logistic_regression
```

**6. Treinar modelo de gols:**

```bash
python app/scripts/train_goals_model.py \
  --model-name palpite-goals-model \
  --version v1.0.0 \
  --train-until 2018-12-31 \
  --test-from 2019-01-01 \
  --test-until 2022-12-31 \
  --algorithm poisson_regression
```

**7. Gerar previsões de resultado:**

```bash
python app/scripts/predict_upcoming_matches.py \
  --model-name palpite-result-model \
  --version v1.0.0 \
  --from-date 2026-06-01 \
  --to-date 2026-07-31
```

**8. Gerar previsões de gols:**

```bash
python app/scripts/predict_match_goals.py \
  --model-name palpite-goals-model \
  --version v1.0.0 \
  --from-date 2026-06-01 \
  --to-date 2026-07-31 \
  --top-scores 10
```

**9. Calibração conjunta resultado + placar:**

```bash
python app/scripts/calibrate_score_result_predictions.py \
  --result-model-name palpite-result-model \
  --result-version v1.0.0 \
  --goal-model-name palpite-goals-model \
  --goal-version v1.0.0 \
  --from-date 2026-06-01 \
  --to-date 2026-07-31 \
  --top-scores 10
```

**Job diário (recalcula métricas e features com a data atual):**

```bash
python app/scripts/run_daily_metrics_job.py --feature-to-date=2026-07-31
```

## Estrutura

```text
ml-service/
├── app/
│   ├── api/        # FastAPI — endpoint de inferência síncrona
│   ├── metrics/    # cálculo de métricas de seleções
│   ├── ml/         # treino e predição de resultado
│   ├── goals/      # treino e predição de gols
│   ├── ensemble/   # calibração conjunta resultado + placar
│   └── scripts/    # scripts batch do pipeline
├── data/           # CSVs históricos e aliases de times
├── models/         # artefatos .joblib dos modelos treinados
└── tests/          # suite de testes pytest
```

## Tabelas preenchidas no banco

| Tabela                      | Conteúdo                                          |
| --------------------------- | ------------------------------------------------- |
| `teams`, `team_aliases`     | Seleções e mapeamento de nomes                    |
| `team_metrics`              | Métricas agregadas por seleção e data             |
| `team_metric_snapshots`     | Snapshots com componentes detalhados de métricas  |
| `match_features`            | Features pré-jogo para treino e inferência        |
| `historical_matches`        | Placares históricos para labels de treino         |
| `ml_models`                 | Registro de modelos treinados com métricas        |
| `prediction_runs`           | Log de execuções de predição batch                |
| `match_predictions`         | Probabilidades de resultado por partida           |
| `goal_models`               | Registro de modelos de gols treinados             |
| `match_goal_predictions`    | Expected goals, placar mais provável e mercados   |
| `match_score_probabilities` | Top placares com probabilidades calibradas        |

## Qualidade

```bash
pytest
```
