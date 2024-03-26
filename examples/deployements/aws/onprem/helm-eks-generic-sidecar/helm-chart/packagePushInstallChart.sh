#!/bin/bash


# Check if all required arguments are provided
if [ $# -ne 3 ]; then
  echo "Usage: $0 <chart_name> <ecr_repository_uri> <region>"
  exit 1
fi

# Assign command-line arguments to variables
chart_name="$1"
ecr_repository_uri="$2"
region="$3"

# Package the Helm chart in the current directory
helm package .

# Get the AWS ECR login token and log in to the repository
aws ecr get-login-password --region "$region" | helm registry login \
  --username AWS \
  --password-stdin "$ecr_repository_uri"

# Push the Helm chart to ECR
helm push "$(ls -t *.tgz | head -1)" "oci://$ecr_repository_uri"

# Ensure the Helm chart is not already installed
helm uninstall "$chart_name"
# Install the Helm chart from ECR
helm install "$chart_name" "oci://$ecr_repository_uri/$chart_name" --version 0.1.0 -f values.yaml