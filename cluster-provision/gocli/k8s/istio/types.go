package k8s

import "k8s.io/client-go/dynamic"

type IstioOpt struct {
	client *dynamic.DynamicClient
}
