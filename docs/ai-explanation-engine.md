# PalpitAI Explanation Engine

## Objetivo

A Etapa 4 usa a PalpitAI apenas para explicar previsões já calculadas. A PalpitAI não decide vencedor, probabilidades, placar provável, gols esperados, métricas, rankings ou confiança numérica.

O app nunca chama a PalpitAI diretamente. O app lê explicações previamente salvas em `prediction_explanations`.

## O que a PalpitAI faz

- Transforma dados de `match_predictions`, `match_goal_predictions`, `match_score_probabilities` e `match_features` em texto curto.
- Escreve em português do Brasil.
- Retorna JSON estruturado validado pelo backend.

## O que a PalpitAI não faz

- Não calcula probabilidades.
- Não altera placar provável.
- Não inventa estatísticas.
- Não gera recomendação de aposta financeira.
- Não roda dentro de request HTTP do app.

## Fluxo batch

1. O worker busca partidas no intervalo informado.
2. Ignora partidas sem `match_prediction` ou sem `match_goal_prediction`.
3. Monta `input_snapshot`.
4. Chama o provider da PalpitAI.
5. Valida JSON, campos obrigatórios, `bet_style` e linguagem proibida.
6. Salva em `prediction_explanations`.

## Variáveis de ambiente

```bash
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-2.5-flash
GEMINI_RATE_LIMIT_COOLDOWN_SECONDS=1800
GEMINI_RATE_LIMIT_MAX_WAITS=1
GEMINI_REQUEST_DELAY_SECONDS=15
GEMINI_TIMEOUT_SECONDS=30
```

`GEMINI_MODEL` pode ser trocado sem alterar código.
`GEMINI_REQUEST_DELAY_SECONDS` controla o intervalo entre chamadas do worker para respeitar limites de RPM.
`GEMINI_RATE_LIMIT_COOLDOWN_SECONDS` controla a pausa longa antes de retentar o mesmo jogo quando a quota estoura.
`GEMINI_RATE_LIMIT_MAX_WAITS` limita quantas pausas longas podem acontecer em uma execução.

## Como rodar

```bash
cd backend
make migrate
make explanations FROM_DATE=2026-06-01 TO_DATE=2026-07-31 LIMIT=50
```

`FROM_DATE` e `TO_DATE` são obrigatórios. `LIMIT` é opcional (default: 15).

Equivalente direto:

```bash
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
Rate limited: false
Rate limit waits: 0
Prompt version: prediction-explanation-v1
```

## Prompt

O prompt recebe dados estruturados da partida, probabilidades, expected goals, top placares e métricas principais. Ele instrui a PalpitAI a retornar somente JSON, sem markdown, sem alterar números e sem linguagem de certeza.

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

- Não logar `GEMINI_API_KEY`.
- Preservar explicações `generated` quando uma nova tentativa falhar ou for limitada pela API.
- Ao atingir rate limit persistente da Gemini, esperar o cooldown configurado e retentar o mesmo jogo.
- Parar o batch se o limite persistir depois das pausas configuradas.
- Não gerar explicação quando faltarem previsões essenciais.
- Sobrescrever explicação `generated` somente quando a nova geração também terminar como `generated`.
- Permitir reprocessar registros `failed` ou `skipped`.

## Endpoint de leitura

As explicações geradas são retornadas pelo endpoint:

```
GET /api/v1/matches/{matchID}/prediction
```

O campo `explanation` é opcional na resposta: se ainda não houver explicação gerada para a partida, o campo é omitido. O app lida com a ausência sem erro.

## Próximos passos

- Adicionar dashboard operacional para falhas de geração.
- Versionar novos prompts com `prediction-explanation-v2`.
