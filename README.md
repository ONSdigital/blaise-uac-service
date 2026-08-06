# Blaise UAC Service ![Bus](.github/bus.png)

Blaise UAC Service (BUS) generates and validates UACs (Unique Access Codes) for Blaise web collection, used by DQS and the BUS-UI.

The service returns existing UACs where available, or generates new ones on demand. Each UAC is linked to a specific questionnaire and case ID in Datastore, and is used by the CAWI Portal to direct respondents to their questionnaire.

## Endpoints

The API exposes UAC management endpoints under the `/uacs` route group, plus health-check endpoints.

### UAC Endpoints

- `UACInstrumentGenerateEndpoint`

```http
POST /uacs/instrument/:instrumentName
```

Generates UACs for an instrument's CAWI cases (fetched from Blaise), then returns all UACs for that instrument.

- `UACGetAllEndpoint`

```http
GET /uacs/instrument/:instrumentName
```

Returns all UACs for the specified instrument.

- `UACGetAllByCaseIDEndpoint`

```http
GET /uacs/instrument/:instrumentName/bycaseid
```

Returns all UACs for the specified instrument, ordered/grouped by case ID.

- `UACCountEndpoint`

```http
GET /uacs/instrument/:instrumentName/count
```

Returns the total number of UACs for the specified instrument.

- `UACGenerateEndpoint`

```http
POST /uacs/generate
```

Generates UACs from a request body containing `instrument_name` and `case_ids`, then returns all UACs for that instrument.

- `GetUacInfoEndpoint`

```http
POST /uacs/uac
```

Returns details for a single UAC from a request body containing `uac`.

- `AdminDeleteEndpoint`

```http
DELETE /uacs/admin/instrument/:instrumentName
```

Deletes all UAC data for the specified instrument.

- `ListInstrumentsEndpoint`

```http
GET /uacs/instruments
```

Returns the list of instruments that currently have UAC data.

- `ImportEndpoint`

```http
POST /uacs/import
```

Imports an array of UAC strings and returns the number imported.

- `UACDisableEndpoint`

```http
PATCH /uacs/uac/disable/:uac
```

Disables the specified UAC.

- `UACEnableEndpoint`

```http
PATCH /uacs/uac/enable/:uac
```

Enables the specified UAC.

- `UACGetAllDisabledEndpoint`

```http
GET /uacs/uac/:instrumentName/disabled
```

Returns all disabled UACs for the specified instrument.

### Health Endpoints

- `HealthEndpoint`

```http
GET /health
```

Returns the service health status.

- `HealthEndpoint`

```http
GET /bus/:version/health
```

Returns the service health status and echoes the provided `version` path parameter.

## Local Development

### Prerequisites

- [Go](https://go.dev/)
- [Google Cloud SDK (`gcloud` CLI)](https://cloud.google.com/sdk/)

### Clone and install dependencies

```sh
git clone https://github.com/ONSDigital/blaise-uac-service.git
cd blaise-uac-service
make install
```

### Authenticate with Google Cloud (keyless)

Use service account impersonation to auth:

```sh
gcloud auth login
gcloud config set project ons-blaise-v2-dev-<sandbox>
gcloud auth application-default login --impersonate-service-account=ons-blaise-v2-dev-<sandbox>@appspot.gserviceaccount.com
```

### Start an IAP tunnel to Blaise REST API

Run this in a separate terminal and keep it running:

```sh
gcloud compute start-iap-tunnel restapi-1 80 --local-host-port=localhost:8080 --zone europe-west2-a
```

Expected output includes `Listening on port [8080]`.

### Configure environment variables

Set environment variables using `export` (same style as the scripts in [scripts/README.md](scripts/README.md)):

```sh
export DATASTORE_PROJECT=ons-blaise-v2-dev-<sandbox>
export BLAISE_BASE_URL=http://localhost:8080
```

### Run the service

```sh
go run .
```

The service listens on `http://localhost:8082` by default.

Health check:

```sh
curl http://localhost:8082/health
```

Get generated UACs:

```sh
curl http://localhost:8082/uacs/instrument/<questionnaire name>
```

## Running the tests 

```sh
go test ./...
```
