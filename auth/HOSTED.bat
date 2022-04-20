
WHERE gcloud  >nul 2>nul
IF %ERRORLEVEL% NEQ 0 (
echo 'ERROR: This script requires the gcloud command. Please install it to continue.'
goto :eof
)

:: Calls to the hosted service are secure.
set APG_REGISTRY_INSECURE=

:: Get the service address.
set APG_REGISTRY_AUDIENCES=https://apigeeregistry.googleapis.com
set APG_REGISTRY_ADDRESS=apigeeregistry.googleapis.com:443

:: The auth token is generated for the gcloud logged-in user.
FOR /F %%g IN ('gcloud config list account --format "value(core.account)"') DO set APG_REGISTRY_CLIENT_EMAIL=%%g
FOR /F %%g IN ('gcloud auth print-access-token %APG_REGISTRY_CLIENT_EMAIL%') DO set APG_REGISTRY_TOKEN=%%g

:: Calls don't use an API key.
set APG_REGISTRY_API_KEY=
