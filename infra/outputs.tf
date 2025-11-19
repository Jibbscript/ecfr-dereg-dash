output "kubernetes_cluster_name" {
  value       = module.gke.name
  description = "GKE Cluster Name"
}

output "kubernetes_endpoint" {
  value       = module.gke.endpoint
  description = "GKE Endpoint"
}
