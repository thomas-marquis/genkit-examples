package main

import (
	"context"
	"genkit-examples/internal/agent"
	"genkit-examples/internal/infrastructure"
	"log"
	"net/http"
	"time"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/server"
	"github.com/spf13/viper"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	mistralclient "github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	viper.SetConfigFile("settings.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	mistralApiKey := viper.GetString("mistral.apiKey")

	ctx := context.Background()
	g := genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(mistralApiKey, mistral.WithClientOptions(
				mistralclient.WithClientTimeout(45*time.Second)),
			),
		))

	glRepo := infrastructure.NewGroceryListRepositoryImpl()

	menuAgent := agent.NewMenuAgent(g, glRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat", genkit.Handler(menuAgent.Flow()))
	if err := server.Start(ctx, "127.0.0.1:3400", mux); err != nil {
		log.Fatal(err)
	}
}
