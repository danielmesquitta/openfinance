package entity

type Company struct {
	ID          string `json:"cnpj"`
	Name        string `json:"razao_social"`
	TradingName string `json:"nome_fantasia"`
}
