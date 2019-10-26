# Cloud Build and Cloud Source Repositories Example

The following example shows how you can setup an automated deployment pipeline using [Google Cloud Build](https://cloud.google.com/cloud-build/),
[Cloud Source Repositories](https://cloud.google.com/source-repositories) and a [Google Kubernetes Engine (GKE) cluster](https://cloud.google.com/kubernetes-engine/).

## What is Google Cloud Build?

Cloud Build lets you build software quickly across all languages. Get complete control over defining custom workflows
for building, testing, and deploying across multiple environments such as VMs, serverless, Kubernetes, or Firebase.
You can find out more on the [Cloud Build](https://cloud.google.com/cloud-build/) website.

## What is a Google Cloud Source Repository?

A Google Cloud Source Repository is a fully featured, private [Git](https://git-scm.com/) repository hosted on Google
Cloud Platform. These repositories let you develop and deploy an app or service in a space that provides collaboration
and version control for your code. You can find out more on the [Cloud Source Repositories documentation](https://cloud.google.com/source-repositories/docs/).

## Overview

1. Next configure `kubectl` to use the GKE cluster

```
$ gcloud container clusters get-credentials example-private-cluster --region europe-west3
```

```
$ kubectl expose deployment test-repo2 --type=LoadBalancer --port 80 --target-port 8080
```

```
$ kubectl get services -w
```

Once you’re done testing, you may want to clean up all the infrastructure you’ve deployed so GCP doesn’t charge you money for it. First, delete the
Kubernetes Service:

```bash
$ kubectl delete service test-repo2
```

### Manually Submitting Builds

```bash
$ gcloud builds submit --config=cloudbuild.yaml \
    --substitutions=TAG_NAME="test"
```

##

Cloud Build executes your builds using a service account, a special Google account that executes builds on your behalf. The email for the Cloud Build service
account is [PROJECT_NUMBER]@cloudbuild.gserviceaccount.com. When you enable the Cloud Build API, the service account is automatically created and granted the
Cloud Build Service Account role for your project. This role is sufficient for several tasks, including fetching code from Cloud Source Repositories, pushing and
pulling Docker images to Container Registry, however it does not allow Cloud Build to deploy to Kubernetes Engine clusters. Therefore you need to manually enable
your service account to perform these actions by granting the account additional IAM roles. For more information, refer to: https://cloud.google.com/cloud-build/docs/securing-builds/set-service-account-permissions.
