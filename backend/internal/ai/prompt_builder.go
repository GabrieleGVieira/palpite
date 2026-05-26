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
