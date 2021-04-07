# Configure Image Registry

This document describes the method to configure the image registry for `containerd` for use with the `cri` plugin.

## Configure Registry Endpoint

With containerd, `docker.io` is the default image registry. You can also set up other image registries similar to docker.

To configure image registries create/modify the `/etc/containerd/config.toml` as follows:

```toml
# Config file is parsed as version 1 by default.
# To use the long form of plugin names set "version = 2"
# explicitly use v2 config format
version = 2

[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
    endpoint = ["https://registry-1.docker.io"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."test.https-registry.io"]
    endpoint = ["https://HostIP1:Port1"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."test.http-registry.io"]
    endpoint = ["http://HostIP2:Port2"]
  # wildcard matching is supported but not required.
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."*"]
    endpoint = ["https://HostIP3:Port3"]
```

The default configuration can be generated by `containerd config default > /etc/containerd/config.toml`.

The endpoint is a list that can contain multiple image registry URLs split by commas. When pulling an image
from a registry, containerd will try these endpoint URLs one by one, and use the first working one. Please note
that if the default registry endpoint is not already specified in the endpoint list, it will be automatically
tried at the end with scheme `https` and path `v2`, e.g. `https://gcr.io/v2` for `gcr.io`.

As an example, for the image `gcr.io/library/busybox:latest`, the endpoints are:

* `gcr.io` is configured: endpoints for `gcr.io` + default endpoint `https://gcr.io/v2`.
* `*` is configured, and `gcr.io` is not: endpoints for `*` + default
  endpoint `https://gcr.io/v2`.
* None of above is configured: default endpoint `https://gcr.io/v2`.

After modify this config, you need restart the `containerd` service.

## Configure Registry TLS Communication

`cri` plugin also supports configuring TLS settings when communicating with a registry.

To configure the TLS settings for a specific registry, create/modify the `/etc/containerd/config.toml` as follows:

```toml
# explicitly use v2 config format
version = 2

# The registry host has to be a domain name or IP. Port number is also
# needed if the default HTTPS or HTTP port is not used.
[plugins."io.containerd.grpc.v1.cri".registry.configs."my.custom.registry".tls]
    ca_file   = "ca.pem"
    cert_file = "cert.pem"
    key_file  = "key.pem"
```

In the config example shown above, TLS mutual authentication will be used for communications with the registry endpoint located at <https://my.custom.registry>.
`ca_file` is file name of the certificate authority (CA) certificate used to authenticate the x509 certificate/key pair specified by the files respectively pointed to by `cert_file` and `key_file`.

`cert_file` and `key_file` are not needed when TLS mutual authentication is unused.

```toml
# explicitly use v2 config format
version = 2

[plugins."io.containerd.grpc.v1.cri".registry.configs."my.custom.registry".tls]
    ca_file   = "ca.pem"
```

To skip the registry certificate verification:

```toml
# explicitly use v2 config format
version = 2

[plugins."io.containerd.grpc.v1.cri".registry.configs."my.custom.registry".tls]
  insecure_skip_verify = true
```

## Configure Registry Credentials

`cri` plugin also supports docker like registry credential config.

To configure a credential for a specific registry, create/modify the
`/etc/containerd/config.toml` as follows:

```toml
# explicitly use v2 config format
version = 2

# The registry host has to be a domain name or IP. Port number is also
# needed if the default HTTPS or HTTP port is not used.
[plugins."io.containerd.grpc.v1.cri".registry.configs."gcr.io".auth]
  username = ""
  password = ""
  auth = ""
  identitytoken = ""
```

The meaning of each field is the same with the corresponding field in `.docker/config.json`.

Please note that auth config passed by CRI takes precedence over this config.
The registry credential in this config will only be used when auth config is
not specified by Kubernetes via CRI.

After modifying this config, you need to restart the `containerd` service.

### Configure Registry Credentials Example - GCR with Service Account Key Authentication

If you don't already have Google Container Registry (GCR) set-up then you need to do the following steps:

* Create a Google Cloud Platform (GCP) account and project if not already created (see [GCP getting started](https://cloud.google.com/gcp/getting-started))
* Enable GCR for your project (see [Quickstart for Container Registry](https://cloud.google.com/container-registry/docs/quickstart))
* For authentication to GCR: Create [service account and JSON key](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key)
* The JSON key file needs to be downloaded to your system from the GCP console
* For access to the GCR storage: Add service account to the GCR storage bucket with storage admin access rights (see [Granting permissions](https://cloud.google.com/container-registry/docs/access-control#grant-bucket))

Refer to [Pushing and pulling images](https://cloud.google.com/container-registry/docs/pushing-and-pulling) for detailed information on the above steps.

> Note: The JSON key file is a multi-line file and it can be cumbersome to use the contents as a key outside of the file. It is worthwhile generating a single line format output of the file. One way of doing this is using the `jq` tool as follows: `jq -c . key.json`

It is beneficial to first confirm that from your terminal you can authenticate with your GCR and have access to the storage before hooking it into containerd. This can be verified by performing a login to your GCR and
pushing an image to it as follows:

```console
docker login -u _json_key -p "$(cat key.json)" gcr.io

docker pull busybox

docker tag busybox gcr.io/your-gcp-project-id/busybox

docker push gcr.io/your-gcp-project-id/busybox

docker logout gcr.io
```

Now that you know you can access your GCR from your terminal, it is now time to try out containerd.

Edit the containerd config (default location is at `/etc/containerd/config.toml`)
to add your JSON key for `gcr.io` domain image pull
requests:

```toml
version = 2

[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["https://registry-1.docker.io"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."gcr.io"]
      endpoint = ["https://gcr.io"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs]
    [plugins."io.containerd.grpc.v1.cri".registry.configs."gcr.io".auth]
      username = "_json_key"
      password = 'paste output from jq'
```

> Note: `username` of `_json_key` signifies that JSON key authentication will be used.

Restart containerd:

```console
service containerd restart
```

Pull an image from your GCR with `crictl`:

```console
$ sudo crictl pull gcr.io/your-gcp-project-id/busybox

DEBU[0000] get image connection
DEBU[0000] connect using endpoint 'unix:///run/containerd/containerd.sock' with '3s' timeout
DEBU[0000] connected successfully using endpoint: unix:///run/containerd/containerd.sock
DEBU[0000] PullImageRequest: &PullImageRequest{Image:&ImageSpec{Image:gcr.io/your-gcr-instance-id/busybox,},Auth:nil,SandboxConfig:nil,}
DEBU[0001] PullImageResponse: &PullImageResponse{ImageRef:sha256:78096d0a54788961ca68393e5f8038704b97d8af374249dc5c8faec1b8045e42,}
Image is up to date for sha256:78096d0a54788961ca68393e5f8038704b97d8af374249dc5c8faec1b8045e42
```

---

NOTE: The configuration syntax used in this doc is in version 2 which is the recommended since `containerd` 1.3. For the previous config format you can reference [https://github.com/containerd/cri/blob/release/1.2/docs/registry.md](https://github.com/containerd/cri/blob/release/1.2/docs/registry.md).