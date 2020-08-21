# `ProjectKey`

The `ProjectKey` custom resource allows for the provisioning and management of client keys that belong to a Sentry project. These client keys contain a [Sentry DSN](https://docs.sentry.io/error-reporting/quickstart/#configure-the-sdk) value that tells your application's Sentry SDK where to send events to.

## Usage

A `ProjectKey` supports the following fields in its spec:

- `project` (required)

  Slug of the Sentry project that this project key should be created under.

- `name` (required)

  Name of the Sentry project key.

### `ProjectKey` Secrets

When creating a `ProjectKey`, the Sentry operator will automatically provision a Kubernetes Secret containing the associated Sentry DSN in the same namespace. It will inherit the name of your `ProjectKey`, suffixed with `sentry-projectkey-`.

Any labels and annotations attached to `ProjectKey`s are also automatically propagated to their affiliated Secret.

For example, the [basic `ProjectKey` example](#basic-projectkey) below will result in the creation of a Secret like the following:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sentry-projectkey-bar-production
type: Opaque
data:
  SENTRY_DSN: <dsn-value>
```

You can then configure your Pods to automatically consume the injected Sentry DSN value using a `secretKeyRef` to the above Secret:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: bar
spec:
  restartPolicy: OnFailure
  containers:
    - name: bar
      image: ubuntu:18.04
      command:
        - echo
      args:
        - $(SENTRY_DSN)
      env:
        - name: SENTRY_DSN
          valueFrom:
            secretKeyRef:
              name: sentry-projectkey-bar-production
              key: SENTRY_DSN
```

## Examples

#### Basic `ProjectKey`

```yaml
apiVersion: sentry.kubernetes.jaceys.me/v1alpha1
kind: ProjectKey
metadata:
  name: bar-production
spec:
  project: bar
  name: production
```
