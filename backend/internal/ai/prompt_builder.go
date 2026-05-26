package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Prompt struct {
	System string
	User   string
}

func BuildPredictionExplanationPrompt(input ExplanationPromptInput) (Prompt, error) {
	payload, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return Prompt{}, err
	}

	system := strings.TrimSpace(`
Voce e um assistente do PalpitAI. Sua tarefa e explicar previsoes ja calculadas para usuarios de bolao.
Responda sempre em portugues do Brasil, com linguagem simples, curta e amigavel.
Nao calcule vencedor, probabilidades, placar provavel, gols esperados, rankings ou confianca numerica.
Nao altere nenhum numero recebido. Nao invente estatisticas.
Nao sugira aposta real com dinheiro e nao use termos como "garantido", "certeza", "aposta segura" ou "lucro".
Evite certeza absoluta. Se a confidence for low, mencione incerteza no risk_alert.
Retorne apenas JSON valido no formato solicitado, sem markdown.
`)

	user := fmt.Sprintf(strings.TrimSpace(`
Explique a previsao abaixo para um usuario comum de bolao.

Regras:
- summary: no maximo 240 caracteres.
- main_reasons: entre 2 e 4 itens.
- bet_style deve seguir a maior probabilidade: conservative se >= 60, moderate se >= 45 e < 60, risky se < 45.
- user_tip deve ajudar a interpretar o palpite, sem prometer acerto.
- risk_alert pode ser null, exceto quando confidence for low.

Dados calculados:
%s
`), string(payload))

	return Prompt{System: system, User: user}, nil
}

func BuildBatchPredictionExplanationPrompt(inputs []ExplanationPromptInput) (Prompt, error) {
	payload, err := json.MarshalIndent(inputs, "", "  ")
	if err != nil {
		return Prompt{}, err
	}

	system := strings.TrimSpace(`
Voce e um assistente do PalpitAI. Sua tarefa e explicar previsoes ja calculadas para usuarios de bolao.
Responda sempre em portugues do Brasil, com linguagem simples, curta e amigavel.
Voce deve apenas explicar e enriquecer a previsao recebida.
Nao recalcule vencedor, probabilidades, placar provavel, gols esperados, rankings ou confianca numerica.
Nao altere nenhum numero recebido. Nao invente estatisticas, lesoes, noticias, escalacoes ou contexto externo.
Nao sugira aposta real com dinheiro e nao use termos como "garantido", "certeza", "aposta segura" ou "lucro".
Retorne apenas JSON valido no formato solicitado, sem markdown.
`)

	user := fmt.Sprintf(strings.TrimSpace(`
Explique as previsoes abaixo para usuarios comuns de bolao.

Regras:
- Voce recebeu exatamente %d partidas.
- Retorne um objeto com a chave predictions.
- Voce DEVE retornar exatamente %d itens no array predictions.
- predictions deve ser uma lista de objetos com match_id, explanation, key_factors e risk_level.
- Cada item deve conter o mesmo match_id recebido.
- Nao omita partidas.
- explanation deve ter no maximo 500 caracteres.
- key_factors deve ter de 2 a 4 itens curtos.
- risk_level deve ser low, medium ou high.
- Explique a previsao do modelo; nao recalcule nem substitua as probabilidades.
- Se nao tiver certeza, gere uma explicacao usando apenas os dados fornecidos.
- Se algum dado estiver ausente, explique com base apenas nos dados disponiveis.
- Retorne apenas JSON valido.

Dados calculados:
%s
`), len(inputs), len(inputs), string(payload))

	return Prompt{System: system, User: user}, nil
}

func ExplanationJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"summary", "main_reasons", "risk_alert", "bet_style", "user_tip"},
		"properties": map[string]any{
			"summary": map[string]any{
				"type":      "string",
				"maxLength": 240,
			},
			"main_reasons": map[string]any{
				"type":     "array",
				"minItems": 2,
				"maxItems": 4,
				"items": map[string]any{
					"type":      "string",
					"maxLength": 180,
				},
			},
			"risk_alert": map[string]any{
				"type":      []string{"string", "null"},
				"maxLength": 180,
			},
			"bet_style": map[string]any{
				"type": "string",
				"enum": []string{"conservative", "moderate", "risky"},
			},
			"user_tip": map[string]any{
				"type":      "string",
				"maxLength": 220,
			},
		},
	}
}

func GeminiExplanationSchema() map[string]any {
	return map[string]any{
		"type":     "OBJECT",
		"required": []string{"summary", "main_reasons", "risk_alert", "bet_style", "user_tip"},
		"properties": map[string]any{
			"summary": map[string]any{
				"type":      "STRING",
				"maxLength": 240,
			},
			"main_reasons": map[string]any{
				"type":     "ARRAY",
				"minItems": 2,
				"maxItems": 4,
				"items": map[string]any{
					"type":      "STRING",
					"maxLength": 180,
				},
			},
			"risk_alert": map[string]any{
				"type":      "STRING",
				"nullable":  true,
				"maxLength": 180,
			},
			"bet_style": map[string]any{
				"type": "STRING",
				"enum": []string{"conservative", "moderate", "risky"},
			},
			"user_tip": map[string]any{
				"type":      "STRING",
				"maxLength": 220,
			},
		},
	}
}

func GeminiBatchExplanationSchema() map[string]any {
	return map[string]any{
		"type":     "OBJECT",
		"required": []string{"predictions"},
		"properties": map[string]any{
			"predictions": map[string]any{
				"type": "ARRAY",
				"items": map[string]any{
					"type":     "OBJECT",
					"required": []string{"match_id", "explanation", "key_factors", "risk_level"},
					"properties": map[string]any{
						"match_id": map[string]any{
							"type": "STRING",
						},
						"explanation": map[string]any{
							"type":      "STRING",
							"maxLength": 500,
						},
						"key_factors": map[string]any{
							"type":     "ARRAY",
							"minItems": 2,
							"maxItems": 4,
							"items": map[string]any{
								"type":      "STRING",
								"maxLength": 180,
							},
						},
						"risk_level": map[string]any{
							"type": "STRING",
							"enum": []string{"low", "medium", "high"},
						},
					},
				},
			},
		},
	}
}
