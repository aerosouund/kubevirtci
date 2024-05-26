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

	maxRetries := 4
	var hostname string
	for i := 0; i < maxRetries; i++ {
		if hostname != "" {
			break
		}
		hostname, _ = runCMD("hostnamectl --transient", true)
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
	for i := 0; i < maxRetries; i++ {
		if crioActive == "active" {
			break
		}
		crioActive, err = runCMD("systemctl is-active crio", true)
		fmt.Println("crio status:", crioActive)
		time.Sleep(time.Second * 3)
		// if err != nil {
		// 	panic(err)
		// }
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
		if foundIPV6 {
			break
		}
		fmt.Println("ipv6 not ready, sleeping for 2 seconds")
		time.Sleep(time.Second * 2)
	}

	err = os.Setenv("KUBECONFIG", "/etc/kubernetes/admin.conf")
	if err != nil {
		panic(err)
	}

	cmds := []string{
		"kubeadm init --config " + kubeadmConfigPath + " -v5",
		`kubectl patch deployment coredns -n kube-system -p "$(cat /provision/kubeadm-patches/add-security-context-deployment-patch.yaml)"`,
		"kubectl create -f " + cniManifest,
		"kubectl taint nodes node01 node-role.kubernetes.io/control-plane:NoSchedule-",
		"kubectl create -f /provision/local-volume.yaml",
		"mkdir -p /var/lib/rook",
		"chcon -t container_file_t /var/lib/rook",
	}

	for _, cmd := range cmds {
		_, err = runCMD(cmd, true)
		if err != nil {
			panic(err)
		}
	}
}

func runCMD(cmd string, stdOut bool) (string, error) {
	var stdout, stderr bytes.Buffer

	command := exec.Command("bash", "-c", cmd)
	if !stdOut {
		command.Stdout = &stdout
		command.Stderr = &stderr
	}

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf(stderr.String())
	}
	if stdOut {
		return "", nil
	}
	return stdout.String(), nil
}
