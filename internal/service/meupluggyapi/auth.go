package meupluggyapi

import (
	"encoding/json"
	"net/http"
)

type AuthenticateResponse struct {
	APIKey string `json:"apiKey"`
}

func (m *MeuPluggyAPIClient) Authenticate() error {
	url := m.BaseURL.String() + "/auth"

	req, _ := http.NewRequest("POST", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)

	data := AuthenticateResponse{}
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	m.Token = data.APIKey

	return nil
}
