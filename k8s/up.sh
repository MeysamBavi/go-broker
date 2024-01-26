#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

STORE=$1
if [ STORE = "" ]; then
    echo "no argument for store"
    exit 1
fi

eval $(minikube -p minikube docker-env)
docker build -t go-broker ./..

kubectl create configmap broker-configmap \
       --from-env-file="./broker-with-$STORE.env" || :

kubectl create configmap envoy-configmap \
      --from-file="../envoy.yaml" || :

kubectl create configmap prometheus-configmap \
      --from-file="../prometheus.yml" || :

kubectl create configmap grafana-dashboards \
      --from-file="../grafana" || :
      
for obj in "$STORE" "envoy" "broker" "prometheus" "grafana" "jaeger" ; do
    if [ $obj = "memory" ]; then
        continue
    fi
    kubectl create -f "$obj-deployment.yaml"
    kubectl create -f "$obj-service.yaml"
done
