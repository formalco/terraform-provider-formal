#!/bin/bash

 

echo "$(basename $(pwd)): Cleaning up the Terraform state if exists"
rm -f .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup output.txt
rm -rf .terraform

expected_error="Group cannot be deleted because termination_protection is set to true"

echo "$(basename $(pwd)): Initializing Terraform"
output=$(terraform init 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Terraform init failed"

    exit $exit_code
fi

echo "$(basename $(pwd)): Running Terraform plan and apply with termination_protection=true"
output=$(terraform apply -auto-approve -var-file=state1.tfvars 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Terraform apply with termination_protection=true failed"

    exit $exit_code
fi

echo "$(basename $(pwd)): Running Terraform destroy with termination_protection=true"
output=$(terraform destroy -auto-approve -var-file=state1.tfvars 2>&1)
exit_code=$?
if [ "$exit_code" -eq 1 ] && echo "$output" | grep -q "$expected_error"; then
    echo "$(basename $(pwd)): Success, the destroy command with termination_protection=true failed as expected"
else
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Failure, the destroy command with termination_protection=true did not fail as expected"

    exit 1
fi

echo "$(basename $(pwd)): Running Terraform plan and apply with termination_protection=false"
output=$(terraform apply -auto-approve -var-file=state2.tfvars 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Terraform apply failed with termination_protection=false"

    exit $exit_code
fi

echo "$(basename $(pwd)): Running Terraform destroy with termination_protection=false"
output=$(terraform destroy -auto-approve -var-file=state2.tfvars 2>&1)
exit_code=$?
if [ $exit_code -ne 0 ]; then
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Terraform destroy failed with termination_protection=false"

    exit $exit_code
fi

echo "$(basename $(pwd)): Terraform test Success"
