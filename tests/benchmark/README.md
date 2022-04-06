# BenchMark tests

Following commands help run benchmark tests against 100 apis with 3 versions/api.
Every version has one api spec.

```
    export ITERATION_SIZE=10
    export REGISTRY_PROJECT_IDENTIFIER=bench
    export API_VERSIONS_COUNT=3
    apg admin create-project --project_id=$REGISTRY_PROJECT_IDENTIFIER

    go test --bench=Create ./tests/benchmark \
        --registry_project=$REGISTRY_PROJECT_IDENTIFIER \
        --version_count=$API_VERSIONS_COUNT \
        --benchtime=${ITERATION_SIZE}x --timeout=0

    go test --bench=Update ./tests/benchmark \
        --registry_project=$REGISTRY_PROJECT_IDENTIFIER \
        --version_count=$API_VERSIONS_COUNT \
        --benchtime=${ITERATION_SIZE}x  --timeout=0

    go test --bench=Get ./tests/benchmark \
        --registry_project=$REGISTRY_PROJECT_IDENTIFIER \
        --version_count=$API_VERSIONS_COUNT \
        --benchtime=${ITERATION_SIZE}x  --timeout=0

    go test --bench=List ./tests/benchmark \
        --registry_project=$REGISTRY_PROJECT_IDENTIFIER \
        --version_count=$API_VERSIONS_COUNT \
        --benchtime=${ITERATION_SIZE}x  --timeout=0

    go test --bench=Delete ./tests/benchmark \
        --registry_project=$REGISTRY_PROJECT_IDENTIFIER \
        --version_count=$API_VERSIONS_COUNT \
        --benchtime=${ITERATION_SIZE}x  --timeout=0

    apg admin delete-project --name=projects/$REGISTRY_PROJECT_IDENTIFIER --force
```