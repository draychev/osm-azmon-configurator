#!/bin/bash

set -aueo pipefail

# shellcheck disable=SC1091
source .env

make build
make docker-push

./scripts/create-container-registry-creds.sh "${OSM_NAMESPACE}"

OSM_AZMON_POD_NAME=${OSM_AZMON_POD_NAME:-osm-azmon-configurator}

kubectl delete pod -n "${OSM_NAMESPACE}" "${OSM_AZMON_POD_NAME}" || true

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: ${OSM_AZMON_POD_NAME}
  namespace: ${OSM_NAMESPACE}
spec:
  containers:
  - name: osm-azmon-configurator
    image: "${CTR_REGISTRY}/${OSM_AZMON_POD_NAME}:${CTR_TAG}"
    command: ["/osm-azmon-configurator"]
    args:
      - "--verbosity"
      - "trace"
      - "--mesh-name"
      - "osm"
      - "--osm-namespace"
      - "osm-system"
      - "--azmon-configmap-name"
      - "azmon-config"

    env:
      - name: "OSM_HUMAN_DEBUG_LOG"
        value: "true"

  serviceAccount: osm
  serviceAccountName: osm

  imagePullSecrets:
    - name: "$CTR_REGISTRY_CREDS_NAME"
EOF


echo -e "\n\nkubectl get pod -n ${OSM_NAMESPACE} ${OSM_AZMON_POD_NAME}\n\n"
echo -e "\n\nkubectl describe pod -n ${OSM_NAMESPACE} ${OSM_AZMON_POD_NAME}\n\n"
