package entity

type Company struct {
	ID          string `json:"cnpj,omitzero"`
	Name        string `json:"razao_social,omitzero"`
	TradingName string `json:"nome_fantasia,omitzero"`
}
