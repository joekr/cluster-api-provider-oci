FROM ubuntu:22.04
LABEL description="Custom container with RDMA perftest tools"

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -y \
    && apt-get install -y --no-install-recommends \
        apt-utils udev bash sysstat iproute2 \
        rdma-core ibverbs-utils infiniband-diags perftest mstflint \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

CMD ["bash"]