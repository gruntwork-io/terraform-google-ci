# GCR Registry

This Terraform Module prepares a [GCR registry](https://cloud.google.com/container-registry/) and applies read and write
IAM roles to the bucket.

When granting a role for the individual storage bucket rather than for the entire GCP project, we have to ensure that we
have first pushed an image to Container Registry in the wanted host location so that the underlying storage bucket exists.
If input variable `init_registry` is `true`, the module pushes a minimal image to the registry to have the GCS bucket
created, so that we can apply the IAM bindings.

The module uses `gcloud` and `docker` commands to initialize the registry. The script does _not_ currently work on Windows.

## How do you use this module?

See the [root README](/README.md) for instructions on using modules.

## Core concepts

To understand core concepts like what is GCR Registry and how to work with one, see the [Container Registry documentation](https://cloud.google.com/container-registry/).
