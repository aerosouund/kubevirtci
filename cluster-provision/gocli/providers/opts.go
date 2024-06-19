package providers

func WithNodes(nodes interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Nodes = nodes.(uint)
	}
}

func WithNuma(numa interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Numa = numa.(uint)
	}
}

func WithMemory(memory interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Memory = memory.(string)
	}
}

func WithCPU(cpu interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.CPU = cpu.(uint)
	}
}

func WithSecondaryNics(secondaryNics interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SecondaryNics = secondaryNics.(uint)
	}
}

func WithQemuArgs(qemuArgs interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.QemuArgs = qemuArgs.(string)
	}
}

func WithKernelArgs(kernelArgs interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.KernelArgs = kernelArgs.(string)
	}
}

func WithBackground(background interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Background = background.(bool)
	}
}

func WithReverse(reverse interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Reverse = reverse.(bool)
	}
}

func WithRandomPorts(randomPorts interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RandomPorts = randomPorts.(bool)
	}
}

func WithSlim(slim interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Slim = slim.(bool)
	}
}

func WithVNCPort(vncPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.VNCPort = uint16(vncPort.(uint))
	}
}

func WithHTTPPort(httpPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.HTTPPort = uint16(httpPort.(uint))
	}
}

func WithHTTPSPort(httpsPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.HTTPSPort = uint16(httpsPort.(uint))
	}
}

func WithRegistryPort(registryPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RegistryPort = uint16(registryPort.(uint))
	}
}

func WithOCPort(ocpPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.OCPort = uint16(ocpPort.(uint))
	}
}

func WithK8sPort(k8sPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.K8sPort = uint16(k8sPort.(uint))
	}
}

func WithSSHPort(sshPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SSHPort = uint16(sshPort.(uint))
	}
}

func WithPrometheusPort(prometheusPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.PrometheusPort = uint16(prometheusPort.(uint))
	}
}

func WithGrafanaPort(grafanaPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.GrafanaPort = uint16(grafanaPort.(uint))
	}
}

func WithDNSPort(dnsPort interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.DNSPort = uint16(dnsPort.(uint))
	}
}

func WithNFSData(nfsData interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.NFSData = nfsData.(string)
	}
}

func WithEnableCeph(enableCeph interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableCeph = enableCeph.(bool)
	}
}

func WithEnableIstio(enableIstio interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableIstio = enableIstio.(bool)
	}
}

func WithEnableCNAO(enableCNAO interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableCNAO = enableCNAO.(bool)
	}
}

func WithEnableNFSCSI(enableNFSCSI interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableNFSCSI = enableNFSCSI.(bool)
	}
}

func WithEnablePrometheus(enablePrometheus interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePrometheus = enablePrometheus.(bool)
	}
}

func WithEnablePrometheusAlertManager(enablePrometheusAlertManager interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePrometheusAlertManager = enablePrometheusAlertManager.(bool)
	}
}

func WithEnableGrafana(enableGrafana interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableGrafana = enableGrafana.(bool)
	}
}

func WithDockerProxy(dockerProxy interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.DockerProxy = dockerProxy.(string)
	}
}

func WithGPU(gpu interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.GPU = gpu.(string)
	}
}

func WithNvmeDisks(nvmeDisks interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.NvmeDisks = nvmeDisks.([]string)
	}
}

func WithScsiDisks(scsiDisks interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.ScsiDisks = scsiDisks.([]string)
	}
}

func WithRunEtcdOnMemory(runEtcdOnMemory interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.RunEtcdOnMemory = runEtcdOnMemory.(bool)
	}
}

func WithEtcdCapacity(etcdCapacity interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EtcdCapacity = etcdCapacity.(string)
	}
}

func WithHugepages2M(hugepages2M interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.Hugepages2M = hugepages2M.(uint)
	}
}

func WithEnableRealtimeScheduler(enableRealtimeScheduler interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableRealtimeScheduler = enableRealtimeScheduler.(bool)
	}
}

func WithEnableFIPS(enableFIPS interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableFIPS = enableFIPS.(bool)
	}
}

func WithEnablePSA(enablePSA interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnablePSA = enablePSA.(bool)
	}
}

func WithSingleStack(singleStack interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.SingleStack = singleStack.(bool)
	}
}

func WithEnableAudit(enableAudit interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.EnableAudit = enableAudit.(bool)
	}
}

func WithUSBDisks(usbDisks interface{}) KubevirtProviderOption {
	return func(c *KubevirtProvider) {
		c.USBDisks = usbDisks.([]string)
	}
}
