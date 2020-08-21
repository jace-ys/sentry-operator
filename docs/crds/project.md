# `Project`

The `Project` custom resource allows for the provisioning and management of Sentry projects.

## Usage

A `Project` supports the following fields in its spec:

- `team` (required)

  Slug of the Sentry team that this project should be created under.

- `name` (required)

  Name of the Sentry project.

- `slug` (required)

  Slug of the Sentry project.

  It is generally recommended to use the same value as the project's name, as Sentry has some quirky behaviour about handling the uniqueness of slugs.

## Examples

#### Basic `Project`

```yaml
apiVersion: sentry.kubernetes.jaceys.me/v1alpha1
kind: Project
metadata:
  name: bar
spec:
  team: foo
  name: bar
  slug: bar
```
