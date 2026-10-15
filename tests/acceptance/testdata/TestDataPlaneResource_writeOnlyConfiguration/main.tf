variable "id" {
  type = string
}

variable "name" {
  type = string
}

variable "mongodb_host" {
  type = string
}

variable "mongodb_port" {
  type = number
}

resource "am_data_plane" "test" {
  id   = var.id
  name = var.name
  type = "mongodb"
  configuration = jsonencode({
    mongodb = {
      dbname = "gravitee-am-${var.id}"
      host   = var.mongodb_host
      port   = var.mongodb_port
    }
  })
}

resource "am_domain" "on_data_plane" {
  key           = "${var.id}-domain"
  name          = "Terraform acceptance on a data plane"
  path          = "/${var.id}-domain"
  data_plane_id = am_data_plane.test.id
}
