# Authentication methods with kubevirt client libraries 

## Introduction

We will use a simple application to demonstrate how the kubevirt client library authenticates with your k8s cluster. We will use the example application in the "client-go" library with a small modification to it, to allow for running both locally and within in the cluster. This tutorial assumes you have some knowledge of Go, and is not meant to be a Go training doc.

## Requirements

In order to run this application locally you will need to have the Go programming language installed on your machine. The steps listed here were tested with Go version 1.16.
You will also need to have admin rights in your k8s cluster. We will be testing that we have properly authenticated to the system by listing out the VMI and VM instances in your cluster that you have access too. If you do not have any running VMs in your cluster, create a new project and deploy a virtual machine. Finally we will be leveraging an OpenShift cluster for this demo, and will be using the "oc" command for our command line interactions with the cluster.

## Setup

### Compiling our test application

Start by cloning this repo and compiling our test applicaiton:

```
$ git clone https://github.com/xphyr/kubevirt-apiauth.git
$ cd kubevirt-apiauth
$ cd listvms
$ go build
```

Test and ensure that the applicaiton compiled correctly. If you have a working k8s context, running this command may return some values. If you are not logged in, or do not have a current context, you will get an error. This is OK, we will discuss authentication next.:

```
$ ./listvms
2021/06/23 16:51:28 cannot obtain KubeVirt vm list: Get "http://localhost:8080/apis/kubevirt.io/v1alpha3/namespaces/default/virtualmachines": dial tcp 127.0.0.1:8080: connect: connection refused
```

Now that we have a working test application, we will move onto the next step, authentication.

## Running our application externally leveraging a kubeconfig file

The default authentication file for Kubernetes is the [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file. We will not be going into details of this file, but you can click the link to goto the documentation on the kubeconfig file to learn more about it.

### Using the default kubeconfig

If you havent already done so, log into your OpenShift cluster with the oc command:

```
$ oc login https://api.<clustername>.<basedomain> -u <userName>
password: ********
Login successful.

You have access to 80 projects, the list has been suppressed. You can list all projects with ' projects'

Using project "myvms".
```

We now have a valid kubeconfig. This file by default is stored in your home directory at `~/.kube/config`. You should now be able to run our test application and get some results (assuming you have some running vms in your cluster)

```
$ ./listvms/listvms
./listvms/listvms
Type                       Name                    Namespace     Status
VirtualMachineInstance     vm-fedora-ephemeral     myvms         Running
```

This is great, but there is an issue. The authentication methood we used (logging in with a named user/password) will time out after 24 hours. This wont work for a long running application. You don't want to have to keep logging in every 23.5 hours to keep your application up and running. How do we handle this?  We create a ServiceAccount within our k8s cluster and authenticate as the service account.

[Service Accounts](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/) are accounts for processes as opposed to users. By default they are scoped to a namespace, but you can give service accounts access to other namepsaces through RBAC rules that we will discuss later. In this demo, we will be using the "myvms" project/namespace, so the service account we create will be initially scoped only to this namespace.

Start by creating a new service account called "mykubevirtrunner":

```
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

```
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

The data listed for the "token" key is the information we will use in the next step.

### Creating a kubeconfig for the service account

We will create a new kubeconfig file that leverages the service account and token we just created. The easiest way to do this is to create an empty kubeconfig file, and use the oc command to log in with the new token. We will start by setting the KUBECONFIG environment variable to point to a file in our local directory, and then using the oc command loginto the cluster:

```
$ export KUBECONFIG=$(pwd)/sa-kubeconfig
$ oc login https://api.<clustername>.example.com --token=<paste token from last step here>
$ oc whoami
system:serviceaccount:myvms:mykubevirtrunner
```

We now have a kubeconfig that authenticates us as the service account that we created in the last section. Try running our test program again:

```
$ listvms/listvms
get error:  `2021/06/23 18:12:04 cannot obtain KubeVirt vm list: virtualmachines.kubevirt.io is forbidden: User "system:serviceaccount:myvms:mykubevirtrunner" cannot list resource "virtualmachines" in API group "kubevirt.io" in the namespace "myvms"`
```

You can see we are now using our service account, but that service account doenst have the right permissions... We will start simple and give the service account the "view" role, which will allow the service account to see the objects within the myvms namespace:

```
$ oc policy add-role-to-user view system:serviceaccount:myvms:mykubevirtrunner
```

Now run the listvms command again:

```
./listvms/listvms
Type                       Name                    Namespace     Status
VirtualMachineInstance     vm-fedora-ephemeral     myvms         Running
```

Success! Our application is now using the service account that we created for authentication to the cluster. The service account can be extended by adding additional default roles to the account, or by creating custom roles that limit the scope of the service account to only the exact actions you want to take.

# Running in K8s:

So all of this is great if you want to run the application outside of your cluster ... but what if you want your application to run INSIDE you cluster. You could create a kubeconfig file, and add it to your namespace as a secret and then mount that secret as a volume inside your pod, but there is an easier way that continues to leverage the service account that we created. 

get error:  `2021/06/23 18:12:04 cannot obtain KubeVirt vm list: virtualmachines.kubevirt.io is forbidden: User "system:serviceaccount:myvms:default" cannot list resource "virtualmachines" in API group "kubevirt.io" in the namespace "myvms"`

Need to fix permissions/roles

start by adding a service account in your project 

`$ oc create sa mykubevirtrunner`


We will start by giving "view" permissions to our serviceAccount:

$ oc policy add-role-to-user view system:serviceaccount:myvms:mykubevirtrunner

Now update our deployment to use this account for deployment:

Find the spec section for your container and add `serviceAccount: mykubevirtrunner`

Start your pod back up and you will now see that it can see the pods in your namespace:

```
Type                       Name                    Namespace     Status
VirtualMachineInstance     vm-fedora-ephemeral     myvms         Running
awaiting signal
```




# References

https://github.com/kubevirt/client-go

No service account key rotation process: https://github.com/kubernetes/kubernetes/issues/20165
