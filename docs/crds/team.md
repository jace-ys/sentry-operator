# `Team`

The `Team` custom resource allows for the provisioning and management of Sentry teams.

## Usage

A `Team` supports the following fields in its spec:

- `name` (required)

  Name of the Sentry team.

- `slug` (required)

  Slug of the Sentry team.

  It is generally recommended to use the same value as the team's name, as Sentry has some quirky behaviour about handling the uniqueness of slugs.

## Examples

#### Basic `Team`

```yaml
apiVersion: sentry.kubernetes.jaceys.me/v1alpha1
kind: Team
metadata:
  name: foo
spec:
  name: foo
  slug: foo
```
