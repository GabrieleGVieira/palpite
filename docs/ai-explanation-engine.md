# PalpitAI AI Explanation Engine

## Objetivo

A Etapa 4 usa IA apenas para explicar previsões já calculadas. A IA não decide vencedor, probabilidades, placar provável, gols esperados, métricas, rankings ou confiança numérica.

O app nunca chama IA diretamente. O app lê explicações previamente salvas em `prediction_explanations`.

## O que a IA faz

- Transforma dados de `match_predictions`, `match_goal_predictions`, `match_score_probabilities` e `match_features` em texto curto.
- Escreve em português do Brasil.
- Retorna JSON estruturado validado pelo backend.

## O que a IA não faz

- Não calcula probabilidades.
- Não altera placar provável.
- Não inventa estatísticas.
- Não gera recomendação de aposta financeira.
- Não roda dentro de request HTTP do app.

## Fluxo batch

1. O worker busca partidas no intervalo informado.
2. Ignora partidas sem `match_prediction` ou sem `match_goal_prediction`.
3. Monta `input_snapshot`.
4. Chama o provider de IA.
5. Valida JSON, campos obrigatórios, `bet_style` e linguagem proibida.
6. Salva em `prediction_explanations`.

## Variáveis de ambiente

```bash
OPENAI_API_KEY=...
OPENAI_MODEL=gpt-4.1-mini
OPENAI_TIMEOUT_SECONDS=30
```

`OPENAI_MODEL` pode ser trocado sem alterar código.

## Como rodar

```bash
cd backend
make migrate
go run ./internal/workers/generate_prediction_explanations \
  --from-date=2026-06-01 \
  --to-date=2026-07-31 \
  --limit=50
```

Saída esperada:

```text
AI explanation generation finished
Processed: 50
Generated: 43
Skipped: 5
Failed: 2
Prompt version: prediction-explanation-v1
```

## Prompt

O prompt recebe dados estruturados da partida, probabilidades, expected goals, top placares e métricas principais. Ele instrui a IA a retornar somente JSON, sem markdown, sem alterar números e sem linguagem de certeza.

## Resposta

Formato obrigatório:

```json
{
  "summary": "string curta com no máximo 240 caracteres",
  "main_reasons": ["motivo 1", "motivo 2"],
  "risk_alert": "string curta ou null",
  "bet_style": "conservative",
  "user_tip": "string curta"
}
```

`bet_style` é validado pelo backend:

- `conservative`: maior probabilidade >= 60
- `moderate`: maior probabilidade >= 45 e < 60
- `risky`: maior probabilidade < 45

## Tabelas

- `prediction_explanations`
- `match_predictions`
- `match_goal_predictions`
- `match_score_probabilities`
- `match_features`

`input_snapshot` e `raw_response` são salvos para auditoria e debug.

## Cuidados

- Não logar `OPENAI_API_KEY`.
- Não gerar explicação quando faltarem previsões essenciais.
- Não sobrescrever explicação `generated` com a mesma `prompt_version`.
- Permitir reprocessar registros `failed` ou `skipped`.

## Próximos passos

- Expor explicação no endpoint futuro de leitura de previsão.
- Adicionar dashboard operacional para falhas de geração.
- Versionar novos prompts com `prediction-explanation-v2`.

