[![ci-badge]][ci-workflow] [![release-badge]][release-workflow]

[ci-badge]: https://github.com/jace-ys/sentry-operator/workflows/ci/badge.svg
[ci-workflow]: https://github.com/jace-ys/sentry-operator/actions?query=workflow%3Aci
[release-badge]: https://github.com/jace-ys/sentry-operator/workflows/release/badge.svg
[release-workflow]: https://github.com/jace-ys/sentry-operator/actions?query=workflow%3Arelease

# Sentry Operator

A Kubernetes operator for automating the provisioning and management of [Sentry](<(https://sentry.io/)>) resources via Kubernetes CRDs.

## Foreword

**Until the Sentry API ([currently v0](https://docs.sentry.io/api/#versioning)) reaches a stable version, the Sentry operator might undergo breaking changes and will thus be marked as not production-ready - use this at your own risk.**

The Sentry operator assumes you have a single Sentry organization under which it will create its resources; it currently does not support multi-organization requirements.

## Features

- Provisioning and management of Sentry teams, projects and project keys.
- Automated creation of Kubernetes Secrets containing [Sentry DSNs](https://docs.sentry.io/error-reporting/quickstart/#configure-the-sdk).
- Support for [on-premise instances of Sentry](https://github.com/getsentry/onpremise).

## Installation

See documentation on [Installing](docs/installing.md).

## CRDs

The following CRDs are available:

- [`Team`](docs/crds/team.md)
- [`Project`](docs/crds/project.md)
- [`ProjectKey`](docs/crds/projectkey.md)

To get a better idea on using these CRDs, take a look at the [examples](examples). Depending on your setup, you may or may not need to use all of them.

## Limitations

- The Sentry API doesn't provide endpoints for managing the members of a team. This limits the usefulness of the `Team` CRD as you'll still need to manually add and delete team members via the Sentry UI.
- Changing a project's team via the Sentry API has been [deprecated](https://docs.sentry.io/api/projects/put-project-details/). To do this, you'll need to manually update the team via the Sentry UI, and update your `Project` spec accordingly to reflect the change.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Roadmap

- [ ] Add E2E tests
- [ ] Expose Prometheus metrics

## License

See [LICENSE](LICENSE).
