#!/bin/sh

os=$(go env GOOS)
arch=$(go env GOARCH)

# Kubebuilder
curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/
sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder
export PATH=$PATH:/usr/local/kubebuilder/bin

# Kustomize
curl -L https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.8.0/kustomize_v3.8.0_${os}_${arch}.tar.gz | tar -xz -C /tmp/
sudo mv /tmp/kustomize /usr/local/bin/kustomize
