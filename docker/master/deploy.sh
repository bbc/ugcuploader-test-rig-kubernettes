rm -rf src
cp -R ../../src .
REPO="$1.dkr.ecr.$2.amazonaws.com/ugcloadtest/jmeter-master:latest"
echo "Repo: $REPO"
aws ecr delete-repository --force --repository-name ugcloadtest/jmeter-master
aws ecr create-repository --repository-name ugcloadtest/jmeter-master
docker build --pull --no-cache  -t ugcloadtest/jmeter-master .
docker tag ugcloadtest/jmeter-master:latest $REPO
docker push $REPO 
