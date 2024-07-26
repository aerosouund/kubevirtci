package prometheus

import (
	"embed"
	"io/fs"
	"path/filepath"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type PrometheusOpt struct {
	grafanaEnabled      bool
	alertmanagerEnabled bool

	client k8s.K8sDynamicClient
}

func NewPrometheusOpt(c k8s.K8sDynamicClient, grafanaEnabled, alertmanagerEnabled bool) *PrometheusOpt {
	return &PrometheusOpt{
		grafanaEnabled:      grafanaEnabled,
		alertmanagerEnabled: alertmanagerEnabled,
		client:              c,
	}
}

func (o *PrometheusOpt) Exec() error {
	for _, dir := range []string{"prometheus-operator", "prometheus", "monitors", "kube-state-metrics", "node-exporter"} {
		err := fs.WalkDir(f, "manifests/"+dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".yaml" {
				yamlData, err := f.ReadFile(path)
				if err != nil {
					return err
				}
				if err := o.client.Apply(yamlData); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	if o.alertmanagerEnabled {
		for _, dir := range []string{"alertmanager", "alertmanager-rules"} {
			err := fs.WalkDir(f, "manifests/"+dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && filepath.Ext(path) == ".yaml" {
					yamlData, err := f.ReadFile(path)
					if err != nil {
						return err
					}
					if err := o.client.Apply(yamlData); err != nil {
						return err
					}
				}
				return nil
			})

			if err != nil {
				return err
			}
		}
	}

	if o.grafanaEnabled {
		err := fs.WalkDir(f, "manifests/grafana", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".yaml" {
				yamlData, err := f.ReadFile(path)
				if err != nil {
					return err
				}
				if err := o.client.Apply(yamlData); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
