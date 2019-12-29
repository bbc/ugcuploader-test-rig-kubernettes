export AWS_DEFAULT_REGION=eu-west-2
aws ecr delete-repository --force --repository-name ugcloadtest/jmeter-slave
aws ecr create-repository --repository-name ugcloadtest/jmeter-slave 
docker build --pull --no-cache  -t ugcloadtest/jmeter-slave .
docker -v tag  ugcloadtest/jmeter-slave:latest 546933502184.dkr.ecr.eu-west-2.amazonaws.com/ugcloadtest/jmeter-slave:latest
docker -v push  546933502184.dkr.ecr.eu-west-2.amazonaws.com/ugcloadtest/jmeter-slave:latest
