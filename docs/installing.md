# Installing Sentry Operator

## Requirements

- An existing Kubernetes cluster of version 1.11.3+

## Installation

To install the operator into your cluster, run:

```shell
kubectl apply -f https://github.com/jace-ys/sentry-operator/releases/download/v0.1.1/release.yaml
```

You can install a specific release using a different version number. Find all possible versions under [releases](https://github.com/jace-ys/sentry-operator/releases).

## Configuration

The Sentry operator assumes you have a single Sentry organization under which it will create its resources; you will need to configure the operator with access to your organization for it to work.

To do this, the operator is configured to read environment variables from a Kubernetes Secret with the name `sentry-operator-config` in the `sentry-operator-system` namespace.

To create the Secret, run:

```shell
kubectl create secret generic sentry-operator-config \
  --namespace sentry-operator-system \
  --from-literal SENTRY_ORGANIZATION=<required> \
  --from-literal SENTRY_TOKEN=<required> \
  --from-literal SENTRY_URL=<optional>
```

#### Configuration Options

The following configuration options are available:

- `SENTRY_ORGANIZATION` (required)

  The slug of the Sentry organization to be managed.

- `SENTRY_TOKEN` (required)

  The authentication token for communicating with the Sentry API. This token requires the following scopes:

  - `org:admin`, `org:write`, `org:read`
  - `team:admin`, `team:write`, `team:read`
  - `project:admin`, `project:write`, `project:read`

- `SENTRY_URL` (optional)

  The URL of the Sentry server. Defaults to `https://sentry.io/`.
