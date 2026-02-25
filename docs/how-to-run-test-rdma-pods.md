# How to Run the RDMA Validation Pods

This guide walks through standing up the RDMA smoke-test workload on an OCI compute cluster using the manifests in this repository. It captures all of the tweaks we made while iterating on `docs/rdma_pod_test.yaml` so you can reproduce the working configuration end-to-end.

## Prerequisites

1. **Custom RDMA-enabled worker image**
   - Install the Mellanox userspace and tooling in your node image (run inside the image build pipeline):
     ```bash
     sudo dnf install -y rdma-core libibverbs-utils infiniband-diags perftest mstflint
     sudo modprobe mlx5_core mlx5_ib rdma_ucm ib_uverbs ib_umad
     echo -e "mlx5_core\nmlx5_ib" | sudo tee /etc/modules-load.d/rdma.conf
     ```
   - Confirm `/dev/infiniband` exists, snapshot the image, and update the Cluster API template to use the new worker image OCID.

2. **Pre-provisioned OCI Compute Cluster (HPC)**
   - The RDMA worker template assumes the cluster lives inside an OCI Compute Cluster. Update `computeClusterId` in your `cluster-template-rdma-test.yaml` before creating the workload cluster.

3. **RDMA device plugin deployed**
   - Deploy the Mellanox RDMA shared device plugin (`k8s-rdma-shared-dev-plugin`) so `/dev/infiniband` and RDMA resources are advertised to Kubernetes.
   - Verify each worker exposes a non-zero `rdma/hca_shared_devices_a` value: `kubectl describe node <node> | grep -A2 rdma/`.

4. **Network security rules**
   - Allow **TCP 18515** (ib_write_bw control channel) inbound/outbound between the two RDMA workers.
   - Allow **UDP 4791** (RoCEv2 data plane) inbound/outbound between the same nodes/subnets.
   - Apply rules in both the subnet security list and any Network Security Groups attached to the instances.

5. **Node labels**
   - Label the nodes you want to use for the test so the pods land on different workers:
     ```bash
     kubectl label node <server-node> rdma.networking/role=server --overwrite
     kubectl label node <client-node> rdma.networking/role=client --overwrite
     ```

## Manifest Overview (`docs/rdma_pod_test.yaml`)

Key aspects of the working manifest:

- Uses a headless `Service` plus a single-replica `StatefulSet` so the server pod gets a stable DNS name (`rdma-server-0.rdma-server.rdma-test.svc.cluster.local`).
- Runs both pods with `hostNetwork: true` and `ClusterFirstWithHostNet` to expose the host’s RDMA interfaces.
- Mounts the host’s `/dev/infiniband` and `/sys/class/infiniband` into each pod via `hostPath` volumes so device files are available when using host networking.
- Executes `ib_write_bw` with:
  - `--ib-dev mlx5_0 --ib-port 1`
  - `--gid-index ${GID_INDEX}` (we used `3`, matching the routable `::ffff:10.x.x.x` GID slot)
  - `--mtu 1024` (the active MTU reported by the HCA)
  - `--qp-timeout 14 --sl 0`
  - `--rdma_cm` to let RDMA CM handle the handshake over UDP 4791.
- The client takes `SERVER_HOST_IP` (the server node’s private IP, e.g., `10.0.67.206`) from an environment variable and connects directly to that address.

Here is the current manifest for reference (truncated to highlight the critical sections):

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: rdma-test
---
apiVersion: v1
kind: Service
metadata:
  name: rdma-server
  namespace: rdma-test
spec:
  clusterIP: None
  selector:
    app: rdma-perftest
    role: server
  ports:
  - name: rdma
    port: 18515
    targetPort: 18515
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: rdma-server
  namespace: rdma-test
spec:
  serviceName: rdma-server
  replicas: 1
  selector:
    matchLabels:
      app: rdma-perftest
      role: server
  template:
    metadata:
      labels:
        app: rdma-perftest
        role: server
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      volumes:
      - name: rdma-devices
        hostPath:
          path: /dev/infiniband
      - name: rdma-sys
        hostPath:
          path: /sys/class/infiniband
      containers:
      - name: rdma-server
        image: <your-registry>/rdma-test:latest
        env:
        - name: GID_INDEX
          value: "3"
        command:
        - bash
        - -c
        - |
          set -euo pipefail
          echo "===== ibv_devices ====="
          ibv_devices
          echo "===== /dev/infiniband ====="
          ls -l /dev/infiniband || true
          echo "===== ib_write_bw server ====="
          ib_write_bw \
            --ib-dev mlx5_0 \
            --ib-port 1 \
            --report_gbits \
            --run_infinitely \
            --duration 5 \
            --gid-index ${GID_INDEX} --qp-timeout 14 --sl 0 --mtu 1024 --rdma_cm
        volumeMounts:
        - name: rdma-devices
          mountPath: /dev/infiniband
        - name: rdma-sys
          mountPath: /sys/class/infiniband
          readOnly: true
        resources:
          limits:
            rdma/hca_shared_devices_a: 1
          requests:
            rdma/hca_shared_devices_a: 1
