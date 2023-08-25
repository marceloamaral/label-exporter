# label-exporter
Exports pod labels as prometheus metrics to facilitate metric aggregation.


## How it works
Configure the comma separed list of label prefix to be exporte with the parameter `--label-prefix`. The default prefixes are `le__` and `l__`.

For example, to export the labels starting with `application.label` or `application.process`, configure the label-exporter command with `--label-prefix="application.label,application.process"`, updating the file: `manifest/exporter.yaml`.

The main reason to configure label prefixes is to minimize the metric cardinality and avoid prometheus overhead. But, for some use cases, the user might want to expose all pod labels of all pods in the system. To enable that, use the paramenter `--expose-all=true`.

The prometheus metric will be: `pod_label_exporter(pod_name, pod_namespace, ...)`

## How to build
To create a docker container just do `make build_image`

## Running in a Kind cluster
You will need to load the image with `./script/load-images-to-kind.sh`

## Deploing
The pre-requisit is to have prometheus already installed.
`kubectl create -f manifest/`

