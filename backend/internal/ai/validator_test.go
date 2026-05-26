package ai

import "testing"

func TestValidatorAcceptsValidJSON(t *testing.T) {
	raw := []byte(`{"summary":"Brasil aparece como leve favorito.","main_reasons":["Tem leve vantagem nas metricas.","Os placares indicam jogo equilibrado."],"risk_alert":"O empate ainda tem chance relevante.","bet_style":"moderate","user_tip":"Interprete como favoritismo leve, nao como garantia."}`)
	response, err := ParseAndValidateExplanation(raw, validPromptInput())
	if err != nil {
		t.Fatalf("ParseAndValidateExplanation() error = %v", err)
	}
	if response.BetStyle != "moderate" {
		t.Fatalf("BetStyle = %s", response.BetStyle)
	}
}

func TestValidatorRejectsInvalidBetStyle(t *testing.T) {
	raw := []byte(`{"summary":"Brasil aparece como leve favorito.","main_reasons":["Tem leve vantagem nas metricas.","Os placares indicam jogo equilibrado."],"risk_alert":"O empate ainda tem chance relevante.","bet_style":"safe","user_tip":"Interprete com cuidado."}`)
	if _, err := ParseAndValidateExplanation(raw, validPromptInput()); err == nil {
		t.Fatalf("expected invalid bet_style error")
	}
}

func TestValidatorRejectsEmptyMainReasons(t *testing.T) {
	raw := []byte(`{"summary":"Brasil aparece como leve favorito.","main_reasons":[],"risk_alert":"O empate ainda tem chance relevante.","bet_style":"moderate","user_tip":"Interprete com cuidado."}`)
	if _, err := ParseAndValidateExplanation(raw, validPromptInput()); err == nil {
		t.Fatalf("expected main_reasons error")
	}
}
