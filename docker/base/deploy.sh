REPO="$1.dkr.ecr.$2.amazonaws.com/ugcloadtest/jmeter-base:latest"
aws ecr delete-repository --force --repository-name ugcloadtest/jmeter-base
aws ecr create-repository --repository-name ugcloadtest/jmeter-base 
docker build --no-cache -t ugcloadtest/jmeter-base .
docker tag ugcloadtest/jmeter-base:latest $REPO
docker push $REPO
