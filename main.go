package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/webserver"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Config struct {
	Serverpark         string `default:"gusty"`
	DatastoreProject   string `required:"true" split_words:"true"`
	BlaiseBaseUrl      string `required:"true" split_words:"true"`
	Port               string `default:"8082"`
	UacKind            string `default:"uac" split_words:"true"`
	GoogleCloudProject string `split_words:"true"`
	GAEService         string `default:"bus" split_words:"true"`
}

func NewTracerProvider(config *Config) *sdktrace.TracerProvider {
	exporter, err := texporter.New(texporter.WithProjectID(config.GoogleCloudProject))
	if err != nil {
		log.Fatal("Could not create google trace exporter")
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(config.GAEService))),
	)
}

func NewTracerMiddleware(config *Config, traceProvider *sdktrace.TracerProvider) gin.HandlerFunc {
	return otelgin.Middleware(config.GAEService, otelgin.WithTracerProvider(traceProvider))
}

func main() {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx := context.Background()
	datastoreClient, err := datastore.NewClient(ctx, config.DatastoreProject)
	if err != nil {
		log.Fatal(err.Error())
	}

	blaiseRestAPI := &blaiserestapi.BlaiseRestApi{
		Serverpark: config.Serverpark,
		BaseUrl:    config.BlaiseBaseUrl,
		Client:     &http.Client{},
	}
	uacGenerator := uacgenerator.NewUacGenerator(datastoreClient, config.UacKind)

	server := &webserver.Server{
		BlaiseRestApi:    blaiseRestAPI,
		UacGenerator:     uacGenerator,
		TracerMiddleWare: NewTracerMiddleware(&config, NewTracerProvider(&config)),
	}

	httpRouter := server.SetupRouter()
	httpRouter.Run(fmt.Sprintf(":%s", config.Port))
}
