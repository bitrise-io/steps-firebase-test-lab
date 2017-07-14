## Secrets management

- `Generic File Storage` is not used.
- Secrets are set using `App Environment Variables` which are available in pull requests. 
  Only private repos with trusted committers should use the Firebase Test Lab step.
- Optionally `Secret Environment Variables` may be used however they will not be available
  on pull request builds.

## Environment Variables

Env | Description
--- | ---
GCLOUD_USER    | client_email from key.json
GCLOUD_PROJECT | project_id from key.json
GCLOUD_KEY     | key.json for a [service account](https://cloud.google.com/compute/docs/access/service-accounts)
APP_APK        | app apk to test
TEST_APK       | test apk containing tests to execute

## To Do

- Run `errcheck -asserts=true -blank=true .` automatically
