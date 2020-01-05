
aws iam create-policy --policy-name ugcupload-eks-jmeter-policy --policy-document file://./i-am-policy-jmeter.json
aws iam create-policy --policy-name ugcupload-eks-control-policy --policy-document file://./i-am-policy-control.json
eksctl create cluster -f cluster.yaml
kubectl apply -k "github.com/kubernetes-sigs/aws-ebs-csi-driver/deploy/kubernetes/overlays/stable/?ref=master"
kubectl create -f csi-storage-class.yaml