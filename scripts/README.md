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

Get a service account JSON key:
```
gcloud iam service-accounts keys create keys.json --iam-account ons-blaise-v2-dev-<sandbox>@appspot.gserviceaccount.com
```


## update_instrument_name

Updates all instances of an instrument_name! #ronseal

Obviously be **VERY CAREFUL** if running in prod!

Set some local env vars:

Unix:
```
export GOOGLE_APPLICATION_CREDENTIALS=keys.json
export PROJECT_ID=ons-blaise-v2-dev-<sandbox>
export OLD_INSTRUMENT_NAME=lms2212_rr1
export NEW_INSTRUMENT_NAME=lms2212_rr5
```

Windows:
```
set GOOGLE_APPLICATION_CREDENTIALS=keys.json
set PROJECT_ID=ons-blaise-v2-dev-<sandbox>
set OLD_INSTRUMENT_NAME=lms2212_rr1
set NEW_INSTRUMENT_NAME=lms2212_rr5
```

Run da ting:
```
go run update_instrument_name.go
```


## disable_uacs

Updates all cases with specific uac's to be disabled

Obviously be **VERY CAREFUL** if running in prod!

Set some local env vars:

Unix:
```
export GOOGLE_APPLICATION_CREDENTIALS=keys.json
export PROJECT_ID=ons-blaise-v2-dev-<sandbox>
export UACS_TO_DISABLE=<uac>,<uac>
```

Windows:
```
set GOOGLE_APPLICATION_CREDENTIALS=keys.json
set PROJECT_ID=ons-blaise-v2-dev-<sandbox>
set UACS_TO_DISABLE=<uac>,<uac>
```

Run da ting:
```
go run disable_uacs.go
```

## enable_uacs

Updates all cases with specific uac's to be enabled

Obviously be **VERY CAREFUL** if running in prod!

Set some local env vars:

Unix:
```
export GOOGLE_APPLICATION_CREDENTIALS=keys.json
export PROJECT_ID=ons-blaise-v2-dev-<sandbox>
export UACS_TO_ENABLE=<uac>,<uac>,<uac>
```

Windows:
```
set GOOGLE_APPLICATION_CREDENTIALS=keys.json
set PROJECT_ID=ons-blaise-v2-dev-<sandbox>
set UACS_TO_ENABLE=<uac>,<uac>,<uac>
```

Run da ting:
```
go run enable_uacs.go
```