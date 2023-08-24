# label-exporter
Exports pod labels as prometheus metrics to facilitate metric aggregation.


## How it works
To minimize the prometheus metric cardinality, the label-exporter only exports labels of pods that contains the label: `label-exporter:`.
Additionally, only the labels with the prefix `le__` will be exported, where the prefix is a configurable parameter when running the label-exporter.

To change the label prefix, for example to `application.label`, you need to update the `manifest/exporter.yaml` file with `--label-prefix="application.label"`.

For some use cases, the user might want to expose all pod labels of all pods in the system. To enable that, use the paramenter `--expose-all=true`, updating the `manifest/exporter.yaml` file.

The prometheus metric will be: `label_exporter(pod_name, pod_namespace, ...)`

## How to build
To create a docker container just do `make build_image`

## Running in a Kind cluster
You will need to load the image with `./script/load-images-to-kind.sh`

## Deploing
The pre-requisit is to have prometheus already installed.
`kubectl create -f manifest/`

