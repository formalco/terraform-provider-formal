# BigQuery Sidecar & Data Classifier Satellite Helm Chart

This document provides instructions for installing and configuring the Formal
BigQuery Sidecar and Data Classifier Satellite Helm chart on your Kubernetes cluster.

## 1. Overview

This Helm chart automates the deployment process of two applications, the BigQuery Sidecar, and the Data Classifier Satellite, which are essential for managing and classifying BigQuery data traffic within your Kubernetes environment.

## 2. Prerequisites

Before proceeding with the installation of this Helm chart, please ensure that
you have:

- A running Kubernetes cluster, version 1.19 or above.
- Helm, version 3.0 or above, installed and configured on your local machine or
  CI/CD environment.
- Appropriate permissions to deploy resources to the desired namespace.

## 3. Quick Start

To quickly deploy the applications on your cluster with default configurations, run the following command, replacing bq-sidecar with your desired release name:

```shell
helm repo add formal http://localhost:8090

helm install formal/bq-sidecar --generate-name
```

## 4. Configuration Options

You can customize the deployment through the following configurations in your `values.yaml` file:

### General Configurations

- namespace: The deployment namespace (Default: default).
- replicaCount: Number of replicas (Default: 1).
- imagePullSecret: The secret containing docker registry credentials (Default: registry-test-secret).

### Containers Configurations

#### BigQuery Sidecar Container

- containers.bigQuerySidecar.name: Container name (Default: bigquery-sidecar-app).
- containers.bigQuerySidecar.image: Docker image (Default: formalco/docker-prod-bigquery-sidecar:latest).
- containers.bigQuerySidecar.resources: Resource requests and limits.

#### Data Classifier Satellite Container

- containers.dataClassifierSatellite.name: Container name (Default: data-classifier-satellite-app).
- containers.dataClassifierSatellite.image: Docker image (Default: formalco/docker-prod-data-classifier-satellite:latest).
- containers.dataClassifierSatellite.resources: Resource requests and limits.

### Config Maps

#### BigQuery Sidecar ConfigMap

Customizable parameters for BigQuery Sidecar behavior including TLS configuration, PII detection, server connections, and more.

#### Data Classifier Satellite ConfigMap

Configuration related to PII detection methods and confidence levels, as well as TLS certificate data.

### Service Configuration

- service.type: Kubernetes service type (Default: ClusterIP).
- service.port: Exposed service port (Default: 443).

### Secret Configuration

secret.dockerconfigjson: Docker registry secret in JSON format.