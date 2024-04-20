package meupluggyapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type ListTransactionsResponse struct {
	Total      int64    `json:"total"`
	TotalPages int64    `json:"totalPages"`
	Page       int64    `json:"page"`
	Results    []Result `json:"results"`
}

type Result struct {
	ID          string      `json:"id"`
	Description Description `json:"description"`
	Amount      float64     `json:"amount"`
	Date        time.Time   `json:"date"`
	Category    Category    `json:"category"`
	PaymentData PaymentData `json:"paymentData"`
	Type        ResultType  `json:"type"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type PaymentData struct {
	Payer         *Payer         `json:"payer"`
	PaymentMethod *PaymentMethod `json:"paymentMethod"`
	Receiver      *Payer         `json:"receiver"`
}

type Payer struct {
	Name *string `json:"name"`
}

type Category string

const (
	Clothing               Category = "Clothing"
	Electricity            Category = "Electricity"
	Electronics            Category = "Electronics"
	FoodAndDrinks          Category = "Food and drinks"
	GymsAndFitnessCenters  Category = "Gyms and fitness centers"
	HospitalClinicsAndLabs Category = "Hospital clinics and labs"
	Houseware              Category = "Houseware"
	Housing                Category = "Housing"
	Investments            Category = "Investments"
	LandmarksAndMuseums    Category = "Landmarks and museums"
	Pharmacy               Category = "Pharmacy"
	Services               Category = "Services"
	Telecommunications     Category = "Telecommunications"
	Transfers              Category = "Transfers"
	University             Category = "University"
)

type Description string

const (
	AplicaçãoRDB          Description = "Aplicação RDB"
	PagamentoDeFatura     Description = "Pagamento de fatura"
	PagamentoEfetuado     Description = "Pagamento efetuado"
	ResgateRDB            Description = "Resgate RDB"
	TransferênciaEnviada  Description = "Transferência enviada"
	TransferênciaRecebida Description = "Transferência Recebida"
)

type PaymentMethod string

const (
	Boleto PaymentMethod = "BOLETO"
	Pix    PaymentMethod = "PIX"
	Ted    PaymentMethod = "TED"
)

type ResultType string

const (
	Credit ResultType = "CREDIT"
	Debit  ResultType = "DEBIT"
)

func (m *MeuPluggyAPIClient) ListTransactions(
	accountID string,
	from, to *time.Time,
) (*ListTransactionsResponse, error) {
	url := m.BaseURL

	url.Path = "/transactions"
	query := url.Query()

	query.Add("accountId", accountID)
	query.Add("limit", "500")

	if from != nil {
		url.Query().Add("from", time.DateOnly)
	}

	if to != nil {
		url.Query().Add("to", time.DateOnly)
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", m.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	data := &ListTransactionsResponse{}
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
