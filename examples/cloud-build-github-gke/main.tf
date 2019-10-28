# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY A CI/CD PIPELINE USING CLOUD BUILD, GKE AND GITHUB IN GOOGLE CLOUD PLATFORM
# This is an example of how to setup an automated CI/CD pipeline in GCP using Cloud Build and a GKE cluster that is
# triggered from GitHub.
# ---------------------------------------------------------------------------------------------------------------------

terraform {
  # The modules used in this example have been updated with 0.12 syntax, additionally we depend on a bug fixed in
  # version 0.12.7.
  required_version = ">= 0.12.7"
}

# ---------------------------------------------------------------------------------------------------------------------
# PREPARE PROVIDERS
# ---------------------------------------------------------------------------------------------------------------------

provider "google" {
  version = "~> 2.10"
  project = var.project
  region  = var.region
}

provider "google-beta" {
  version = "~> 2.10"
  project = var.project
  region  = var.region
}

# ---------------------------------------------------------------------------------------------------------------------
# CREATE A CLOUD BUILD TRIGGER
# ---------------------------------------------------------------------------------------------------------------------

resource "google_cloudbuild_trigger" "cloud_build_trigger" {
  provider    = "google-beta"
  description = "GitHub Repository Trigger ${var.github_owner}/${var.github_repository} (${var.branch_name})"

  github {
    owner = var.github_owner
    name  = var.github_repository
    push {
      branch = var.branch_name
    }
  }

  # These substitutions have been defined in the sample app's cloudbuild.yaml file.
  # See: https://github.com/gruntwork-io/sample-app-docker/blob/master/cloudbuild.yaml#L40
  substitutions = {
    _GCR_REGION           = var.gcr_region
    _GKE_CLUSTER_LOCATION = var.location
    _GKE_CLUSTER_NAME     = var.cluster_name
  }

  # The filename argument instructs Cloud Build to look for a file in the root of the repository.
  # Either a filename or build template (below) must be provided.
  filename = "cloudbuild.yaml"

  # Remove the filename and substitutions arguments above and uncomment the code below if you'd prefer to define your
  # build steps in Terraform code.
  # build {
  #   # install the app dependencies
  #   step {
  #     name = "gcr.io/cloud-builders/npm"
  #     args = ["install"]
  #   }
  #
  #   # execute the tests
  #   step {
  #     name = "gcr.io/cloud-builders/npm"
  #     args = ["run", "test"]
  #   }
  #
  #   # build an artifact using the docker builder
  #   step {
  #     name = "gcr.io/cloud-builders/docker"
  #     args = ["build", "--build-arg", "NODE_ENV=production", "-t", "gcr.io/$PROJECT_ID/$REPO_NAME:$SHORT_SHA", "."]
  #   }
  #
  #   # push the artifact to a GCR repository
  #   step {
  #     name = "gcr.io/cloud-builders/docker"
  #     args = ["push", "${var.gcr_region}.gcr.io/$PROJECT_ID/$REPO_NAME:$SHORT_SHA"]
  #   }
  #
  #   # deploy the app to a GKE cluster using the `gke-deploy` builder and expose it
  #   # using a load balancer on port 80.
  #   # https://github.com/GoogleCloudPlatform/cloud-builders/tree/master/gke-deploy
  #   step {
  #     name = "gcr.io/cloud-builders/gke-deploy"
  #     args = ["run", "--image=${var.gcr_region}.gcr.io/$PROJECT_ID/$REPO_NAME:$SHORT_SHA", "--location", "${var.location}", "--cluster", "${var.cluster_name}", "--expose", "80"]
  #   }
  # }
}

# ---------------------------------------------------------------------------------------------------------------------
# CREATE A CUSTOM SERVICE ACCOUNT TO USE WITH THE GKE CLUSTER
# ---------------------------------------------------------------------------------------------------------------------

module "gke_service_account" {
  source = "git::git@github.com:gruntwork-io/terraform-google-gke.git//modules/gke-service-account?ref=v0.3.8"

  name        = var.cluster_service_account_name
  project     = var.project
  description = var.cluster_service_account_description
}

