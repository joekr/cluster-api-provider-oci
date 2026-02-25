# RDMA Cluster Bring-Up & Validation

This checklist captures everything required to bring up the RDMA test
cluster from this repo’s template, install the RDMA device plugin, and
verify a workload can consume the exposed resources.

## 0. Build / Update the RDMA Worker Image

Before launching the cluster, bake the RDMA tooling into your custom
worker image:

```bash
sudo dnf install -y rdma-core libibverbs-utils infiniband-diags
sudo modprobe mlx5_core
sudo modprobe mlx5_ib
sudo modprobe rdma_ucm ib_uverbs ib_umad
echo -e "mlx5_core\nmlx5_ib" | sudo tee /etc/modules-load.d/rdma.conf
```

Run the above inside your image-builder pipeline (packer, OCI custom
image, etc.), verify `/dev/infiniband` exists, then create a snapshot and
update `OCI_NODE_IMAGE_ID` in the cluster template parameters to use the
new image.

## 1. Create the Cluster (Management Workstation)

> **Prerequisite:** The RDMA worker template assumes a pre-provisioned
> OCI Compute Cluster (HPC) in the target compartment/region. Update the
> `computeClusterId` field in `templates/cluster-template-rdma-test.yaml`
> with the OCID of that cluster before generating manifests.

1. Export the usual cluster parameters (compartment OCID, control plane
   image, worker image, SSH key, etc.).
2. Generate and apply the cluster manifest:

   ```bash
   NAMESPACE=default \
   clusterctl generate cluster ${CLUSTER_NAME} \
     --kubernetes-version ${KUBERNETES_VERSION} \
     --control-plane-machine-count=1 \
     --worker-machine-count=2 \
     --from templates/cluster-template-rdma-test.yaml \
   | kubectl apply -f -
   ```

   > **Note:** Set `KUBERNETES_VERSION` to match the kubelet version baked
   > into the custom images (e.g. `v1.34.1`).

3. Wait for the control plane and workers to join:

   ```bash
   kubectl get nodes
   ```

## 2. Install CNI + CCM

1. Install Calico:

   ```bash
   kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
   ```

2. Install the OCI Cloud Controller Manager (plus CSI if required by the
   test):

   install CCM https://github.com/oracle/oci-cloud-controller-manager

3. Confirm all nodes report `Ready` before proceeding.

## 3. Deploy the RDMA Device Plugin

1. Clone the Mellanox device plugin repo (or download the release tarball):

   ```bash
   git clone https://github.com/Mellanox/k8s-rdma-shared-dev-plugin.git
   cd k8s-rdma-shared-dev-plugin
   ```

2. Update the ConfigMap to select Mellanox devices (copy/paste from our
   template):

   ```bash
   cat <<'EOF' > deployment/k8s/base/config/config.json
   {
     "periodicUpdateInterval": 300,
     "configList": [
       {
         "resourceName": "hca_shared_devices_a",
         "rdmaHcaMax": 1000,
         "selectors": {
           "vendors": ["15b3"]
         }
       }
     ]
   }
   EOF
   ```

3. Deploy the overlay (includes namespace, service account, DaemonSet):

   ```bash
   kubectl apply -k deployment/k8s/overlay
   ```

4. Confirm the DaemonSet is running:

   ```bash
   kubectl get pods -n kube-system | grep rdma-shared-dp
   ```

   You should see one `rdma-shared-dp-ds-xxxxx` pod per worker in
   `Running` state.

## 4. Verify RDMA Resources on Nodes

Check each worker for non-zero RDMA capacity:

```bash
kubectl describe node <worker-name> | grep -A2 rdma/
```

Expect output similar to `rdma/hca_shared_devices_a: 1k`.

Optionally label RDMA-capable nodes for targeted scheduling:

```bash
kubectl label node <worker-name> rdma.networking/ready=true
```

## 6. Run an RDMA Smoke Test Pod

Use the upstream Mellanox validation pod, which runs `ibv_devinfo` and
exercises the device plugin allocation:

