
envsubst_cmd = "./hack/tools/bin/envsubst"
kubectl_cmd = "./hack/tools/bin/kubectl"
helm_cmd = "./hack/tools/bin/helm"
tools_bin = "./hack/tools/bin"

update_settings(k8s_upsert_timeout_secs = 60)  # on first tilt up, often can take longer than 30 seconds

#Add tools to path
os.putenv("PATH", os.getenv("PATH") + ":" + tools_bin)

keys = ["OCI_TENANCY_ID", "OCI_USER_ID", "OCI_CREDENTIALS_FINGERPRINT", "OCI_REGION", "OCI_CREDENTIALS_KEY_PATH"]

# set defaults
settings = {
    "allowed_contexts": [
        "kind-capoci",
    ],
    "deploy_cert_manager": True,
    "preload_images_for_kind": True,
    "kind_cluster_name": "capoci",
    "capi_version": "v1.4.1",
    "cert_manager_version": "v1.11.0",
    "kubernetes_version": "v1.24.3",
}

# global settings
settings.update(read_json(
    "tilt-settings.json",
    default = {},
))

if settings.get("trigger_mode") == "manual":
    trigger_mode(TRIGGER_MODE_MANUAL)

if "allowed_contexts" in settings:
    allow_k8s_contexts(settings.get("allowed_contexts"))

if "default_registry" in settings:
    default_registry(settings.get("default_registry"))

tilt_helper_dockerfile_header = """
# Tilt image
FROM golang:1.21.8 as tilt-helper
# Support live reloading with Tilt
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN wget --output-document /restart.sh --quiet https://raw.githubusercontent.com/tilt-dev/rerun-process-wrapper/master/restart.sh  && \
    wget --output-document /start.sh --quiet https://raw.githubusercontent.com/tilt-dev/rerun-process-wrapper/master/start.sh && \
    chmod +x /start.sh && chmod +x /restart.sh && chmod +x /go/bin/dlv && \
    touch /process.txt && chmod 0777 /process.txt `# pre-create PID file to allow even non-root users to run the image`
"""

tilt_dockerfile_header = """
FROM gcr.io/distroless/base:debug as tilt
WORKDIR /
COPY --from=tilt-helper /process.txt .
COPY --from=tilt-helper /start.sh .
COPY --from=tilt-helper /restart.sh .
COPY --from=tilt-helper /go/bin/dlv .
COPY manager .
"""

def validate_auth():
    substitutions = settings.get("kustomize_substitutions", {})
    os.environ.update(substitutions)
    for sub in substitutions:
        if sub[-4:] == "_B64":
            os.environ[sub[:-4]] = base64_decode(os.environ[sub])
            print("{} was not specified in tilt-settings.json, attempting to load {}".format(base64_decode(os.environ[sub]), sub))
    missing = [k for k in keys if not os.environ.get(k)]
    if missing:
        fail("missing kustomize_substitutions keys {} in tilt-setting.json".format(missing))

# Users may define their own Tilt customizations in tilt.d. This directory is excluded from git and these files will
# not be checked in to version control.
def include_user_tilt_files():
    user_tiltfiles = listdir("tilt.d")
    for f in user_tiltfiles:
        include(f)

# deploy CAPI
def deploy_capi():
    version = settings.get("capi_version")
    capi_uri = "https://github.com/kubernetes-sigs/cluster-api/releases/download/{}/cluster-api-components.yaml".format(version)
    cmd = "curl -sSL {} | {} | {} apply -f -".format(capi_uri, envsubst_cmd, kubectl_cmd)
    local(cmd, quiet = False)
    if settings.get("extra_args"):
        extra_args = settings.get("extra_args")
        if extra_args.get("core"):
            core_extra_args = extra_args.get("core")
            if core_extra_args:
                for namespace in ["capi-system", "capi-webhook-system"]:
                    patch_args_with_extra_args(namespace, "capi-controller-manager", core_extra_args)
        if extra_args.get("kubeadm-bootstrap"):
            kb_extra_args = extra_args.get("kubeadm-bootstrap")
            if kb_extra_args:
                patch_args_with_extra_args("capi-kubeadm-bootstrap-system", "capi-kubeadm-bootstrap-controller-manager", kb_extra_args)

