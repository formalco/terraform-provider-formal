#!/bin/bash
base_dir="."

subdirs=$(find "$base_dir" -mindepth 1 -maxdepth 1 -type d ! -name "not_included_yet")

# Iterate over the subdirectories
for subdir in $subdirs; do
    # Navigate to the subdirectory
    cd "$subdir" || continue

    echo "$subdir: Testing"
    ./run_test.sh
    test_exit_code=$?
    if [ $test_exit_code -ne 0 ]; then
        echo ""
        echo "$subdir: Test failed. Exiting script."
        exit $test_exit_code
    fi

    # Return to the original directory
    cd - > /dev/null
done