```bash
kubectl apply -f https://raw.githubusercontent.com/Mellanox/k8s-rdma-shared-dev-plugin/master/example/test-hca-pod.yaml
```

Watch for completion and review the output:

```bash
kubectl wait --for=condition=Ready pod/test-hca-pod
kubectl logs test-hca-pod
```

The log should include HCA details (`hca_id: mlx5_0`, etc.), confirming
RDMA is available inside the container.

```
crw-------. 1 root root 231,   6 Feb 24 20:53 umad6
crw-------. 1 root root 231,   7 Feb 24 20:53 umad7
crw-------. 1 root root 231,   8 Feb 24 20:53 umad8
crw-------. 1 root root 231,   9 Feb 24 20:53 umad9
crw-rw-rw-. 1 root root 231, 192 Feb 24 20:53 uverbs0
crw-rw-rw-. 1 root root 231, 193 Feb 24 20:53 uverbs1
crw-rw-rw-. 1 root root 231, 202 Feb 24 20:53 uverbs10
crw-rw-rw-. 1 root root 231, 203 Feb 24 20:53 uverbs11
crw-rw-rw-. 1 root root 231, 204 Feb 24 20:53 uverbs12
crw-rw-rw-. 1 root root 231, 205 Feb 24 20:53 uverbs13
crw-rw-rw-. 1 root root 231, 206 Feb 24 20:53 uverbs14
crw-rw-rw-. 1 root root 231, 207 Feb 24 20:53 uverbs15
crw-rw-rw-. 1 root root 231, 208 Feb 24 20:53 uverbs16
crw-rw-rw-. 1 root root 231, 209 Feb 24 20:53 uverbs17
crw-rw-rw-. 1 root root 231, 194 Feb 24 20:53 uverbs2
crw-rw-rw-. 1 root root 231, 195 Feb 24 20:53 uverbs3
crw-rw-rw-. 1 root root 231, 196 Feb 24 20:53 uverbs4
crw-rw-rw-. 1 root root 231, 197 Feb 24 20:53 uverbs5
crw-rw-rw-. 1 root root 231, 198 Feb 24 20:53 uverbs6
crw-rw-rw-. 1 root root 231, 199 Feb 24 20:53 uverbs7
crw-rw-rw-. 1 root root 231, 200 Feb 24 20:53 uverbs8
crw-rw-rw-. 1 root root 231, 201 Feb 24 20:53 uverbs9

/sys/class/infiniband:
total 0
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_0 -> ../../devices/pci0000:3f/0000:3f:01.1/0000:40:00.0/0000:41:00.0/0000:42:00.0/0000:43:10.0/0000:44:00.0/0000:45:00.0/0000:46:00.0/0000:47:00.0/0000:48:00.0/infiniband/mlx5_0
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_1 -> ../../devices/pci0000:3f/0000:3f:01.1/0000:40:00.0/0000:41:00.0/0000:42:00.0/0000:43:10.0/0000:44:00.0/0000:45:00.0/0000:46:00.0/0000:47:00.0/0000:48:00.1/infiniband/mlx5_1
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_10 -> ../../devices/pci0000:ba/0000:ba:01.1/0000:bb:00.0/0000:bc:00.0/0000:bd:00.0/0000:be:10.0/0000:bf:00.0/0000:c0:00.0/0000:c1:00.0/0000:c2:00.0/0000:c3:00.0/infiniband/mlx5_10
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_11 -> ../../devices/pci0000:ba/0000:ba:01.1/0000:bb:00.0/0000:bc:00.0/0000:bd:00.0/0000:be:10.0/0000:bf:00.0/0000:c0:00.0/0000:c1:00.0/0000:c2:00.0/0000:c3:00.1/infiniband/mlx5_11
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_12 -> ../../devices/pci0000:ba/0000:ba:01.1/0000:bb:00.0/0000:bc:00.0/0000:bd:00.0/0000:be:10.0/0000:bf:00.0/0000:c0:04.0/0000:c4:00.0/0000:c5:10.0/0000:d1:00.0/infiniband/mlx5_12
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_13 -> ../../devices/pci0000:ba/0000:ba:01.1/0000:bb:00.0/0000:bc:00.0/0000:bd:00.0/0000:be:10.0/0000:bf:00.0/0000:c0:04.0/0000:c4:00.0/0000:c5:10.0/0000:d1:00.1/infiniband/mlx5_13
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_14 -> ../../devices/pci0000:7e/0000:7e:01.1/0000:7f:00.0/0000:80:00.0/0000:81:00.0/0000:82:10.0/0000:83:00.0/0000:84:04.0/0000:88:00.0/0000:89:00.0/0000:8a:00.0/infiniband/mlx5_14
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_15 -> ../../devices/pci0000:7e/0000:7e:01.1/0000:7f:00.0/0000:80:00.0/0000:81:00.0/0000:82:10.0/0000:83:00.0/0000:84:04.0/0000:88:00.0/0000:89:00.0/0000:8a:00.1/infiniband/mlx5_15
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_16 -> ../../devices/pci0000:7e/0000:7e:01.1/0000:7f:00.0/0000:80:00.0/0000:81:00.0/0000:82:10.0/0000:83:00.0/0000:84:08.0/0000:8e:00.0/0000:8f:10.0/0000:94:00.0/infiniband/mlx5_16
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_17 -> ../../devices/pci0000:7e/0000:7e:01.1/0000:7f:00.0/0000:80:00.0/0000:81:00.0/0000:82:10.0/0000:83:00.0/0000:84:08.0/0000:8e:00.0/0000:8f:10.0/0000:94:00.1/infiniband/mlx5_17
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_2 -> ../../devices/pci0000:3f/0000:3f:01.1/0000:40:00.0/0000:41:00.0/0000:42:00.0/0000:43:10.0/0000:44:00.0/0000:45:04.0/0000:49:00.0/0000:4a:10.0/0000:4c:00.0/infiniband/mlx5_2
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_3 -> ../../devices/pci0000:3f/0000:3f:01.1/0000:40:00.0/0000:41:00.0/0000:42:00.0/0000:43:10.0/0000:44:00.0/0000:45:04.0/0000:49:00.0/0000:4a:10.0/0000:4c:00.1/infiniband/mlx5_3
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_4 -> ../../devices/pci0000:2c/0000:2c:03.1/0000:2d:00.0/infiniband/mlx5_4
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_5 -> ../../devices/pci0000:2c/0000:2c:03.1/0000:2d:00.1/infiniband/mlx5_5
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_6 -> ../../devices/pci0000:00/0000:00:01.1/0000:01:00.0/0000:02:00.0/0000:03:00.0/0000:04:10.0/0000:05:00.0/0000:06:04.0/0000:0a:00.0/0000:0b:00.0/0000:0c:00.0/infiniband/mlx5_6
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_7 -> ../../devices/pci0000:00/0000:00:01.1/0000:01:00.0/0000:02:00.0/0000:03:00.0/0000:04:10.0/0000:05:00.0/0000:06:04.0/0000:0a:00.0/0000:0b:00.0/0000:0c:00.1/infiniband/mlx5_7
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_8 -> ../../devices/pci0000:00/0000:00:01.1/0000:01:00.0/0000:02:00.0/0000:03:00.0/0000:04:10.0/0000:05:00.0/0000:06:08.0/0000:11:00.0/0000:12:10.0/0000:16:00.0/infiniband/mlx5_8
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 mlx5_9 -> ../../devices/pci0000:00/0000:00:01.1/0000:01:00.0/0000:02:00.0/0000:03:00.0/0000:04:10.0/0000:05:00.0/0000:06:08.0/0000:11:00.0/0000:12:10.0/0000:16:00.1/infiniband/mlx5_9

/sys/class/net:
total 0
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 eth0 -> ../../devices/virtual/net/eth0
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 lo -> ../../devices/virtual/net/lo
lrwxrwxrwx. 1 root root 0 Feb 24 20:53 tunl0 -> ../../devices/virtual/net/tunl0
```

When finished, remove the validation pod:

```bash
kubectl delete pod test-hca-pod
```

You now have a validated RDMA-enabled workload cluster ready for further
testing.
