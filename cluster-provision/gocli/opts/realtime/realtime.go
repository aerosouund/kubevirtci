package realtime

import utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"

type RealtimeOpt struct {
	sshPort uint16
}

func NewRealtomeOpt(sshPort uint16) *RealtimeOpt {
	return &RealtimeOpt{
		sshPort: sshPort,
	}
}

func (o *RealtimeOpt) Exec() error {
	cmds := []string{
		"sudo echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf",
		"sudo sysctl --system",
	}

	for _, cmd := range cmds {
		if _, err := utils.JumpSSH(o.sshPort, 1, cmd, true); err != nil {
			return err
		}
	}
	return nil
}
