package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gobwas/ws"
	log "github.com/jensneuse/abstractlogger"
	"github.com/jensneuse/graphql-go-tools/pkg/ast"
	"github.com/jensneuse/graphql-go-tools/pkg/execution"
	"github.com/jensneuse/graphql-go-tools/pkg/execution/datasource"
	gqlHTTP "github.com/jensneuse/graphql-go-tools/pkg/http"
	"github.com/jensneuse/graphql-go-tools/pkg/playground"
	"go.uber.org/zap"
)

var (
	schemaFile       string
	loggerConfigFile string
)

func logger() log.Logger {
	logger, _ := zap.NewProduction()
	return log.NewZapLogger(logger, log.DebugLevel)
}

func startServer(doc *ast.Document) {
	logger := logger()
	mux := http.NewServeMux()
	graphqlEndpoint := "/graphql"
	// base, err := datasource.NewBaseDataSourcePlanner(schemaData, datasource.PlannerConfiguration{}, abstractlogger.NoopLogger)
	base := &datasource.BasePlanner{
		Config:     datasource.PlannerConfiguration{},
		Log:        logger,
		Definition: doc,
	}

	err := base.RegisterDataSourcePlannerFactory("SchemaDataSource", datasource.SchemaDataSourcePlannerFactoryFactory{})
	if err != nil {
		log.Error(err)
	}

	handler := execution.NewHandler(base, nil)

	upgrader := &ws.DefaultHTTPUpgrader
	upgrader.Header = make(http.Header)
	upgrader.Header.Add("Sec-Websocket-Protocol", "graphql-ws")
	mux.HandleFunc("/time", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write(fakeResponse())
		if err != nil {
			log.Error(err)
		}
	})
	mux.Handle(graphqlEndpoint, gqlHTTP.NewGraphqlHTTPHandlerFunc(handler, logger, upgrader))
	playgroundURLPrefix := "/playground"
	playgroundURL := ""
	pg := playground.New(playground.Config{
		PathPrefix:                      playgroundURLPrefix,
		PlaygroundPath:                  playgroundURL,
		GraphqlEndpointPath:             graphqlEndpoint,
		GraphQLSubscriptionEndpointPath: graphqlEndpoint,
	})
	handlers, _ := pg.Handlers()
	for _, k := range handlers {
		mux.Handle(k.Path, k.Handler)
	}
	addr := "0.0.0.0:9111"
	log.String("add", addr)
	fmt.Printf("Access Playground on: http://%s%s%s\n", prettyAddr(addr), playgroundURLPrefix, playgroundURL)
	log.Error(http.ListenAndServe(addr, mux))
}

func fakeResponse() []byte {
	return []byte(`{"week_number":45,"utc_offset":"+01:00","utc_datetime":"2019-11-07T14:02:02.475928+00:00","unixtime":1573135322,"timezone":"Europe/Berlin","raw_offset":3600,"dst_until":null,"dst_offset":0,"dst_from":null,"dst":false,"day_of_year":311,"day_of_week":4,"datetime":"` + time.Now().String() + `","client_ip":"92.216.144.100","abbreviation":"CET"}`)
}

func prettyAddr(addr string) string {
	return strings.Replace(addr, "0.0.0.0", "localhost", -1)
}
