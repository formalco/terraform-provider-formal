variable "region" {
  default = "ap-southeast-2"
}

# variable "formal_api_key" {
#   type      = string
#   sensitive = true
# }

variable "name" {}
variable "environment" {}

variable "cidr" {
  default = "172.0.0.0/16"
}
variable "private_subnets" {
  default = ["172.0.0.0/20", "172.0.32.0/20", "172.0.64.0/20"]
}
variable "public_subnets" {
  default = ["172.0.16.0/20", "172.0.48.0/20", "172.0.80.0/20"]
}
variable "availability_zones" {
  default = ["ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"]
}

variable "container_cpu" {
  default = 2048
}
variable "container_memory" {
  default = 4096
}

variable "chart_oci" {
  default = "oci://public.ecr.aws/d6o8b0b1/formal-http-helm-chart"
}

variable "formal_sidecar_cert" {
  default = "{\"private_key\":\"-----BEGIN PRIVATE KEY-----\\nMIHcAgEBBEIASZV9Q3MUyNr2VAIuWCGZqk4c72/ZdX1GNoXs/UHXOopzNQUHiFyx\\n+0oX4bhjLS2/oEQRrw3W8ZhR/002+kfbLTKgBwYFK4EEACOhgYkDgYYABADkpBzG\\nRlZoqGdvDvQgDcQbZ4sA+NfXbBLFwJYPZi2WWZhgr9MPBmksyMqLB1+Z/p5OOHKD\\n7oKSHUiVMCJG3dgbUQGEW8siejU2nTJBzRCXQyq8NZmoxaR8PYSouL34X/P953yx\\nN9TljAb42wFQ1hVKOAYf9GckjKjOMe3AgEukovtPUw==\\n-----END PRIVATE KEY-----\\n\",\"certificate\":\"-----BEGIN CERTIFICATE-----\\nMIICXzCCAeagAwIBAgIRANg3AHWki9WL7h4KQqwzvC4wCgYIKoZIzj0EAwQwJjER\\nMA8GA1UECgwIRm9ybWFsQ28xETAPBgNVBAMMCEZvcm1hbENvMB4XDTIzMTAwNTEy\\nMDcwOVoXDTI0MTAwNTEzMDcwOVowXDEtMCsGA1UEChMkNDJjMjVhMTctYTcxNy00\\nY2E3LTk3ZTEtMTBmM2IzYzNkMGQyMSswKQYDVQQDDCJzaWRlY2FyXzAxaGJ6ejV5\\nYnJlcWE5NzlmeWoyZTJuYTZxMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQA5KQc\\nxkZWaKhnbw70IA3EG2eLAPjX12wSxcCWD2YtllmYYK/TDwZpLMjKiwdfmf6eTjhy\\ng+6Ckh1IlTAiRt3YG1EBhFvLIno1Np0yQc0Ql0MqvDWZqMWkfD2EqLi9+F/z/ed8\\nsTfU5YwG+NsBUNYVSjgGH/RnJIyozjHtwIBLpKL7T1OjfDB6MAkGA1UdEwQCMAAw\\nHwYDVR0jBBgwFoAUHzMZhnpYvTnAz3loX51e0mRgGSowHQYDVR0OBBYEFDxDpa7s\\nvtUSatXEHUW3TAGQLmiIMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEF\\nBQcDAQYIKwYBBQUHAwIwCgYIKoZIzj0EAwQDZwAwZAIwEydob+4h08B42vt51Vy8\\njhegSNKZn9Q994DEcSOoELHHNnUfy4BDqSq1+Jj9xnJQAjAgbiEjZHY414riuxnq\\nMBEk7xklEvkXDw/q8PKzjDo8CiTdDvqWXhs54rZYjC1PWnk=\\n-----END CERTIFICATE-----\"}"
}

variable "formal_data_classifier_cert" {
  default = "{\"private_key\":\"-----BEGIN EC PRIVATE KEY-----\\nMIHcAgEBBEIB+MQEDmzegplFVXuVOAjXGiHHQ8MqfC04+SE8SAS1n7LJ+ApDC1lp\\nEjBPWAdOc891HAk7iX74OuGYqsxDL4tkBrWgBwYFK4EEACOhgYkDgYYABAEPSpil\\nXL2FJ4//qveaMxMnojUfGG91IPXPIMxfzSfnMI6adIGhs4YkMuCrt6R7OI4TOVkK\\n4tMjiR2fiTctlkilvADZGltC6V+o2TJBnzyg1hepL/HMXJjcAhxgJSF0GnsJkbc8\\nHWi8IFXo24tQQRi48OGZUyniC7zvxfwk02kdEInO8w==\\n-----END EC PRIVATE KEY-----\\n\",\"certificate\":\"-----BEGIN CERTIFICATE-----\\nMIICYjCCAeigAwIBAgIRAKlQ1EKavj0qT64rEmlDwbkwCgYIKoZIzj0EAwQwJjER\\nMA8GA1UECgwIRm9ybWFsQ28xETAPBgNVBAMMCEZvcm1hbENvMB4XDTIzMTAwNDEx\\nMzk0MloXDTI0MTAwNDEyMzk0MlowXjEtMCsGA1UEChMkNDJjMjVhMTctYTcxNy00\\nY2E3LTk3ZTEtMTBmM2IzYzNkMGQyMS0wKwYDVQQDDCRzYXRlbGxpdGVfMDFoYnhi\\nNnpzNmU4enI0dnFzeXQzYjBlMTYwgZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAEP\\nSpilXL2FJ4//qveaMxMnojUfGG91IPXPIMxfzSfnMI6adIGhs4YkMuCrt6R7OI4T\\nOVkK4tMjiR2fiTctlkilvADZGltC6V+o2TJBnzyg1hepL/HMXJjcAhxgJSF0GnsJ\\nkbc8HWi8IFXo24tQQRi48OGZUyniC7zvxfwk02kdEInO86N8MHowCQYDVR0TBAIw\\nADAfBgNVHSMEGDAWgBQfMxmGeli9OcDPeWhfnV7SZGAZKjAdBgNVHQ4EFgQUU4Jt\\nQljRx98RUjfP51iQKECEPfMwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsG\\nAQUFBwMBBggrBgEFBQcDAjAKBggqhkjOPQQDBANoADBlAjEAjuetaDfGTQ0Rcj+U\\nzm8vm9nmHcOv2W3INKRrzk3/l74PyN3gtEGAKEwZSZpHJojRAjA7EIn0ZKPxKWDT\\nL+on1EAGIZjMeAAs61qWZqYvOD/d4DNDxNPxTnXg6u1Ba9hVT08=\\n-----END CERTIFICATE-----\"}"
}