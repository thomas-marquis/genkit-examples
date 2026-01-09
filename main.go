package main

import (
	"context"
	"fmt"
	"genkit-examples/internal/agent/chat"
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
			mistral.NewPlugin(mistralApiKey,
				mistral.WithClientOptions(mistralclient.WithClientTimeout(40*time.Second)),
			),
		))

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		viper.GetString("db.user"),
		viper.GetString("db.password"),
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.name"),
	)

	bookRepo, err := infrastructure.NewBookRepositoryImpl(10, connStr)
	if err != nil {
		panic(err)
	}

	notionKey := viper.GetString("notion.secretKey")

	c, err := chat.New(ctx, g, bookRepo,
		chat.WithNotionApiKey(notionKey),
		chat.WithLLM("mistral/mistral-medium-latest"),
	)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat", c.ToHandler())
	if err := server.Start(ctx, "127.0.0.1:3400", mux); err != nil {
		log.Fatal(err)
	}
}
