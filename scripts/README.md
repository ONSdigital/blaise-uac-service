# Yay some scripts...

Just some useful support scripts for running locally...

## Prerequisites to running scripts

Login to the GCP CLI:

```
gcloud auth login
```

Point the GCP CLI to the appropriate project:

```
gcloud config set project ons-blaise-v2-dev-<sandbox>
```

Use service account impersonation to auth:

```
gcloud auth application-default login --impersonate-service-account=ons-blaise-v2-dev-<sandbox>@appspot.gserviceaccount.com
```

## update_instrument_name

Updates instrument_name UACs have been generated for

Obviously be **VERY CAREFUL** if running in prod!

Set some local env vars:

Unix:
```
export PROJECT_ID=ons-blaise-v2-dev-<sandbox>
export OLD_INSTRUMENT_NAME=lms2212_rr1
export NEW_INSTRUMENT_NAME=lms2212_rr5
```

Windows:
```
set PROJECT_ID=ons-blaise-v2-dev-<sandbox>
set OLD_INSTRUMENT_NAME=lms2212_rr1
set NEW_INSTRUMENT_NAME=lms2212_rr5
```

Run da ting:
```
go run update_instrument_name.go
```

## disable_uacs

Disables specified UACs

Obviously be **VERY CAREFUL** if running in prod!

Set some local env vars:

Unix:
```
export PROJECT_ID=ons-blaise-v2-dev-<sandbox>
export UACS_TO_DISABLE=<uac>,<uac>,<uac>,<uac>,<uac>
```

Windows:
```
set PROJECT_ID=ons-blaise-v2-dev-<sandbox>
set UACS_TO_DISABLE=<uac>,<uac>,<uac>,<uac>,<uac>
```

Run da ting:
```
go run disable_uacs.go
```

## enable_uacs

Enables specified UACs

Obviously be **VERY CAREFUL** if running in prod!

Set some local env vars:

Unix:
```
export PROJECT_ID=ons-blaise-v2-dev-<sandbox>
export UACS_TO_ENABLE=<uac>,<uac>,<uac>
```

Windows:
```
set PROJECT_ID=ons-blaise-v2-dev-<sandbox>
set UACS_TO_ENABLE=<uac>,<uac>,<uac>
```

Run da ting:
```
go run enable_uacs.go
```
