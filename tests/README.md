# tests

This directory contains tests and demo walkthroughs of the Registry API.

# BenchMarks

Following commands help run benchmark tests against 100 apis with 3 revisions/api 

```
    export ITERATION_SIZE=100
    export REGISTRY_PROJECT_IDENTIFIER=bench
    export REGISTRY_VERSION_COUNT=3
    # Prefix for the name of the API
    export API_NAME_PREFIX=test
    # If you want the api names to start at 401 set the following variable to 400, defaults to 0
    export API_NAME_START_OFFSET=400
    apg admin create-project --project_id=$REGISTRY_PROJECT_IDENTIFIER
    go test --bench=Create ./tests/benchmark --benchtime=${ITERATION_SIZE}x --timeout=0
    go test --bench=Update ./tests/benchmark --benchtime=${ITERATION_SIZE}x  --timeout=0
    go test --bench=Get ./tests/benchmark --benchtime=${ITERATION_SIZE}x  --timeout=0
    go test --bench=List ./tests/benchmark --benchtime=${ITERATION_SIZE}x  --timeout=0
    go test --bench=Delete ./tests/benchmark --benchtime=${ITERATION_SIZE}x  --timeout=0
    apg admin delete-project --name=projects/$REGISTRY_PROJECT_IDENTIFIER --force
```