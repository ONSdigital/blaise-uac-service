# Blaise UAC Service

Blaise UAC service (BUS) is a service used to generate and validate UACs for blaise web collection. 

![bus](./bus.jpeg)

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