# PalpitAI Prediction Engine

## Target labels

O modelo prevê três classes:

- `HOME_WIN`: placar final do time A/casa maior que o visitante.
- `DRAW`: placares iguais.
- `AWAY_WIN`: placar final do visitante/time B maior que o time A.

Os placares finais são usados apenas para criar o label. Eles não entram no vetor de features.

## Features usadas

O pipeline consome `match_features` com ordem fixa em `app/ml/feature_columns.py`:

- `elo_diff`
- `fifa_rank_diff`
- `home_elo_score`, `away_elo_score`
- `home_attack_score`, `away_attack_score`
- `home_defense_score`, `away_defense_score`
- `home_recent_form_score`, `away_recent_form_score`
- `home_avg_goals_scored`, `away_avg_goals_scored`
- `home_avg_goals_conceded`, `away_avg_goals_conceded`
- `home_world_cup_history_score`, `away_world_cup_history_score`
- `neutral`

As features devem representar somente informação conhecida antes da data do jogo. Rankings, ELO e métricas precisam ser calculados com corte temporal anterior à partida.

## Split temporal

O split principal não é aleatório:

- treino: jogos até `--train-until`
- validação: `--validation-from` até `--validation-until`, quando informado
- teste: `--test-from` até `--test-until`

O código rejeita splits em que treino invade validação ou teste.

## Calibração

As probabilidades são calibradas com `CalibratedClassifierCV`, usando `sigmoid` como padrão. Quando há validação temporal suficiente, a calibração usa esse período. Caso contrário, o treino usa calibração cruzada dentro do conjunto de treino quando há amostras suficientes por classe.

## Métricas

`metrics_json` em `ml_models` registra:

- `accuracy`
- `balanced_accuracy`
- `log_loss`
- `brier_score`
- `classification_report`
- `confusion_matrix`
- `number_of_train_samples`
- `number_of_test_samples`
- `class_distribution`

## Fluxo batch

1. `train_model.py` carrega `match_features`, monta labels, aplica split temporal, treina, calibra, avalia e salva `.joblib`.
2. O mesmo script registra o modelo em `ml_models`.
3. `predict_upcoming_matches.py` carrega um modelo ativo, busca jogos futuros em `match_features`, gera probabilidades e salva em `match_predictions`.
4. Cada execução de predição é registrada em `prediction_runs`.

## Integração com Go Backend

O Go Backend continua sendo a API principal do app. Nesta etapa, o Python acessa o PostgreSQL para treino e batch prediction. O backend lê `match_predictions` via `backend/internal/predictions` e pode expor endpoints futuros sem chamar o modelo em tempo real.

Um client HTTP opcional em `backend/internal/ml` está preparado para `POST /predict`, mas o fluxo recomendado para o app é leitura das previsões já persistidas.

## Por que LLM não entra nesta etapa

Esta etapa produz previsões estruturadas e probabilísticas por modelo supervisionado. Não há chamada a LLM, geração de explicações ou uso de API externa. A camada explicativa fica reservada para a Etapa 4, consumindo as previsões salvas pelo pipeline de ML.

