set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
version_file=$PROJECT_ROOT/interfacer/src/browsh/version.go
line=$(cat $version_file | grep 'browshVersion')
version=$(echo $line | grep -o '".*"' | sed 's/"//g')
echo -n $version
