terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 4.0.0"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_satellite" "main" {
  name = "demo-data-classifier"
}
