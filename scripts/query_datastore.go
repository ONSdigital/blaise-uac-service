package main

import (
		"strings"
		"time"
		"os"
        "context"
        "log"
		"fmt"
        "cloud.google.com/go/datastore"
		"google.golang.org/api/iterator"
)

type dsEntityStruct struct {
	AppointmentInfo string `json:"appointment_info" datastore:"appointment_info"`
	BusyDials int `json:"busy_dials" datastore:"busy_dials"`
	CallEndTime time.Time `json:"call_end_time" datastore:"call_end_time"`
	CallNumber int `json:"call_number" datastore:"call_number"`
	CallResult string `json:"call_result" datastore:"call_result"`
	CallStartTime time.Time `json:"call_start_time" datastore:"call_start_time"`
	Cohort string `json:"cohort" datastore:"cohort"`
	InstrumentName string `json:"questionnaire_name" datastore:"questionnaire_name"`
	CaseID string `json:"serial_number" datastore:"serial_number"`
	DialNumber int `json:"dial_number" datastore:"dial_number"`
	DialSecs int `json:"dial_secs" datastore:"dial_secs"`
	Interviewer string `json:"interviewer" datastore:"interviewer"`
	NumberOfInterviews int `json:"number_of_interviews" datastore:"number_of_interviews"`
	OutcomeCode string `json:"outcome_code" datastore:"outcome_code"`
	Status string `json:"status" datastore:"status"`
	Survey string `json:"survey" datastore:"survey"`
	UpdateInfo string `json:"update_info" datastore:"update_info"`
	Wave int `json:"wave" datastore:"wave"`
}

func main() {

	projectID := os.Getenv("PROJECT_ID")
	instrumentNames := []string{}

	ctx := context.Background()	

	dsClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer dsClient.Close()

	for _, instrumentName := range instrumentNames {
	callHistoryCount := 0
	callHistoryToDeleteCount := 0
	dsQuery := datastore.NewQuery("CallHistory").FilterField("questionnaire_name", "=", instrumentName)

	dsIt := dsClient.Run(ctx, dsQuery)
	dsItCount := *dsIt

	for {
		var entity dsEntityStruct
		_, err := dsItCount.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		callHistoryCount++
	}

	fmt.Println("Found", callHistoryCount, "datastore entries for instrument name", instrumentName)

	for {

		var entity dsEntityStruct

		key, err := dsIt.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		dsEntityKey := datastore.NameKey("CallHistory", key.Name, nil)

		if strings.Contains(dsEntityKey.Name, instrumentName) {
			entityName := dsEntityKey.Name
			instrumentIndex := strings.Index(entityName, instrumentName)
			if instrumentIndex != -1 {
				entityName = entityName[:instrumentIndex] + entityName[instrumentIndex+len(instrumentName)+1:]
			}

			dsEntityKeyToDelete := datastore.NameKey("CallHistory", entityName, nil)
			fmt.Println(dsEntityKeyToDelete)

			// Check if the new entity name exists as a key
			err := dsClient.Get(ctx, dsEntityKeyToDelete, &entity)
			if err == nil {
				fmt.Println("Deleting entry", dsEntityKeyToDelete.Name, "from instrument name", instrumentName)
				callHistoryToDeleteCount++
				if err := dsClient.Delete(ctx, dsEntityKeyToDelete); err != nil {
					log.Fatal(err)
				} else {
					fmt.Println("Deleted entry", dsEntityKeyToDelete.Name, "from instrument name", instrumentName)
				}
			} else{
				fmt.Println("Entity", dsEntityKeyToDelete.Name, "not found in datastore")
			}
		}
		}
		fmt.Println("Deleted", callHistoryToDeleteCount, "datastore entries for instrument name", instrumentName)	
	}
}

