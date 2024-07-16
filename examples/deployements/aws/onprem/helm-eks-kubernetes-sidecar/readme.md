## How it works?

In order to install the Helm chart on the Kubernetes cluster, we have to provision:
1. a EKS cluster (AWS)
2. a ECR repository (containing the helm chart .tgz)

Therefore, the first step is to run the terraform to provision those two resources.

The second step is to run the helm-chart script by running the following command:
```./packagePushInstallChart.sh <name_chart> <account_id>.dkr.ecr.<region>.amazonaws.com <region>```

### Troubleshooting
1. Make sure that <name_chart> is matching the name of the chart in `Chart.yaml`
2. Make sure that you have configure your AWS credentials when running the terraform / script file.