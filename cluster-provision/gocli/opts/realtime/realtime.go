package realtime

import utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"

type RealtimeOpt struct {
	sshClient utils.SSHClient
}

func NewRealtimeOpt(sc utils.SSHClient) *RealtimeOpt {
	return &RealtimeOpt{
		sshClient: sc,
	}
}

func (o *RealtimeOpt) Exec() error {
	cmds := []string{
		"echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf",
		"sysctl --system",
	}

	for _, cmd := range cmds {
		if _, err := o.sshClient.SSH(cmd, true); err != nil {
			return err
		}
	}
	return nil
}
