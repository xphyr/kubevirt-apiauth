# Authentication methods with kubevirt Go Library

## Introduction

We will use a simple application to demonstrate the different authentication methoods. We will run the app both inside the cluster and external to the cluster.

## Setup

We will be testing that we have properly authenticated to the system by listing out the VMI and VM instances in your cluster that you have access too. If you do not have any running VMs in your cluster, create a new project and deploy a virtual machine.

# Running externally leveraging a kubeconfig file

Generate a kuebconfig from the oc command:

```
$ export KUBECONFIG=$(pwd)/kubeconfig
$ oc login https://api.<clustername>.example.com
The server uses a certificate signed by an unknown authority.
You can bypass the certificate check, but any data you send to the server could be intercepted by others.
Use insecure connections? (y/n): y

Authentication required for https://api.ocp47.xphyrlab.net:6443 (openshift)
Username: markd
Password:
Login successful.
```
This has created a kubeconfig in your current working directory called "kubeconfig", however it is only good for a short period of time and will time out.

If you want a long running password that will not expire, we are going to want to create a service account, and get the authtoken for that account. Start by logging into your cluster as an administrator and create a new namespace. We will be using this namespace for our service account.

```
unset KUBECONFIG
$ oc login
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
$ oc describe secret mykubevirtrunner-token-qfkmc -n myvms
Name:         mykubevirtrunner-token-qfkmc
Namespace:    myvms
Labels:       <none>
Annotations:  kubernetes.io/service-account.name: mykubevirtrunner
              kubernetes.io/service-account.uid: a02b25e8-39bf-4e4b-8680-918521dc0a6b

Type:  kubernetes.io/service-account-token

Data
====
token:       eyJhbGciOiJSUzI1NiIsImtpZCI6InE0UjNxaS1mY190RXYyTXB5OHNTUE5fTlp6bHdPVWl1WmQ3U3VQajBDTW8ifQeyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJteXZtcyIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJteWt1YmV2aXJ0cnVubmVyLXRva2VuLXFma21jIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Im15a3ViZXZpcnRydW5uZXIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJhMDJiMjVlOC0zOWJmLTRlNGItODY4MC05MTg1MjFkYzBhNmIiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6bXl2bXM6bXlrdWJldmlydHJ1bm5lciJ9N5FHuY5V7I-71AS9P9qRdm4y6ZKVR0X8ZTndNbHVIvPFvGW58kqvpZEhIlz4oUYVcSBeHiFGaJU_xqVqQJNm5C5axM4E1gqUpMbLgm61ha3C47c7IziMXhofzDidggKo0DbofYlkJdU4iuk45llf5m-Eg1dUVwGVXEOUmuQB5G7blim6iqiyd3tNVZP2LnziPycxFpMpBLiwLqeQFiQddE3c3z4_6CoFLcFy1RsCt2Bkby3r-NhMsDxvPlGrTf92C7wHYcgCVnGZ8-OgNMnhkbcUjZgPPmygbkbbtuqi6Rqq0xDnuVoxfrCkasdknAVf--2Pq45scEXwr9OHqiIc-A
ca.crt:          7230 bytes
namespace:       5 bytes
service-ca.crt:  8443 bytes
$ export KUBECONFIG=$(pwd)/kubeconfig_token
$ oc login https://api.<clustername>.example.com --token=eyJhbGciOiJSUzI1NiIsImtpZCI6InE0UjNxaS1mY190RXYyTXB5OHNTUE5fTlp6bHdPVWl1WmQ3U3VQajBDTW8ifQeyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJteXZtcyIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJteWt1YmV2aXJ0cnVubmVyLXRva2VuLXFma21jIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Im15a3ViZXZpcnRydW5uZXIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJhMDJiMjVlOC0zOWJmLTRlNGItODY4MC05MTg1MjFkYzBhNmIiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6bXl2bXM6bXlrdWJldmlydHJ1bm5lciJ9N5FHuY5V7I-71AS9P9qRdm4y6ZKVR0X8ZTndNbHVIvPFvGW58kqvpZEhIlz4oUYVcSBeHiFGaJU_xqVqQJNm5C5axM4E1gqUpMbLgm61ha3C47c7IziMXhofzDidggKo0DbofYlkJdU4iuk45llf5m-Eg1dUVwGVXEOUmuQB5G7blim6iqiyd3tNVZP2LnziPycxFpMpBLiwLqeQFiQddE3c3z4_6CoFLcFy1RsCt2Bkby3r-NhMsDxvPlGrTf92C7wHYcgCVnGZ8-OgNMnhkbcUjZgPPmygbkbbtuqi6Rqq0xDnuVoxfrCkasdknAVf--2Pq45scEXwr9OHqiIc-A
```

Now run our test program:
```
$ listvms/listvms
get error:  `2021/06/23 18:12:04 cannot obtain KubeVirt vm list: virtualmachines.kubevirt.io is forbidden: User "system:serviceaccount:myvms:mykubevirtrunner" cannot list resource "virtualmachines" in API group "kubevirt.io" in the namespace "myvms"`
```

You can see we are now using our service account, but that service account doenst have the right permissions... 

```
$ oc policy add-role-to-user view system:serviceaccount:myvms:mykubevirtrunner
```

Now run the listvms command again:

```
./listvms/listvms
Type                       Name                    Namespace     Status
VirtualMachineInstance     vm-fedora-ephemeral     myvms         Running
awaiting signal
^C
interrupt
exiting
```

Start by compiling our test program:

```
$ cd listvms
$ make listvms

# Running in K8s:

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
