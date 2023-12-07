#!/bin/bash

echo "$(basename $(pwd)): Cleaning up the Terraform state if exists"
rm -f .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup output.txt
rm -rf .terraform

expected_error="expected formal_identity_type to be one of"

echo "$(basename $(pwd)): Initializing Terraform"
output=$(terraform init 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Terraform init failed"

    exit $exit_code
fi

state=$(cat state1.tfvars)
echo "$(basename $(pwd)): Running Terraform plan with $state"
output=$(terraform plan -var-file=state1.tfvars 2>&1)
exit_code=$?
if [ "$exit_code" -eq 1 ] && echo "$output" | grep -q "$expected_error"; then
    echo "$(basename $(pwd)): Success, the plan command with $state failed as expected"
else
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Failure, the plan command with $state did not fail as expected"

    exit 1
fi

state=$(cat state2.tfvars)
echo "$(basename $(pwd)): Running Terraform plan with $state"
output=$(terraform plan -var-file=state2.tfvars 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "Error:"
    echo "$output"
    echo "$(basename $(pwd)): Terraform plan command with $state failed"

    exit $exit_code
fi

echo "$(basename $(pwd)): Terraform test Success"
