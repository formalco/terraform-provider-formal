provider "helm" {
  kubernetes {
    host                   = var.eks_cluster_endpoint
    cluster_ca_certificate = base64decode(var.eks_cluster_certificate_authority_data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      # This requires the awscli to be installed locally where Terraform is executed
      args        = ["eks", "get-token", "--cluster-name", var.eks_cluster_name]
    }
  }
}

resource "helm_release" "example" {
  name        = "formal-http"
  chart  =  var.chart_oci
  version     = "0.2.0"
  # values = [
  #   "${file("./helm/values.yaml")}"
  # ]

  set {
    name  = "replicaCount"
    value = "1"
  }

  set {
    name  = "containers.httpSidecar.name"
    value = "http-sidecar-app"
  }

  set {
    name  = "containers.httpSidecar.image"
    value = "323330447930.dkr.ecr.eu-west-1.amazonaws.com/gusto-http-proxy"
  }


  set {
    name  = "containers.dataClassifierSatellite.name"
    value = "data-classifier-satellite-app"
  }

  set {
    name  = "containers.dataClassifierSatellite.image"
    value = "323330447930.dkr.ecr.eu-west-1.amazonaws.com/gusto-data-classifier"
  }


  set {
    name  = "containers.configMaps.httpSidecar.CLIENT_LISTEN_TLS"
    value = "false"
  }

  set {
    name  = "containers.configMaps.httpSidecar.SERVER_CONNECT_TLS"
    value = "false"
  }

  set {
    name  = "containers.configMaps.httpSidecar.CUSTOMER_TLS_CERT_PRIVATE_KEY"
    value = ""
  }

  set {
    name  = "containers.configMaps.httpSidecar.CUSTOMER_TLS_CERT_FULLCHAIN"
    value = ""
  }

  set {
    name  = "containers.configMaps.httpSidecar.CONTROL_PLANE_CONFIG_URI"
    value = "config.proxy.api.formalcloud.net:443"
  }

  set {
    name  = "containers.configMaps.httpSidecar.STRIP_VALUES_FROM_LOGS"
    value = "false"
  }

  set_sensitive {
    name  = "containers.configMaps.httpSidecar.FORMAL_CONTROL_PLANE_TLS_CERT"
    value = "{\"private_key\":\"-----BEGIN PRIVATE KEY-----\\nMIHcAgEBBEIBvNVnV3bojJ9BaVOv75J/I3Qaowm0o7UObiKJ/7nj9r9CGo8xSNW9\\nZRNbh6kA5oihLWYUltxuGmQtHBrzfTYKpIWgBwYFK4EEACOhgYkDgYYABADpZHNV\\nsLm43ZZHTELTyFBfUKmXxjuKygGaeJu/FT9U7c4TSUrGesgkwVDchl50S91rQWPf\\nor+vVn/qg/h82aIS0gGbJnqpLZZ97nKywB6oce0kANqNprJ55YgyXqXYifHN+QiD\\nFWQXXpghLwz2s/2r3qWntdH5NtNeBqIpbKAz4iUUww==\\n-----END PRIVATE KEY-----\\n\",\"certificate\":\"-----BEGIN CERTIFICATE-----\\nMIICXzCCAeWgAwIBAgIQam2lGKc5xnudvQ7vRuVEvjAKBggqhkjOPQQDBDAmMREw\\nDwYDVQQKDAhGb3JtYWxDbzERMA8GA1UEAwwIRm9ybWFsQ28wHhcNMjMxMDAyMjMy\\nNzU5WhcNMjQxMDAzMDAyNzU5WjBcMS0wKwYDVQQKEyQ0MmMyNWExNy1hNzE3LTRj\\nYTctOTdlMS0xMGYzYjNjM2QwZDIxKzApBgNVBAMMInNpZGVjYXJfMDFoYnNleWVz\\ncGZmZHRnMWZhcGt0ZDBjcjMwgZswEAYHKoZIzj0CAQYFK4EEACMDgYYABADpZHNV\\nsLm43ZZHTELTyFBfUKmXxjuKygGaeJu/FT9U7c4TSUrGesgkwVDchl50S91rQWPf\\nor+vVn/qg/h82aIS0gGbJnqpLZZ97nKywB6oce0kANqNprJ55YgyXqXYifHN+QiD\\nFWQXXpghLwz2s/2r3qWntdH5NtNeBqIpbKAz4iUUw6N8MHowCQYDVR0TBAIwADAf\\nBgNVHSMEGDAWgBQfMxmGeli9OcDPeWhfnV7SZGAZKjAdBgNVHQ4EFgQUJWpz24FJ\\n0zNoJGvoe62yK8MbgR8wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUF\\nBwMBBggrBgEFBQcDAjAKBggqhkjOPQQDBANoADBlAjEA0JOKOu3DZ/3ojDZHhPfw\\nVfU9tGZaB+iJACDij3kKjMheds1qiEXpB06JtKhzTRtZAjAI2oxsJoqU+FenufRt\\n/yJyI1vSaYmMbwKkbF5cV/4AII1tOaOxccQHqlpMkU3XaQ8=\\n-----END CERTIFICATE-----\"}"
  }

  set {
    name  = "containers.configMaps.httpSidecar.DATA_CLASSIFIER_SATELLITE_URI"
    value = "localhost:50055"
  }

  set {
    name  = "containers.configMaps.httpSidecar.PII_SAMPLING_RATE"
    value = "2"
  }

  set {
    name  = "containers.configMaps.dataClassifierSatellite.PII_DETECTION"
    value = "formal"
  }

  set_sensitive {
    name  = "containers.configMaps.dataClassifierSatellite.FORMAL_CONTROL_PLANE_TLS_CERT"
    value = "{\"private_key\":\"-----BEGIN EC PRIVATE KEY-----\\nMIHcAgEBBEIAg3HhGiFhMiAASq3kdkgroTdSnQ2ETJU56RPquP8rnNZd2qKZBSZg\\nZ8ogXWu1r7Z12AQiLVDZceTZJPyLIa1ZA6ugBwYFK4EEACOhgYkDgYYABAEsLeEs\\nON50tV9VDdZXv//F3z+IKKiHrq2DwS8j3PvosCxd50KfjfhZxQpadSlBHJMq0MnM\\nXNohcPrmcGfAmCsD+ACOrrYRZOXohWRhHyoZV7GXzuusCY75RKOeVhOsVtwp1Cgt\\nJ1nzQ9qfxKax3hExHLwkr73CpmFMmiSeNqLCU21QDA==\\n-----END EC PRIVATE KEY-----\\n\",\"certificate\":\"-----BEGIN CERTIFICATE-----\\nMIICYjCCAeigAwIBAgIRANLskLGAORB3aY3VHeKyGK8wCgYIKoZIzj0EAwQwJjER\\nMA8GA1UECgwIRm9ybWFsQ28xETAPBgNVBAMMCEZvcm1hbENvMB4XDTIzMDkyNjE1\\nMzg1OFoXDTI0MDkyNjE2Mzg1OFowXjEtMCsGA1UEChMkNDJjMjVhMTctYTcxNy00\\nY2E3LTk3ZTEtMTBmM2IzYzNkMGQyMS0wKwYDVQQDDCRzYXRlbGxpdGVfMDFoYjk1\\ncWIzNGZ4ZzhlOTVkNWc2YjV3OWIwgZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAEs\\nLeEsON50tV9VDdZXv//F3z+IKKiHrq2DwS8j3PvosCxd50KfjfhZxQpadSlBHJMq\\n0MnMXNohcPrmcGfAmCsD+ACOrrYRZOXohWRhHyoZV7GXzuusCY75RKOeVhOsVtwp\\n1CgtJ1nzQ9qfxKax3hExHLwkr73CpmFMmiSeNqLCU21QDKN8MHowCQYDVR0TBAIw\\nADAfBgNVHSMEGDAWgBQfMxmGeli9OcDPeWhfnV7SZGAZKjAdBgNVHQ4EFgQUjYpw\\nl1QKZy/JaxDyBP1cd0ibt+UwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsG\\nAQUFBwMBBggrBgEFBQcDAjAKBggqhkjOPQQDBANoADBlAjAyMkytb7mYMkoOJaV0\\n8ax+h7vgOtsvQ9NgLHCUdX2iocK4lXBQtcKwvwHLaBCpk0MCMQDGaZDMWxzlm5TN\\nVqJlWPCpxYBJejwLu1LbRnRhL2MvD7I4tPXCQcTbVNCUCUx59Wk=\\n-----END CERTIFICATE-----\"}"
  }


set {
    name  = "service.type"
    value = "LoadBalancer"
  }

  set {
    name  = "service.port"
    value = "443"
  }
}