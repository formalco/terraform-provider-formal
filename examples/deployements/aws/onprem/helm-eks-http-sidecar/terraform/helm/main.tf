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

# locals {
#   http_tls_cert = {
#     containers = {
#       configMaps = {
#         httpSidecar = {
#           FORMAL_CONTROL_PLANE_TLS_CERT = <<EOF
#           [
#             {
#             "private_key": "-----BEGIN EC PRIVATE KEY-----\nMIHcAgEBBEIB+MQEDmzegplFVXuVOAjXGiHHQ8MqfC04+SE8SAS1n7LJ+ApDC1lp\nEjBPWAdOc891HAk7iX74OuGYqsxDL4tkBrWgBwYFK4EEACOhgYkDgYYABAEPSpil\nXL2FJ4//qveaMxMnojUfGG91IPXPIMxfzSfnMI6adIGhs4YkMuCrt6R7OI4TOVkK\n4tMjiR2fiTctlkilvADZGltC6V+o2TJBnzyg1hepL/HMXJjcAhxgJSF0GnsJkbc8\nHWi8IFXo24tQQRi48OGZUyniC7zvxfwk02kdEInO8w==\n-----END EC PRIVATE KEY-----\n",
#             "certificate": "-----BEGIN CERTIFICATE-----\nMIICYjCCAeigAwIBAgIRAKlQ1EKavj0qT64rEmlDwbkwCgYIKoZIzj0EAwQwJjER\nMA8GA1UECgwIRm9ybWFsQ28xETAPBgNVBAMMCEZvcm1hbENvMB4XDTIzMTAwNDEx\nMzk0MloXDTI0MTAwNDEyMzk0MlowXjEtMCsGA1UEChMkNDJjMjVhMTctYTcxNy00\nY2E3LTk3ZTEtMTBmM2IzYzNkMGQyMS0wKwYDVQQDDCRzYXRlbGxpdGVfMDFoYnhi\nNnpzNmU4enI0dnFzeXQzYjBlMTYwgZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAEP\nSpilXL2FJ4//qveaMxMnojUfGG91IPXPIMxfzSfnMI6adIGhs4YkMuCrt6R7OI4T\nOVkK4tMjiR2fiTctlkilvADZGltC6V+o2TJBnzyg1hepL/HMXJjcAhxgJSF0GnsJ\nkbc8HWi8IFXo24tQQRi48OGZUyniC7zvxfwk02kdEInO86N8MHowCQYDVR0TBAIw\nADAfBgNVHSMEGDAWgBQfMxmGeli9OcDPeWhfnV7SZGAZKjAdBgNVHQ4EFgQUU4Jt\nQljRx98RUjfP51iQKECEPfMwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsG\nAQUFBwMBBggrBgEFBQcDAjAKBggqhkjOPQQDBANoADBlAjEAjuetaDfGTQ0Rcj+U\nzm8vm9nmHcOv2W3INKRrzk3/l74PyN3gtEGAKEwZSZpHJojRAjA7EIn0ZKPxKWDT\nL+on1EAGIZjMeAAs61qWZqYvOD/d4DNDxNPxTnXg6u1Ba9hVT08=\n-----END CERTIFICATE-----"
#             }
#           ]
#            EOF 
#         }
#       }
#     }
#   }
# }

resource "helm_release" "example" {
  name        = "formal-http"
  chart  =  var.chart_oci
  version     = "0.4.0"
  values = [
    "${file("./helm/values.yaml")}"
  ]
}