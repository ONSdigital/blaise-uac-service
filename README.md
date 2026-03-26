# Blaise UAC Service ![Bus](.github/bus.png)

Blaise UAC Service (BUS) is a service used to generate and validate UACs (Unique Access Codes) for Blaise web collection. It is utilised by DQS and/or the BUS-UI.

The service will provide existing UACs if they have already been generated or create new ones if more are requested than currently exist. Each UAC is tied to a specific questionnaire and case ID stored in Datastore. These UACs are used by the CAWI portal to direct respondents to their respective questionnaires.

## Running the tests 

```sh
go test ./...
```

# Endpoints

Endpoints have been added to the BUS service to allow for the interaction of UACs.

- UACEnableEndpoint - This endpoint is used to enable the UAC. It is a GET request that takes a "uac" parameter.

```
"/uac/enable/:uac"
```

- UACDisableEndpoint - This endpoint is used to disable the UAC. It is a GET request that takes a "uac" parameter.

``` 
"/uac/disable/:uac"
```

- UACGetAllDisabledEndpoint - This endpoint is used to get a list of all disabled UACs for an instrument. It is a GET 
request that takes an "instrument_name" parameter.

```
"/uac/:instrumentName/disabled"
```