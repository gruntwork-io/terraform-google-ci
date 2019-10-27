output "registry_url" {
  description = "The URL at which the registry can be accessed."
  value       = local.registry_url
}

output "bucket" {
  description = "Name of the underlying GCS bucket."
  value       = local.bucket_name
}
