#!/bin/bash

docker pull quay.io/sustainability/label-exporter:v0.1

kind load docker-image quay.io/sustainability/label-exporter:v0.1 
