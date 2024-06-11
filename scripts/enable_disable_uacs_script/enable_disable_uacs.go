package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

type dsEntityStruct struct {
	InstrumentName string `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string `json:"case_id" datastore:"case_id"`
	Disabled       bool   `json:"disabled" datastore:"disabled"`
}

func main() {

	projectID := os.Getenv("PROJECT_ID")
	uacsToEnableItems := strings.Split(os.Getenv("UACS_TO_ENABLE"), ",")
	uacsToDisableItems := strings.Split(os.Getenv("UACS_TO_DISABLE"), ",")

	uacEnableUpdatedCount := 0
	uacDisableUpdatedCount := 0

	ctx := context.Background()
	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer func(dsClient *datastore.Client) {
		err := dsClient.Close()
		if err != nil {

		}
	}(dsClient)

	dsQuery := datastore.NewQuery("uac")

	dsIterator := dsClient.Run(ctx, dsQuery)

	for {
		var entity dsEntityStruct

		key, err := dsIterator.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for _, uac := range uacsToEnableItems {
			if key.Name == uac {

				dsEntityKey := datastore.NameKey("uac", key.Name, nil)
				entity.Disabled = false

				if _, err := dsClient.Put(ctx, dsEntityKey, &entity); err != nil {
					log.Fatal(err)
				}

				fmt.Println("Enabled UAC", key.Name)
				uacEnableUpdatedCount++
				break
			}
		}

		for _, uac := range uacsToDisableItems {
			if key.Name == uac {
				dsEntityKey := datastore.NameKey("uac", key.Name, nil)

				entity.Disabled = true

				if _, err := dsClient.Put(ctx, dsEntityKey, &entity); err != nil {
					log.Fatal(err)
				}

				fmt.Println("Disabled UAC", key.Name)
				uacDisableUpdatedCount++
				break
			}
		}

	}

	fmt.Println("Enabled", uacEnableUpdatedCount, "UACs")
	fmt.Println("Disabled", uacDisableUpdatedCount, "UACs")
}
