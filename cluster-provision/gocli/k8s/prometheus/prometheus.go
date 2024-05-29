package prometheus

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"
)

//go:embed manifests/*
var f embed.FS

type PrometheusOpt struct {
	grafanaEnabled      bool
	alertmanagerEnabled bool

	client *k8s.K8sDynamicClient
}

func NewPrometheusOpt(c *k8s.K8sDynamicClient, grafanaEnabled, alertmanagerEnabled bool) *PrometheusOpt {
	return &PrometheusOpt{
		grafanaEnabled:      grafanaEnabled,
		alertmanagerEnabled: alertmanagerEnabled,
		client:              c,
	}
}

func (o *PrometheusOpt) Exec() error {
	defaultManifests := []string{
		"prometheus-operator/0namespace-namespace.yaml",
		"prometheus-operator/prometheus-operator-clusterRoleBinding.yaml",
		"prometheus-operator/prometheus-operator-clusterRole.yaml",
		"prometheus-operator/prometheus-operator-serviceAccount.yaml",
		"prometheus-operator/prometheus-operator-0prometheusCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-0podmonitorCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-0probeCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-0thanosrulerCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-0alertmanagerCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-0alertmanagerConfigCustomResourceDefinition.yaml",
		"prometheus-operator/prometheus-operator-service.yaml",
		"prometheus-operator/prometheus-operator-deployment.yaml",
		"prometheus/prometheus-clusterRole.yaml",
		"prometheus/prometheus-clusterRoleBinding.yaml",
		"prometheus/prometheus-roleBindingConfig.yaml",
		"prometheus/prometheus-roleBindingSpecificNamespaces.yaml",
		"prometheus/prometheus-roleConfig.yaml",
		"prometheus/prometheus-roleSpecificNamespaces.yaml",
		"prometheus/prometheus-serviceAccount.yaml",
		"prometheus/prometheus-podDisruptionBudget.yaml",
		"prometheus/prometheus-service.yaml",
		"prometheus/prometheus-prometheus.yaml",
		"monitors/kubernetes-serviceMonitorApiserver.yaml",
		"monitors/kubernetes-serviceMonitorCoreDNS.yaml",
		"monitors/kubernetes-serviceMonitorKubeControllerManager.yaml",
		"monitors/kubernetes-serviceMonitorKubeScheduler.yaml",
		"monitors/kubernetes-serviceMonitorKubelet.yaml",
		"monitors/prometheus-operator-serviceMonitor.yaml",
		"monitors/prometheus-serviceMonitor.yaml",
		"kube-state-metrics/kube-state-metrics-clusterRole.yaml",
		"kube-state-metrics/kube-state-metrics-clusterRoleBinding.yaml",
		"kube-state-metrics/kube-state-metrics-prometheusRule.yaml",
		"kube-state-metrics/kube-state-metrics-serviceAccount.yaml",
		"kube-state-metrics/kube-state-metrics-serviceMonitor.yaml",
		"kube-state-metrics/kube-state-metrics-service.yaml",
		"kube-state-metrics/kube-state-metrics-deployment.yaml",
		"node-exporter/node-exporter-clusterRole.yaml",
		"node-exporter/node-exporter-clusterRoleBinding.yaml",
		"node-exporter/node-exporter-prometheusRule.yaml",
		"node-exporter/node-exporter-serviceAccount.yaml",
		"node-exporter/node-exporter-serviceMonitor.yaml",
		"node-exporter/node-exporter-daemonset.yaml",
		"node-exporter/node-exporter-service.yaml",
	}

	for _, manifest := range defaultManifests {
		err := o.client.Apply(f, manifest)
		if err != nil {
			return err
		}
	}

	if o.alertmanagerEnabled {
		alertmanagerManifests := []string{
			"alertmanager/alertmanager-secret.yaml",
			"alertmanager/alertmanager-serviceAccount.yaml",
			"alertmanager/alertmanager-serviceMonitor.yaml",
			"alertmanager/alertmanager-podDisruptionBudget.yaml",
			"alertmanager/alertmanager-service.yaml",
			"alertmanager/alertmanager-alertmanager.yaml",
			"alertmanager-rules/alertmanager-prometheusRule.yaml",
			"alertmanager-rules/kube-prometheus-prometheusRule.yaml",
			"alertmanager-rules/prometheus-operator-prometheusRule.yaml",
			"alertmanager-rules/prometheus-prometheusRule.yaml",
		}
		for _, manifest := range alertmanagerManifests {
			err := o.client.Apply(f, manifest)
			if err != nil {
				return err
			}
		}
	}

	if o.grafanaEnabled {
		grafanaManifests := []string{
			"grafana/grafana-dashboardDatasources.yaml",
			"grafana/grafana-dashboardDefinitions.yaml",
			"grafana/grafana-dashboardSources.yaml",
			"grafana/grafana-deployment.yaml",
			"grafana/grafana-service.yaml",
			"grafana/grafana-serviceAccount.yaml",
			"grafana/grafana-serviceMonitor.yaml",
		}

		for _, manifest := range grafanaManifests {
			err := o.client.Apply(f, manifest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
