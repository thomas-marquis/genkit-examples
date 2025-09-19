package main

import (
	"context"
	"fmt"
	"genkit-examples/internal/agent"
	"genkit-examples/internal/vectorstore"
	"log"
	"net/http"
	"time"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/server"
	"github.com/spf13/viper"
	"github.com/thomas-marquis/genkit-mistral/mistral"
	"github.com/thomas-marquis/genkit-mistral/mistralclient"
)

func main() {
	viper.SetConfigFile("settings.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	mistralApiKey := viper.GetString("mistral.apiKey")

	ctx := context.Background()
	g := genkit.Init(ctx,
		genkit.WithPlugins(
			mistral.NewPlugin(mistralApiKey, mistral.WithClientConfig(
				mistralclient.Config{ClientTimeout: 40 * time.Second})),
		),
	)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		viper.GetString("db.user"),
		viper.GetString("db.password"),
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.name"),
	)
	v, err := vectorstore.New(connStr)
	if err != nil {
		panic(err)
	}

	a := agent.New(ctx, g, v)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /ask", genkit.Handler(a.ChatFlow()))
	if err := server.Start(ctx, "127.0.0.1:3400", mux); err != nil {
		log.Fatal(err)
	}
}
