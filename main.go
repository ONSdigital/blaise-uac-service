package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/webserver"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Serverpark       string `default:"gusty"`
	DatastoreProject string `required:"true" split_words:"true"`
	BlaiseBaseUrl    string `required:"true" split_words:"true"`
	Port             string `default:"8082"`
	UacKind          string `default:"uac" split_words:"true"`
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
		BlaiseRestApi: blaiseRestAPI,
		UacGenerator:  uacGenerator,
	}

	httpRouter := server.SetupRouter()
	err = httpRouter.Run(fmt.Sprintf(":%s", config.Port))
	if err != nil {
		log.Fatal(err.Error())
	}
}
