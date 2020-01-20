# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# CREATE IAM PERMISSIONS FOR GCR REGISTRY BUCKET
# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE TERRAFORM AND REMOTE STATE STORAGE
# ---------------------------------------------------------------------------------------------------------------------

terraform {
  # The configuration for this backend will be filled in by Terragrunt
  backend "gcs" {}

  # Use Terraform 0.12.x so that we can take advantage of the new language features and GCP functionality as a
  # separate provider via https://github.com/terraform-providers/terraform-provider-google.
  required_version = ">= 0.12.0"
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE THE GCP PROVIDERS
# ---------------------------------------------------------------------------------------------------------------------

provider "google" {
  version = "~> 3.4"
  project = var.project
}

provider "google-beta" {
  version = "~> 3.4"
  project = var.project
}

# ------------------------------------------------------------------------------
# INITIALIZE REGISTRY
#
# The following script pushes a minimal image to the registry to have the GCS bucket created, so we can apply the IAM bindings.
# See: https://cloud.google.com/container-registry/docs/access-control#permissions_and_roles
# Note that the script does _not_ work in Windows.
# ------------------------------------------------------------------------------

resource "null_resource" "init_registry" {
  count = var.init_registry ? 1 : 0

  provisioner "local-exec" {
    command = <<EOF
   gcloud components install docker-credential-gcr --quiet && \
   docker-credential-gcr configure-docker && \
   (echo 'FROM scratch'; echo 'LABEL maintainer=gruntwork.io') | \
   docker build -t ${local.registry_url}/scratch:latest - && \
   docker push ${local.registry_url}/scratch:latest
EOF
  }
}

# ------------------------------------------------------------------------------
# CREATE BUCKET IAM BINDINGS
# ------------------------------------------------------------------------------

resource "google_storage_bucket_iam_binding" "read_binding" {
  count  = length(var.readers) > 0 ? 1 : 0
  bucket = local.bucket_name
  role   = "roles/storage.objectViewer"

  members = var.readers

  depends_on = [null_resource.init_registry]
}

resource "google_storage_bucket_iam_binding" "write_binding" {
  count  = length(var.writers) > 0 ? 1 : 0
  bucket = local.bucket_name
  role   = "roles/storage.admin"

  members = var.writers

  depends_on = [null_resource.init_registry]
}

# ------------------------------------------------------------------------------
# PREPARE LOCALS
# ------------------------------------------------------------------------------

locals {
  // Repository URL will start with [us|eu|asia].gcr.io which will be used in the bucket name
  bucket_name  = "${var.registry_region}.artifacts.${var.project}.appspot.com"
  registry_url = "${var.registry_region}.gcr.io/${var.project}"
}