---
apiVersion: batch/v1
kind: Job
metadata:
  name: rdma-client
  namespace: rdma-test
spec:
  template:
    metadata:
      labels:
        app: rdma-perftest
        role: client
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      restartPolicy: Never
      volumes:
      - name: rdma-devices
        hostPath:
          path: /dev/infiniband
      - name: rdma-sys
        hostPath:
          path: /sys/class/infiniband
      containers:
      - name: rdma-client
        image: <your-registry>/rdma-test:latest
        env:
        - name: GID_INDEX
          value: "3"
        - name: SERVER_HOST_IP
          value: "10.0.67.206"
        command:
        - bash
        - -c
        - |
          set -euo pipefail
          echo "===== ibv_devices ====="
          ibv_devices
          echo "===== /dev/infiniband ====="
          ls -l /dev/infiniband || true
          echo "===== ib_write_bw client ====="
          ib_write_bw \
            ${SERVER_HOST_IP} \
            --ib-dev mlx5_0 \
            --ib-port 1 \
            --report_gbits \
            --gid-index ${GID_INDEX} --qp-timeout 14 --sl 0 --mtu 1024 --rdma_cm \
            --iters 5000
        volumeMounts:
        - name: rdma-devices
          mountPath: /dev/infiniband
        - name: rdma-sys
          mountPath: /sys/class/infiniband
          readOnly: true
        resources:
          limits:
            rdma/hca_shared_devices_a: 1
          requests:
            rdma/hca_shared_devices_a: 1
  backoffLimit: 0
```

Replace `<your-registry>/rdma-test:latest` with the image you built earlier (e.g., the OCI Registry path you pushed to).

## Running the Test

1. Apply/refresh the manifest:
   ```bash
   kubectl apply -f docs/rdma_pod_test.yaml
   ```
   If you need to re-run the client, delete only the job:
   ```bash
   kubectl delete job rdma-client -n rdma-test
   kubectl apply -f docs/rdma_pod_test.yaml
   ```

2. Watch the server output:
   ```bash
   kubectl logs -n rdma-test statefulset/rdma-server -f
   ```
   You should see `ibv_devices`, `/dev/infiniband` listings, and then repeated bandwidth metrics every 5 seconds (or a waiting message until the client connects).

3. Inspect the client results:
   ```bash
   kubectl logs -n rdma-test job/rdma-client
   ```
   A successful run looks like:
   ```
   local address: GID ...10:00:75:67
   remote address: GID ...10:00:67:206
   #bytes     #iterations    BW peak[Gb/sec]    BW average[Gb/sec]   MsgRate[Mpps]
   65536      5000             38.16              38.15              0.072768
   ```

## Troubleshooting Checklist

- **RTR failures with `Invalid MTU`**: Use an MTU supported by `ib_write_bw` (256–4096). We set `--mtu 1024` to match the active value.
- **RTR failures without MTU error**: Ensure UDP 4791 is allowed both ways, and confirm the same `GID_INDEX` is valid on both nodes (look for the `::ffff:10.x.x.x` GID entry).
- **`Unexpected CM event` errors**: Usually caused by UDP 4791 being blocked or a GID mismatch. Open the port and double-check `GID_INDEX`.
- **No `/dev/infiniband` inside host-network pods**: Mount the host paths explicitly (`hostPath` volumes) as shown above.
- **Pods land on the same node**: Reapply the node labels or add anti-affinity.
- **Need different payload sizes**: Add `--size <bytes>` or `--duration <seconds>` to the `ib_write_bw` commands and reapply.

With these settings, the RDMA validation pods now consistently produce ~38 Gb/s between two OCI compute cluster nodes. Adjust the node labels or the `SERVER_HOST_IP` env var to target different host pairs for additional testing.
