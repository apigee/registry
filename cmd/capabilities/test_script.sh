
source auth/GKE.sh

# Create project
curl $APG_REGISTRY_AUDIENCES/v1/projects?project_id=shrutiparab-demo -i \
      -H "Authorization: Bearer $APG_REGISTRY_TOKEN" \
      -H "content-type: application/json" -X POST \
      -d '{"name": "shrutiparab-demo", "display_name": "shrutiparab-demo", "description": "Demo project for shrutiparab@"}'

# Delete project
curl $APG_REGISTRY_AUDIENCES/v1/projects/shrutiparab-demo -i \
      -H "Authorization: Bearer $APG_REGISTRY_TOKEN" \
      -H "content-type: application/json" -X DELETE

