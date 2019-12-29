Table of Contents
=================

  - [Create Cluster](#create--cluster)

    + [S3 Bucket](#s3--bucket)

  - [Create SnapShot of Graphana and Influxdb](#create--snapshot--of--graphana--and--influxdb)
    
  - [Increase Size of NodeGroup](#increase--size--of--nodegroup)
    
    

#Create Cluster

The following steps should be followed to create the Test Rig cluster.

In the folder *artefacts/cluster* The script *create-cluster.sh* can be run to create the cluster.

Below are the steps performed by the script.

| Step | Type                        | Description                                                  | Action                                                       |
| ---- | --------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 1    | Create I am Policy          | This is the policy that gives the pods in the cluster access to AWS resources. [Policy Document](#i-am-policy-jmeter.json). The arn will be used when we are creating ServiceAccount's for the virtual clusters. | *aws iam create-policy --policy-name ugcupload-eks-jmeter-policy --policy-document file://i-am-policy-jmeter.json* |
| 2    | Create cluster              | Create the Test Rig                                          | *eksctl create cluster -f cluster.yaml*                      |
| 3    | EBS CSI driver Installation | Responsible for co-ordinating interactions with AWS volumes  | *kubectl apply -k "github.com/kubernetes-sigs/aws-ebs-csi-driver/deploy/kubernetes/overlays/stable/?ref=master"* |
| 4    | Create a storage class      | The storage class used to access AWS EBS volumes             | *kubectl create -f csi-storage-class.yaml*                   |
| 5    | Create snapshot class       | The class used to associate a snapshot with the storage class. | *kubectl create -f csi-snapshot-class.yaml*                  |



### S3 Bucket 

When the test complete the results are uploaded to an s3 bucket *ugcupload-jmeter*. This must be created in the same account and region of the cluster.

 The I AM policy and ServiceAccount were created to provide access to this resource. 

# Create SnapShot of Graphana and Influxdb

Currently you can not take snapshots of AWS EBS volumes using the approach outlined here: https://kubernetes.io/blog/2018/10/09/introducing-volume-snapshot-alpha-for-kubernetes/. 

This is because EKS contol plane does not support alpha features as outlined here: https://github.com/aws/containers-roadmap/issues/146.  Should be available in the first quarter of 2020.



# Increase Size of NodeGroup

In the folder `kubernetes-util`use the script `scale-node-group.sh`. <br> Usage: *scale-node-group.sh <nodegroup> <size>*

Eg. To increase the amount of nodes for the slaves to 50.

`./scale-node-group.sh jmeter-slaves 50`