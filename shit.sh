kubeadmn_patches_path="cluster-provision/gocli/opts/k8sprovision/conf/patches"
mkdir -p $kubeadmn_patches_path

cat >$kubeadmn_patches_path/kube-apiserver.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/kube-controller-manager.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/kube-scheduler.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/etcd.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF

cat >$kubeadmn_patches_path/add-security-context-deployment-patch.yaml <<EOF
spec:
  template:
    spec:
      securityContext:
        seLinuxOptions:
          type: spc_t
EOF
