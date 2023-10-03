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
    value = ""
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
    value = ""
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