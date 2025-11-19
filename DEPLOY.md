# Deployment Guide

## Local
`docker-compose up`

## GCP
1. `terraform init`
2. `terraform apply -var="project_id=..." -var="gcs_bucket=..."`
3. Build/push images to GCR
4. Apply k8s manifests: deployment.yaml, cronjob.yaml, service.yaml, ingress.yaml
  - Deployment: api image, mount GCS via gcsfuse
  - CronJob: etl image, daily
  - Secrets: from Secret Manager
5. Access via ingress URL
