[![Licence](https://img.shields.io/badge/licence-Apache%202.0-green)]()
[![Helm](https://img.shields.io/badge/release-0.1.0-brightgreen)]()

# helm-adopt

## Overview 

`helm-adopt` is a helm plugin to adopt existing k8s resources into a new generated helm chart, the idea behind the plugin
was inspired by issue [2730](https://github.com/helm/helm/issues/2730). 

the `adopt` plugin allows
you to :
* adopt existing k8s resources by generating a helm chart 
* migrate adopted resources to be controlled by helm  
* create a helm release using secret as driver
* perform a dry-run/debug (optional)

*Note:* `adopt` does not re-create resources.

## Getting started

### Installation

To install the plugin:
```shell
$ helm plugin install https://github.com/HamzaZo/helm-adopt
```
Update to latest
```shell
$ helm plugin update adopt
```
Install a specific version
```shell
$ helm plugin install https://github.com/HamzaZo/helm-adopt --version 0.1.0
```
You can also verify it's been installed using
```shell
$ helm plugin list
```

### Usage
```
$ helm adopt
Adopt k8s resources into a new helm chart. It's expected to match plural resources kinds.

Usage:
  adopt [command]

Available Commands:
  help        Help about any command
  resources   adopt k8s resources into a new generated helm chart

Flags:
  -h, --help   help for adopt

Use "adopt [command] --help" for more information about a command.

```

#### adopt k8s resources
```
Adopt k8s resources into a new generated helm chart 

Examples:
        
    $ helm adopt resources deployments:nginx services:my-svc -o/--output frontend

    $ helm adopt resources deployments:nginx clusterrolebindings:binding-rbac -o/--output frontend -n/--namespace <ns>

    $ helm adopt resources statefulsets:nginx services:my-svc -r/--release RELEASE-NAME -o/--output frontend -c/--kube-context <ctx>

    $ helm adopt resources deployments:nginx services:my-svc -r/--release RELEASE-NAME -o/--output frontend -k/--kubeconfig <kcfg>

Usage:
  adopt resources <pluralKind>:<name> [flags]

Flags:
      --debug                 Show the generated manifests on STDOUT
      --dry-run               Print what resources will be adopted 
  -h, --help                  help for resources
  -c, --kube-context string   name of the kubeconfig context to use
  -k, --kubeconfig string     path to the kubeconfig file
  -n, --namespace string      namespace scope for this request
  -o, --output string         Specify the chart directory of loaded yaml files
  -r, --release string        Specify the name for the generated release


```

#### Example User case
In the following example the `adopt` plugin will :
* Generate a chart directory called `example`
* adopt `frontend`,`backend` deployments and generate `deployments-0.yaml` `deployments-1.yaml` files.
* adopt `frontend`,`backend` services and generate `services-0.yaml` `services-1.yaml` files.
* adopt `default-lr` limitranges and generate `limitranges-0.yaml` file.  
* migrate adopted resources control to `Helm`  
* create `example` release with version `1` using secret as driver.

```
$ helm adopt resources deployments:frontend,backend services:frontend,backend limitranges:default-lr --output example --release example
INFO[0000] Adopting resources..                         
INFO[0000] Generating chart example                     
INFO[0000] Added resource as file deployments-0 into example chart 
INFO[0000] Added resource as file deployments-1 into example chart 
INFO[0000] Added resource as file limitranges-0 into example chart 
INFO[0000] Added resource as file services-0 into example chart 
INFO[0000] Added resource as file services-1 into example chart 
INFO[0000] Chart example is released as example.1 
```

*Note:* You must use a semicolon after each name `<pluralKind>:<name>,<name>,<name>` 
if you need to adopt a bunch of resources under a specific `pluralKind` as shown in the example above.

check the release status 
```
$ helm ls
NAME    NAMESPACE       REVISION        UPDATED                                 STATUS          CHART           APP VERSION
example default         1               2021-05-18 12:04:53.171943 +0200 CEST   deployed        example-0.1.0   0.1.0       
```

```
$ kubectl get secret
NAME                            TYPE                                  DATA   AGE
sh.helm.release.v1.example.v1   helm.sh/release.v1                    1      51s

```

check generated chart directory
```
$ tree example 
example
├── Chart.yaml
├── charts
├── templates
│   ├── _helpers.tpl
│   ├── deployments-0.yaml
│   ├── deployments-1.yaml
│   ├── limitranges-0.yaml
│   ├── services-0.yaml
│   └── services-1.yaml
└── values.yaml

2 directories, 8 files


```

*Hints:* To find out the plural resource for a specific kind use `kubectl api-resources`