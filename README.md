## Connecting a Repository

### Cloud Source Repository

#### Pushing Code

To be able to push code to your Cloud Source Repository, you first need to register your public SSH key with Google Cloud.

https://source.cloud.google.com/user/ssh_keys?register=true

You also have the option of using the Google Cloud SDK

```bash
$ gcloud init && git config --global credential.https://source.developers.google.com.helper gcloud.sh
```

### GitHub

## Triggering Builds Manually

```bash
$ gcloud builds submit . --config=cloudbuild.yaml
```

## Builds

Builds have a 10 minute timeout.

## Enalbing

Granting Cloud Build permission to access the GKE cluster.

gcloud projects add-iam-policy-binding dev-sandbox-228703 --member=serviceAccount:400198790674@cloudbuild.gserviceaccount.com --role=roles/container.developer
