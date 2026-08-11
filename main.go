package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/webserver"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Serverpark       string `default:"gusty"`
	DatastoreProject string `required:"true" split_words:"true"`
	BlaiseBaseURL    string `required:"true" split_words:"true"`
	Port             string `default:"8082"`
	UACKind          string `required:"true" split_words:"true" envconfig:"UAC_KIND"`
}

func main() {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	datastoreClient, err := datastore.NewClient(ctx, config.DatastoreProject)
	if err != nil {
		log.Fatal(err)
	}

	blaiseRESTAPI := &blaiserestapi.BlaiseRESTAPI{
		Serverpark: config.Serverpark,
		BaseURL:    config.BlaiseBaseURL,
		Client:     &http.Client{Timeout: 3 * time.Minute},
	}
	uacService := uacgenerator.NewUACService(datastoreClient, config.UACKind)

	server := &webserver.Server{
		BlaiseRESTAPI: blaiseRESTAPI,
		UACService:    uacService,
	}

	httpRouter := server.SetupRouter()
	err = httpRouter.Run(fmt.Sprintf(":%s", config.Port))
	if err != nil {
		log.Fatal(err)
	}
}
