provider "google" {
  project = var.project_id
  region  = var.region
}

module "gke" {
  source     = "terraform-google-modules/kubernetes-engine/google"
  project_id = var.project_id
  name       = "ecfr-cluster"
  region     = var.region

  enable_autopilot = true

  network    = "default"
  subnetwork = "default"

  ip_range_pods     = ""
  ip_range_services = ""

  http_load_balancing        = true
  horizontal_pod_autoscaling = true
  filestore_csi_driver       = false

  identity_namespace = "enabled"
}

resource "google_storage_bucket" "parquet" {
  name                        = var.parquet_bucket_name
  location                    = var.region
  uniform_bucket_level_access = true
  force_destroy               = var.bucket_force_destroy

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = 365
    }
  }
}

resource "google_storage_bucket" "raw_xml" {
  name                        = var.raw_xml_bucket_name
  location                    = var.region
  uniform_bucket_level_access = true
  force_destroy               = var.bucket_force_destroy

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = 365
    }
  }
}

resource "google_storage_bucket" "web_static" {
  name                        = var.web_static_bucket_name
  location                    = var.region
  uniform_bucket_level_access = true
  force_destroy               = var.bucket_force_destroy
}

resource "google_service_account" "app" {
  account_id   = "ecfr-app"
  display_name = "eCFR app GKE workload identity service account"
}

resource "google_storage_bucket_iam_member" "app_parquet_rw" {
  bucket = google_storage_bucket.parquet.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.app.email}"
}

resource "google_storage_bucket_iam_member" "app_raw_rw" {
  bucket = google_storage_bucket.raw_xml.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.app.email}"
}

resource "google_storage_bucket_iam_member" "app_web_rw" {
  bucket = google_storage_bucket.web_static.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.app.email}"
}

resource "google_service_account_iam_member" "app_workload_identity" {
  service_account_id = google_service_account.app.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${var.app_namespace}/${var.app_k8s_service_account}]"
}

resource "google_secret_manager_secret" "anthropic_key" {
  secret_id = "anthropic-api-key"
  replication {
    auto {}
  }
}

