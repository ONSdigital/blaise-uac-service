# Blaise UAC Service

Blaise UAC service (BUS) is a used to generate and validate UACs for blaise web.

![bus](./bus.jpeg)

## Running the tests

```sh
go test ./...
```

TODO:
- get all uacs for instrument "GET" endpoint
- get errors out of uac generation goroutines
- get caseId and instrument name from UAC
