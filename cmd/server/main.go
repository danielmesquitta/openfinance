package main

import (
	"github.com/danielmesquitta/openfinance/internal/app/http"
)

// @title OpenFinance to Notion API
// @version 1.0
// @description This API is responsible for syncing OpenFinance data to Notion.
// @contact.name Daniel Mesquita
// @contact.email danielmesquitta123@gmail.com
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	http.Start()
}
