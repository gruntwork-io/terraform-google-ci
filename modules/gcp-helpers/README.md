# GCP Helpers

This module contains helper scripts that automate common GCP tasks:

* `install-gcloud`: This script is meant to be run in a CircleCI job to install the latest version of the Google Cloud SDK CLI tool. Currently, only Ubuntu and Debian are supported.

## Installing the helpers

You can install the helpers using the [Gruntwork Installer](https://github.com/gruntwork-io/gruntwork-installer):

```bash
gruntwork-install --module-name "gcp-helpers" --repo "https://github.com/gruntwork-io/terraform-aws-ci" --tag "v0.0.1"
```

We recommend running this command in the `dependencies` section of `circle.yml`:

```yaml
dependencies:
  override:
    # Install the Gruntwork Installer
    - curl -Ls https://raw.githubusercontent.com/gruntwork-io/gruntwork-installer/master/bootstrap-gruntwork-installer.sh | bash /dev/stdin --version v0.0.16

    # Use the Gruntwork Installer to install the gcp-helpers module
    - gruntwork-install --module-name "gcp-helpers" --repo "https://github.com/gruntwork-io/terraform-aws-ci" --tag "v0.0.1"
```
