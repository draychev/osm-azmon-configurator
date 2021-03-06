# osm-azmon-configurator

This small controller watches for Kubernetes Namespaces annotated for joining an OSM mesh and lists them in a ConfigMap.

Details on the OSM annotation ([pkg/constants/constants.go](https://github.com/openservicemesh/osm/blob/release-v0.8/pkg/constants/constants.go#L97)):
```go
	// OSMKubeResourceMonitorAnnotation is the key of the annotation used to monitor a K8s resource
	OSMKubeResourceMonitorAnnotation = "openservicemesh.io/monitored-by"
```

Details on the ConfigMap created:
  - namespace: same as where OSM Controller resides: `osm-system`
  - name: `azmon-config`

```bash
$ kubectl get ConfigMap -n osm-system azmon-config -o json | jq '.data'
```

```json
{
  "namespaces": "bookstore,bookbuyer,bookthief"
}
```
