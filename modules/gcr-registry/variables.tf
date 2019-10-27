# ---------------------------------------------------------------------------------------------------------------------
# MODULE PARAMETERS
# These variables are expected to be passed in by the operator
# ---------------------------------------------------------------------------------------------------------------------

variable "project" {
  description = "The project ID to create the resources in."
  type        = string
}

# See https://cloud.google.com/container-registry/docs/pushing-and-pulling#pushing_an_image_to_a_registry
variable "registry_region" {
  description = "The region for the registry (e.g. 'us', 'eu', 'asia')."
  type        = string
}

# ---------------------------------------------------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# Generally, these values won't need to be changed.
# ---------------------------------------------------------------------------------------------------------------------

variable "readers" {
  description = "List of identities that will be granted read privileges (e.g. 'user:john.doe@acme.com')"
  type        = list(string)
  default     = []
}

variable "writers" {
  description = "List of identities that will be granted read/write privileges (e.g. 'user:john.doe@acme.com')"
  type        = list(string)
  default     = []
}

variable "init_registry" {
  description = "If 'true', then a scratch image is pushed, ensuring the underlying bucket is created. Set to 'false' if you want to skip."
  type        = bool
  default     = true
}