def patch_args_with_extra_args(namespace, name, extra_args):
    args_str = str(local("{} get deployments {} -n {} -o jsonpath={{.spec.template.spec.containers[1].args}}".format(kubectl_cmd, name, namespace)))
    args_to_add = [arg for arg in extra_args if arg not in args_str]
    if args_to_add:
        args = args_str[1:-1].split()
        args.extend(args_to_add)
        patch = [{
            "op": "replace",
            "path": "/spec/template/spec/containers/1/args",
            "value": args,
        }]
        local("{} patch deployment {} -n {} --type json -p='{}'".format(kubectl_cmd, name, namespace, str(encode_json(patch)).replace("\n", "")))

# Build CAPOCI and add feature gates
def capoci():
    # Apply the kustomized yaml for this provider
    yaml = str(kustomizesub("./config/default"))

    # add extra_args if they are defined
    if settings.get("extra_args"):
        oci_extra_args = settings.get("extra_args").get("oci")
        if oci_extra_args:
            yaml_dict = decode_yaml_stream(yaml)
            append_arg_for_container_in_deployment(yaml_dict, "capoci-controller-manager", "capoci-system", "cluster-api-oci-controller", oci_extra_args)
            yaml = str(encode_yaml_stream(yaml_dict))
            yaml = fixup_yaml_empty_arrays(yaml)

    # Set up a local_resource build of the provider's manager binary.
    local_resource(
        "manager",
        cmd = 'mkdir -p .tiltbuild;CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \'-extldflags "-static"\' -o .tiltbuild/manager',
        deps = ["api", "cloud", "config", "controllers", "exp", "feature", "pkg", "go.mod", "go.sum", "main.go", "auth-config.yaml"],
        labels = ["cluster-api"],
    )

    dockerfile_contents = "\n".join([
        tilt_helper_dockerfile_header,
        tilt_dockerfile_header,
    ])

    entrypoint = ["sh", "/start.sh", "/manager"]
    extra_args = settings.get("extra_args")
    if extra_args:
        entrypoint.extend(extra_args)

    # Set up an image build for the provider. The live update configuration syncs the output from the local_resource
    # build into the container.
    docker_build(
        ref = "ghcr.io/oracle/cluster-api-oci-controller-amd64:dev",
        context = "./.tiltbuild/",
        dockerfile_contents = dockerfile_contents,
        target = "tilt",
        entrypoint = entrypoint,
        only = "manager",
        live_update = [
            sync(".tiltbuild/manager", "/manager"),
            run("sh /restart.sh"),
        ],
        ignore = ["templates"],
        network = "host",
    )

    #secret_settings(disable_scrub=True)
    k8s_yaml(blob(yaml))


def append_arg_for_container_in_deployment(yaml_stream, name, namespace, contains_image_name, args):
    for item in yaml_stream:
        if item["kind"] == "Deployment" and item.get("metadata").get("name") == name and item.get("metadata").get("namespace") == namespace:
            containers = item.get("spec").get("template").get("spec").get("containers")
            for container in containers:
                if contains_image_name in container.get("image"):
                    container.get("args").extend(args)

def fixup_yaml_empty_arrays(yaml_str):
    yaml_str = yaml_str.replace("conditions: null", "conditions: []")
    return yaml_str.replace("storedVersions: null", "storedVersions: []")

def waitforsystem():
    local(kubectl_cmd + " wait --for=condition=ready --timeout=300s pod --all -n capi-kubeadm-bootstrap-system")
    local(kubectl_cmd + " wait --for=condition=ready --timeout=300s pod --all -n capi-kubeadm-control-plane-system")
    local(kubectl_cmd + " wait --for=condition=ready --timeout=300s pod --all -n capi-system")

def base64_encode(to_encode):
    encode_blob = local("echo '{}' | tr -d '\n' | base64 - | tr -d '\n'".format(to_encode), quiet = True, echo_off = True)
    return str(encode_blob)

def base64_encode_file(path_to_encode):
    encode_blob = local("cat {} | tr -d '\n' | base64 - | tr -d '\n'".format(path_to_encode), quiet = True)
    return str(encode_blob)

def read_file_from_path(path_to_read):
    str_blob = local("cat {} | tr -d '\n'".format(path_to_read), quiet = True)
    return str(str_blob)

def base64_decode(to_decode):
    decode_blob = local("echo '{}' | base64 --decode -".format(to_decode), quiet = True, echo_off = True)
    return str(decode_blob)

def kustomizesub(folder):
    yaml = local('bash hack/kustomize-sub.sh {}'.format(folder), quiet=True)
    return yaml

##############################
# Actual work happens here
##############################

validate_auth()

include_user_tilt_files()

load("ext://cert_manager", "deploy_cert_manager")

if settings.get("deploy_cert_manager"):
    deploy_cert_manager()

deploy_capi()

capoci()

waitforsystem()
