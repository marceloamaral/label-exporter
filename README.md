# label-exporter
Exports pod labels as prometheus metrics to facilitate metric aggregation.


## How it works
To minimize the prometheus metric cardinality, the label-exporter only exports labels of pods that contains the label: `label-exporter:`.
Additionally, only the labels with the prefix `le__` will be exported, where the prefix is a configurable parameter when running the label-exporter.

The prometheus metric will be: `label_exporter(pod_name, pod_namespace, ...)`

## How to build
To create a docker container just do `make build_image`

## Running in a Kind cluster
You will need to load the image with `./script/load-images-to-kind.sh`

## Deploing
The pre-requisit is to have prometheus already installed.
`kubectl create -f manifest/`

