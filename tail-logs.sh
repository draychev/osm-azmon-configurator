#!/bin/bash

set -aueo pipefail

# shellcheck disable=SC1091
source .env

OSM_AZMON_POD_NAME=${OSM_AZMON_POD_NAME:-osm-azmon-configurator}

kubectl logs -n "${OSM_NAMESPACE}" "${OSM_AZMON_POD_NAME}" -f
