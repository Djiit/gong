package internal

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

var (
	AsanaRestApiUrl = "https://app.asana.com/api/1.0"
)

func ApiClient() *resty.Client {
	client := resty.New()
	client.SetBaseURL(AsanaRestApiUrl)
	client.SetAuthToken(viper.GetString("api-key"))
	return client
}
