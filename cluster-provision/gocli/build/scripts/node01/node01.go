package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"time"
)

func main() {
	kubeadmConfigPath := "/etc/kubernetes/kubeadm.conf"
	cniManifest := "/provision/cni.yaml"

	singleStack := "/home/vagrant/single_stack"
	_, err := os.Stat(singleStack)

	if err == nil {
		kubeadmConfigPath = "/etc/kubernetes/kubeadm_ipv6.conf"
		cniManifest = "/provision/cni_ipv6.yaml"
	}

	auditEnabled := "/home/vagrant/enable_audit"
	_, err = os.Stat(auditEnabled)

	if err != nil {
		file, err := os.Open("/etc/kubernetes/audit/adv-audit.yaml")
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(file)

		if scanner.Scan() {
			apiVersion := scanner.Text()
			auditConf, err := os.Create("/etc/kubernetes/audit/adv-audit.yaml")
			extraContent := `kind: Policy
rules:
	- level: Metadata`
			if err != nil {
				panic(err)
			}

			_, err = auditConf.WriteString(apiVersion)
			if err != nil {
				panic(err)
			}

			_, err = auditConf.WriteString(extraContent)
			if err != nil {
				panic(err)
			}
		} else {
			log.Fatal("File is empty")
		}
	}

	maxRetries := 10
	var hostname string
	for i := 0; i < maxRetries; i++ {
		if hostname != "" {
			break
		}
		hostname, err = runCMD("hostnamectl --transient")
		if err != nil {
			panic(err)
		}
		fmt.Println("transient hostname isn't ready yet, sleeping for 5 seconds")
		time.Sleep(time.Second * 5)
	}
	cgroupFile := "/sys/fs/cgroup/cgroup.controllers"

	_, err = os.Stat(cgroupFile)
	if err != nil {
		crioConfigDir := "/etc/crio/crio.conf.d"
		err = os.Mkdir(crioConfigDir, 0755)
		if err != nil {
			panic(err)
		}
		cgroupConf, err := os.Create("00-cgroupv2.conf")
		if err != nil {
			panic(err)
		}
		config := `[crio.runtime]
conmon_cgroup = "pod"
cgroup_manager = "cgroupfs"`
		_, err = cgroupConf.WriteString(config)
		if err != nil {
			panic(err)
		}
	}

	kubeletSysconf, err := os.ReadFile("/etc/sysconfig/kubelet")
	pattern := regexp.MustCompile(`--cgroup-driver=systemd`)

	modifiedContent := pattern.ReplaceAllString(string(kubeletSysconf), "--cgroup-driver=cgroupfs")

	err = os.WriteFile("/etc/sysconfig/kubelet", []byte(modifiedContent), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	var crioActive string
	for {
		if crioActive == "active" {
			break
		}
		crioActive, err = runCMD("systemctl is-active crio")
		if err != nil {
			panic(err)
		}
	}

	for {
		ifaces, err := net.Interfaces()
		if err != nil {
			panic(err)
		}
		var foundIPV6 bool

		for _, iface := range ifaces {
			if iface.Name == "eth0" {
				addrs, err := iface.Addrs()
				if err != nil {
					panic(err)
				}

				for _, addr := range addrs {
					ipNet, ok := addr.(*net.IPNet)
					if !ok || ipNet.IP.IsLoopback() {
						continue
					}

					if ipNet.IP.To4() == nil && ipNet.IP.IsGlobalUnicast() {
						foundIPV6 = true
					}
				}
			}
		}
		time.Sleep(time.Second * 2)
		if foundIPV6 {
			break
		}
	}
	_, err = runCMD("kubeadm init --config " + kubeadmConfigPath + " v5")
	if err != nil {
		panic(err)
	}

}

func runCMD(cmd string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	command := exec.Command(cmd)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf(stderr.String())
	}
	return stdout.String(), nil
}
