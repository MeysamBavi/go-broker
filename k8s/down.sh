#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

STORE=$1
if [ STORE = "" ]; then
    echo "no argument for store"
    exit 1
fi

kubectl delete configmap \
    broker-configmap envoy-configmap prometheus-configmap grafana-dashboards || :

for obj in "$STORE" "envoy" "broker" "prometheus" "grafana" "jaeger" ; do
    if [ $obj = "memory" ]; then
        continue
    fi
    kubectl delete -f "$obj-deployment.yaml" || :
    kubectl delete -f "$obj-service.yaml" || :
done
