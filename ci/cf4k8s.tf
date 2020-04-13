provider "google" {
  project = "cff-diego-persistence"
  region  = "us-west1"
  zone    = "us-west1-a"
}

variable "cluster_name" {
  type = string
}

resource "google_compute_address" "load_balancer_external_ip" {
  name         = var.cluster_name
  address_type = "EXTERNAL"
  region       = "us-west1"
}

resource "google_dns_managed_zone" "cluster_zone" {
  name        = var.cluster_name
  dns_name    = "${var.cluster_name}.persi.cf-app.com."
  description = "DNS for ${var.cluster_name}.persi.cf-pp.com"
}

resource "google_dns_record_set" "a" {
  name = "*.${google_dns_managed_zone.cluster_zone.dns_name}"
  managed_zone = google_dns_managed_zone.cluster_zone.name
  type = "A"
  ttl  = 300

  rrdatas = [google_compute_address.load_balancer_external_ip.address]
}

resource "google_container_cluster" "test_cluster" {
  name     = var.cluster_name
  location = "us-west1-a"

  remove_default_node_pool = false
  initial_node_count       = 3

  node_config {
    preemptible  = true
    machine_type = "n1-standard-4"
    image_type = "ubuntu"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }

  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = false
    }
  }
}

output "cluster_name" {
  value = var.cluster_name
}

output "name_servers" {
  value = google_dns_managed_zone.cluster_zone.name_servers
}

output "lb_ip" {
  value = google_compute_address.load_balancer_external_ip.address
}

output "zone" {
  value = "us-west1-a"
}