# Authentication methods with KubeVirt client libraries

- [Authentication methods with KubeVirt client libraries](#authentication-methods-with-kubevirt-client-libraries)
  - [Introduction](#introduction)
  - [Requirements](#requirements)
  - [Setup](#setup)
    - [Compiling our test application](#compiling-our-test-application)
  - [Running our application externally leveraging a kubeconfig file](#running-our-application-externally-leveraging-a-kubeconfig-file)
    - [Using the default kubeconfig](#using-the-default-kubeconfig)
    - [Creating a kubeconfig for the service account](#creating-a-kubeconfig-for-the-service-account)
  - [Running in a Kubernetes Cluster](#running-in-a-kubernetes-cluster)
  - [Extending RBAC Role across Namespaces](#extending-rbac-role-across-namespaces)
  - [Creating Custom RBAC Roles](#creating-custom-rbac-roles)
  - [References](#references)

## Introduction

The KubeVirt project supplies a Go client library for interacting with KubeVirt. This allows you to write your own automation and programs quickly and easily. We will use a simple application to demonstrate how the KubeVirt client library authenticates with your k8s cluster. We will use the example application in the "client-go" library with a small modification to it, to allow for running both locally and within in the cluster. This tutorial assumes you have some knowledge of Go, and is not meant to be a Go training doc.

## Requirements

In order to run this application locally you will need to have the Go programming language installed on your machine. The steps listed here were tested with Go version 1.16.
We will be testing that we have properly authenticated to the system by listing out the VMI and VM instances in your cluster that you have access too. If you do not have any running VMs in your cluster, create a new project and deploy a virtual machine. (For guidance in creating a quick test vm see the [Use KubeVirt](https://kubevirt.io/labs/kubernetes/lab1.html) lab for some quick instructions. Finally we will be leveraging an OpenShift cluster for this demo, and will be using the "oc" command for our command line interactions with the cluster.

## Setup

### Compiling our test application

Start by cloning this repo and compiling our test application:

```shell
$ git clone https://github.com/xphyr/kubevirt-apiauth.git
$ cd kubevirt-apiauth
$ cd listvms
$ go build
```

Test and ensure that the application compiled correctly. If you have a working k8s context (or are logged into your OpenShift cluster with the oc command), running this command may return some values. If you are not logged in, or do not have a current context, you will get an error. This is OK, we will discuss authentication next.

```shell
$ ./listvms
2021/06/23 16:51:28 cannot obtain KubeVirt vm list: Get "http://localhost:8080/apis/kubevirt.io/v1alpha3/namespaces/default/virtualmachines": dial tcp 127.0.0.1:8080: connect: connection refused
```

Now that we have a working test application, we will move onto the next step, authentication.

## Running our application externally leveraging a kubeconfig file

The default authentication file for Kubernetes is the [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file. We will not be going into details of this file, but you can click the link to goto the documentation on the kubeconfig file to learn more about it.

### Using the default kubeconfig

If you haven't already done so, log into your OpenShift cluster with the "_oc_" command:

```shell
$ oc login https://api.<clustername>.<basedomain> -u <userName>
password: ********
Login successful.

You have access to 80 projects, the list has been suppressed. You can list all projects with 'oc get projects'

Using project "myvms".
```

We now have a valid kubeconfig. This file by default is stored in your home directory at `~/.kube/config`. You should now be able to run our test application and get some results (assuming you have some running vms in your cluster)

```shell
$ ./listvms/listvms
Type                       Name                    Namespace     Status
VirtualMachineInstance     vm-fedora-ephemeral     myvms         Running
```

This is great, but there is an issue. The authentication method we used (logging in with a named user/password) will time out after 24 hours. This wont work for a long running application. You don't want to have to keep logging in every 23.5 hours to keep your application up and running. How do we handle this?  We create a ServiceAccount within our k8s cluster and authenticate as the service account.

[Service Accounts](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/) are accounts for processes as opposed to users. By default they are scoped to a namespace, but you can give service accounts access to other namespaces through RBAC rules that we will discuss later. In this demo, we will be using the "_myvms_" project/namespace, so the service account we create will be initially scoped only to this namespace.

Start by creating a new service account called "mykubevirtrunner":

```shell
$ oc create sa mykubevirtrunner -n myvms
$ oc describe sa mykubevirtrunner
Name:                mykubevirtrunner
Namespace:           myvms
Labels:              <none>
Annotations:         <none>
Image pull secrets:  mykubevirtrunner-dockercfg-p6ssb
Mountable secrets:   mykubevirtrunner-token-qfkmc
                     mykubevirtrunner-dockercfg-p6ssb
Tokens:              mykubevirtrunner-token-qfkmc
                     mykubevirtrunner-token-vq8pb
Events:              <none>
```

In the describe output you can see that there are two tokens that have been generated, these are the tokens that we will use for our authentication. The tokens listed are the names of the kubernetes secrets that were created that contain the token. Let's retrieve the contents of the _second_ token listed:

```shell
$ oc describe secret mykubevirtrunner-token-qfkmc -n myvms
Name:         mykubevirtrunner-token-qfkmc
Namespace:    myvms
Labels:       <none>
Annotations:  kubernetes.io/service-account.name: mykubevirtrunner
              kubernetes.io/service-account.uid: a02b25e8-39bf-4e4b-8680-918521dc0a6b

Type:  kubernetes.io/service-account-token

Data
====
token:       eyJhbGciOiJSUzI1NiI....
ca.crt:          7230 bytes
namespace:       5 bytes
service-ca.crt:  8443 bytes
```

The data listed for the "token" key is the information we will use in the next step, your output will be much longer, it has been truncated for this document. Ensure when copying the value that you get the entire token value.

### Creating a kubeconfig for the service account

We will create a new kubeconfig file that leverages the service account and token we just created. The easiest way to do this is to create an empty kubeconfig file, and use the "_oc_" command to log in with the new token. We will start by setting the KUBECONFIG environment variable to point to a file in our local directory, and then using the "_oc_" command log into the cluster:

```shell
$ export KUBECONFIG=$(pwd)/sa-kubeconfig
$ oc login https://api.<clustername>.example.com --token=<paste token from last step here>
$ oc whoami
system:serviceaccount:myvms:mykubevirtrunner
```

We now have a kubeconfig that authenticates us as the service account that we created in the last section. Try running our test program again:

```shell
$ listvms/listvms
get error:  `2021/06/23 18:12:04 cannot obtain KubeVirt vm list: virtualmachines.kubevirt.io is forbidden: User "system:serviceaccount:myvms:mykubevirtrunner" cannot list resource "virtualmachines" in API group "kubevirt.io" in the namespace "myvms"`
```

You can see we are now using our service account, but that service account doesn't have the right permissions... We will start simple and give the service account the "kubevirt.io:view" role, which will allow the service account to see the KubeVirt objects within the "_myvms_" namespace:

```shell
$ oc policy add-role-to-user kubevirt.io:view system:serviceaccount:myvms:mykubevirtrunner
clusterrole.rbac.authorization.k8s.io/kubevirt.io:view added: "system:serviceaccount:myvms:mykubevirtrunner"
```

Now run the listvms command again:

```shell
./listvms/listvms
Type                       Name                    Namespace     Status
VirtualMachineInstance     vm-fedora-ephemeral     myvms         Running
```

Success! Our application is now using the service account that we created for authentication to the cluster. The service account can be extended by adding additional default roles to the account, or by creating custom roles that limit the scope of the service account to only the exact actions you want to take.

## Running in a Kubernetes Cluster

So all of this is great if you want to run the application outside of your cluster ... but what if you want your application to run INSIDE you cluster. You could create a kubeconfig file, and add it to your namespace as a secret and then mount that secret as a volume inside your pod, but there is an easier way that continues to leverage the service account that we created. By default kubernetes creates a few environment variables for every pod that indicate that the container is running within kubernetes, and it makes a kubernetes auth token for the service account that the container is running as available at /var/run/secrets/kubernetes.io/serviceaccount/token.  The client-go KubeVirt library can detect that it is running inside a kubernetes hosted container and will transparently use the auth token provided with no additional configuration needed.

A container image with the listvms binary is available at **quay.io/markd/listvms**. We can start a copy of this container using the deployment yaml file located in the 'listvms/listvms_deployment.yaml' file.

Using the "_oc_" command deploy one instance of the pod, and then check the logs of the pod:

```shell
$ oc create -f listvms/listvms_deployment.yaml
$ oc get pods
NAME                                      READY   STATUS    RESTARTS   AGE
listvms-7b8f865c8d-2zqqn                  1/1     Running   0          7m30s
virt-launcher-vm-fedora-ephemeral-4ljg4   2/2     Running   0          24h
$ oc logs listvms-7b8f865c8d-2zqqn
2021/06/23 18:12:04 cannot obtain KubeVirt vm list: virtualmachines.kubevirt.io is forbidden: User "system:serviceaccount:myvms:default" cannot list resource "virtualmachines" in API group "kubevirt.io" in the namespace "myvms"`
```

> **NOTE:** Be sure to deploy this demo application in a namespace that contains at least one running VM or VMI.

The application is unable to run the operation, because it is running as the default service account in the "_myvms_" namespace. If you remember previously we created a service account in this namespace called "mykubevirtrunner". We need only update the deployment to use this service account and we should see some success. Use the "oc edit" command to update the container spec to include the "serviceAccount: mykubevirtrunner" line as show below:

```yaml
    spec:
      containers:
        - name: listvms
          image: quay.io/markd/listvms
      serviceAccount: mykubevirtrunner
      securityContext: {}
      schedulerName: default-scheduler
```

This change will trigger Kubernetes to redeploy your pod, using the new serviceAccount. We should now see some output from our program:

```shell
$ oc get pods
NAME                                      READY   STATUS    RESTARTS   AGE
listvms-7b8f865c8d-2qzzn                  1/1     Running   0          7m30s
virt-launcher-vm-fedora-ephemeral-4ljg4   2/2     Running   0          24h
$ oc logs listvms-7b8f865c8d-2qzzn
Type                       Name                    Namespace     Status
VirtualMachineInstance     vm-fedora-ephemeral     myvms         Running
awaiting signal
```

## Extending RBAC Role across Namespaces

As currently configured, the mykubevirtrunner service account can only "view" KubeVirt resources within its own namespace. If we want to extend that ability to other namespaces, we can add add the view role for other namespaces to the mykubevirtrunner serviceAccount.

```shell
$ oc new-project myvms2
$ <launch an addition vm here>
$ oc policy add-role-to-user kubevirt.io:view system:serviceaccount:myvms:mykubevirtrunner -n myvms2
```

> **NOTE** The listvms demo application does not support listing of vm instances from multiple namespaces by default. The output from the listvms command will not change at this time.

## Creating Custom RBAC Roles

In this demo we use a built in RBAC role called "view", which allows the user/account access to "GET" most objects at a namespace level. What if we wanted to be able to PUT or create new objects (including creating new machines). There are two ways to accomplish this, the quickest would be to give the serviceAccount the "edit" role, which would allow the service account to create and edit most objects within the namespace. However you could also create a new custom role that limits the serviceAccount to only editing VM and VMI instances.

This can be done by creating a custom RBAC Role as described in the KubeVirt documentation [Creating Customer RBAC Roles](https://kubevirt.io/user-guide/operations/authorization/#creating-custom-rbac-roles)


## References

[KubeVirt Client Go](https://github.com/kubevirt/client-go)
[KubeVirt API Access Control](https://kubevirt.io/2018/KubeVirt-API-Access-Control.html)
[KubeVirt Default RBAC Cluster Roles](https://kubevirt.io/user-guide/operations/authorization/)

No service account key rotation process: https://github.com/kubernetes/kubernetes/issues/20165