# ---------------------------------------------------------------------------------------------------------------------
# CONFIGURE THE GCR REGISTRY TO STORE THE CLOUD BUILD ARTIFACTS
# ---------------------------------------------------------------------------------------------------------------------

module "gcr_registry" {
  # When using these modules in your own templates, you will need to use a Git URL with a ref attribute that pins you
  # to a specific version of the modules, such as the following example:
  # source = "github.com/gruntwork-io/terraform-google-ci.git//modules/gcr-registry?ref=v0.1.0"
  source = "../../modules/gcr-registry"

  project         = var.project
  registry_region = var.gcr_region

  # allow the custom service account to pull images from the GCR registry
  readers = ["serviceAccount:${module.gke_service_account.email}"]
}

# ---------------------------------------------------------------------------------------------------------------------
# DEPLOY A PRIVATE GKE CLUSTER IN GOOGLE CLOUD PLATFORM
# ---------------------------------------------------------------------------------------------------------------------

module "gke_cluster" {
  # Use a version of the gke-cluster module that supports Terraform 0.12
  source = "git::git@github.com:gruntwork-io/terraform-google-gke.git//modules/gke-cluster?ref=v0.3.8"

  name = var.cluster_name

  project  = var.project
  location = var.location
  network  = module.vpc_network.network

  # We're deploying the cluster in the 'public' subnetwork to allow outbound internet access
  # See the network access tier table for full details:
  # https://github.com/gruntwork-io/terraform-google-network/tree/master/modules/vpc-network#access-tier
  subnetwork = module.vpc_network.public_subnetwork

  # When creating a private cluster, the 'master_ipv4_cidr_block' has to be defined and the size must be /28
  master_ipv4_cidr_block = var.master_ipv4_cidr_block

  # This setting will make the cluster private
  enable_private_nodes = "true"

  # To make testing easier, we keep the public endpoint available. In production, we highly recommend restricting access to only within the network boundary, requiring your users to use a bastion host or VPN.
  disable_public_endpoint = "false"

  # With a private cluster, it is highly recommended to restrict access to the cluster master
  # However, for testing purposes we will allow all inbound traffic.
  master_authorized_networks_config = [
    {
      cidr_blocks = [
        {
          cidr_block   = "0.0.0.0/0"
          display_name = "all-for-testing"
        },
      ]
    },
  ]

  cluster_secondary_range_name = module.vpc_network.public_subnetwork_secondary_range_name
}

# ---------------------------------------------------------------------------------------------------------------------
# CREATE A NODE POOL
# ---------------------------------------------------------------------------------------------------------------------

resource "google_container_node_pool" "node_pool" {
  provider = google-beta

  name     = "private-pool"
  project  = var.project
  location = var.location
  cluster  = module.gke_cluster.name

  initial_node_count = "1"

  autoscaling {
    min_node_count = "1"
    max_node_count = "3"
  }

  management {
    auto_repair  = "true"
    auto_upgrade = "true"
  }

  node_config {
    image_type   = "COS"
    machine_type = "n1-standard-1"

    labels = {
      private-pools-example = "true"
    }

    # Add a private tag to the instances. See the network access tier table for full details:
    # https://github.com/gruntwork-io/terraform-google-network/tree/master/modules/vpc-network#access-tier
    tags = [
      module.vpc_network.private,
      "private-pool-example",
    ]

    disk_size_gb = "30"
    disk_type    = "pd-standard"
    preemptible  = false

    service_account = module.gke_service_account.email

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  lifecycle {
    ignore_changes = [initial_node_count]
  }

  timeouts {
    create = "30m"
    update = "30m"
    delete = "30m"
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# CREATE A NETWORK TO DEPLOY THE CLUSTER TO
# ---------------------------------------------------------------------------------------------------------------------

module "vpc_network" {
  source = "github.com/gruntwork-io/terraform-google-network.git//modules/vpc-network?ref=v0.2.1"

  name_prefix = "${var.cluster_name}-network-${random_string.suffix.result}"
  project     = var.project
  region      = var.region

  cidr_block           = var.vpc_cidr_block
  secondary_cidr_block = var.vpc_secondary_cidr_block
}

# Use a random suffix to prevent overlap in network names
resource "random_string" "suffix" {
  length  = 4
  special = false
  upper   = false
}
