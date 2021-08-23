[![Maintained by Gruntwork.io](https://img.shields.io/badge/maintained%20by-gruntwork.io-%235849a6.svg)](https://gruntwork.io/?ref=repo_google_ci)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gruntwork-io/terraform-google-ci.svg?label=latest)](https://github.com/gruntwork-io/terraform-google-ci/releases/latest)
![Terraform Version](https://img.shields.io/badge/tf-%3E%3D1.0.x-blue.svg)

# Google CI/CD Modules

This repo contains modules and examples to deploy CI/CD pipelines on Google Cloud Platform using [Google Cloud Build](https://cloud.google.com/cloud-build/).

## Quickstart

If you want to quickly deploy an automated CI/CD pipeline using Google Cloud Build and a Google Kubernetes Engine (GKE) cluster that is triggered from a Google Cloud Source Repository, check out the [cloud-build-csr-gke example documentation](https://github.com/gruntwork-io/terraform-google-ci/tree/master/examples/cloud-build-csr-gke)
for instructions.

## What's in this repo

This repo has the following folder structure:

- [root](https://github.com/gruntwork-io/terraform-google-ci/tree/master): The root folder contains various documentation and licenses.

- [modules](https://github.com/gruntwork-io/terraform-google-ci/tree/master/modules): This folder contains the
  main implementation code for this Module, broken down into multiple standalone submodules.

  The primary module is:

  - [gcr-registry](https://github.com/gruntwork-io/terraform-google-ci/tree/master/modules/gcr-registry): The GCR Registry module is used to
    setup and configure [Google Container Registry](https://cloud.google.com/container-registry/).

- [examples](https://github.com/gruntwork-io/terraform-google-ci/tree/master/examples): This folder contains
  examples of how to use the submodules.

- [test](https://github.com/gruntwork-io/terraform-google-ci/tree/master/test): Automated tests for the submodules
  and examples.

## What is Google Cloud Build?

Cloud Build lets you build software quickly across all languages. Get complete control over defining custom workflows
for building, testing, and deploying across multiple environments such as VMs, serverless, Kubernetes, or Firebase.
You can find out more on the [Cloud Build](https://cloud.google.com/cloud-build/) website.

## What's a Module?

A Module is a canonical, reusable, best-practices definition for how to run a single piece of infrastructure, such
as a database or server cluster. Each Module is written using a combination of [Terraform](https://www.terraform.io/)
and scripts (mostly bash) and include automated tests, documentation, and examples. It is maintained both by the open
source community and companies that provide commercial support.

Instead of figuring out the details of how to run a piece of infrastructure from scratch, you can reuse
existing code that has been proven in production. And instead of maintaining all that infrastructure code yourself,
you can leverage the work of the Module community to pick up infrastructure improvements through
a version number bump.

## Who maintains this Module?

This Module and its Submodules are maintained by [Gruntwork](http://www.gruntwork.io/). If you are looking for help or
commercial support, send an email to
[support@gruntwork.io](mailto:support@gruntwork.io?Subject=Cloud%20Build%20Module).

Gruntwork can help with:

- Setup, customization, and support for this Module.
- Modules and submodules for other types of infrastructure, such as VPCs, Docker clusters, databases, and continuous
  integration.
- Modules and Submodules that meet compliance requirements, such as HIPAA.
- Consulting & Training on AWS, Terraform, and DevOps.

## How do I contribute to this Module?

Contributions are very welcome! Check out the [Contribution Guidelines](https://github.com/gruntwork-io/terraform-google-ci/blob/master/CONTRIBUTING.md)
for instructions.

## How is this Module versioned?

This Module follows the principles of [Semantic Versioning](http://semver.org/). You can find each new release, along
with the changelog, in the [Releases Page](https://github.com/gruntwork-io/terraform-google-ci/releases).

During initial development, the major version will be 0 (e.g., `0.x.y`), which indicates the code does not yet have a
stable API. Once we hit `1.0.0`, we will make every effort to maintain a backwards compatible API and use the MAJOR,
MINOR, and PATCH versions on each release to indicate any incompatibilities.

## License

Please see [LICENSE](https://github.com/gruntwork-io/terraform-google-ci/blob/master/LICENSE) for how the code in this
repo is licensed.

Copyright &copy; 2019 Gruntwork, Inc.
