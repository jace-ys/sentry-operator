# Development

## Requirements

The following tools are required to start developing on the Sentry operator:

- `go v1.13+`
- `kind v0.8.1+`
- `docker v17.03+`
- `kubectl v1.11.3+`
- `kustomize v3.1.0+`
- `kubebuilder v2.3.1+`

You can install Kubebuilder and Kustomize by running:

```shell
./scripts/install.sh
```

You should also have a Sentry organization to test your changes against. It is recommended to [create a sandbox account](https://sentry.io/signup/) for this.

For the operator to communicate with the Sentry API, you will need to provide an authentication token from a user under your organization, with the following scopes:

- `org:admin`, `org:write`, `org:read`
- `team:admin`, `team:write`, `team:read`
- `project:admin`, `project:write`, `project:read`

## Getting Started

1. Start a local Kubernetes cluster using `kind`:

   ```shell
   make cluster
   ```

2. Install the CRDs into your cluster:

   ```shell
   make install
   ```

3. Create a `.env` file in the repository's root containing the following environment variables:

   ```shell
   SENTRY_ORGANIZATION=<organization> # The slug of your Sentry organization
   SENTRY_TOKEN=<token> # Your Sentry authentication token
   ```

4. Run the operator locally:

   ```
   make run
   ```

5. Write your code and verify the expected behaviour against your Sentry organization. Also include any appropriate tests.

6. Run all test suites:

   ```
   make test
   ```

## Resources

- [A Tour of Go](https://tour.golang.org/welcome/1)
- [Kubernetes CRDs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [The Kubebuilder Book](https://book.kubebuilder.io/)
- [Sentry API documentation](https://docs.sentry.io/api/)
- [Ginkgo documentation](https://onsi.github.io/ginkgo/)
- [kind documentation](https://kind.sigs.k8s.io/)
