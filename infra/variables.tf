variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "us-central1"
}

variable "parquet_bucket_name" {
  type        = string
  description = "GCS bucket for parquet files and summaries."
}

variable "raw_xml_bucket_name" {
  type        = string
  description = "GCS bucket for raw XML files from govinfo."
}

variable "web_static_bucket_name" {
  type        = string
  description = "GCS bucket for static frontend files."
}

variable "bucket_force_destroy" {
  type        = bool
  default     = false
  description = "If true, allows terraform destroy to delete buckets even if they contain objects."
}

variable "app_namespace" {
  type        = string
  default     = "default"
  description = "Kubernetes namespace where the app runs."
}

variable "app_k8s_service_account" {
  type        = string
  default     = "ecfr-app"
  description = "Kubernetes ServiceAccount name used by the app."
}

