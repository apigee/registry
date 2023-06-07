::
:: Copyright 2022 Google LLC.
::
:: Licensed under the Apache License, Version 2.0 (the "License");
:: you may not use this file except in compliance with the License.
:: You may obtain a copy of the License at
::
::    http://www.apache.org/licenses/LICENSE-2.0
::
:: Unless required by applicable law or agreed to in writing, software
:: distributed under the License is distributed on an "AS IS" BASIS,
:: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
:: See the License for the specific language governing permissions and
:: limitations under the License.
::


::
:: Configure an environment to run Apigee Registry clients with a Google-hosted service.
::
:: The following assumes you have run `gcloud auth login` and that the current
:: gcloud project is the one with your Apigee Registry instance.
::
@echo off

WHERE gcloud  >nul 2>nul
IF %ERRORLEVEL% NEQ 0 (
echo 'ERROR: This script requires the gcloud command. Please install it to continue.'
goto :eof
)

setlocal

:: set the service address.
set REGISTRY_ADDRESS=apigeeregistry.googleapis.com:443
set REGISTRY_INSECURE=false

FOR /F %%g IN ('gcloud config get project') DO set REGISTRY_PROJECT=%%g
set REGISTRY_LOCATION=global
FOR /F %%g IN ('gcloud config get account') DO set CLIENT_EMAIL=%%g
set REGISTRY_TOKEN_SOURCE=gcloud auth print-access-token %CLIENT_EMAIL%

registry config configurations create hosted ^
  --registry.insecure=%REGISTRY_INSECURE% ^
  --registry.address=%REGISTRY_ADDRESS% ^
  --registry.project=%REGISTRY_PROJECT% ^
  --registry.location=%REGISTRY_LOCATION%

registry config set token-source "%REGISTRY_TOKEN_SOURCE%"

endlocal
