# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added

- VMSS Rate limit collector uses both `AzureConfig` and `Cluster` CRDs to iterate through clusters.

## [2.0.2] - 2020-10-20

### Fixed

- Use same metric namespace/prefix for all exported metrics. This changed the VMSS Rate limit metrics.

## [2.0.1] - 2020-10-12

### Fixed

- Check for double scrape of the same subscription in the rate limit collector.

## [2.0.0] - 2020-10-06

### Changed

- Migrate to go modules.
- Update dependencies.
- Add dependabot configuration.
- Add release workflows.

## [1.0.5] 2020-10-05

### Changed

- Using right VMSS name after node pool changes in azure-operator.

## [1.0.4] 2020-04-23

### Changed

- Use Release.Revision in annotation for Helm 3 compatibility.

## [1.0.3] 2020-04-20

### Changed

- Collector doesn't use credential paramaters, rather reads credentials from secrets in the control plane.
- Removed `Validate()` method for azureClientSetConfig. Use factory method instead.
- Make collectors use utility methods in the `credential` package.

## [1.0.2] 2020-04-19

### Changed

- Deploy as a unique app in app collection.

## [1.0.1] 2020-04-17

### Fixed

- Fixed label selector to filter secrets from control plane.

## [1.0.0]

### Added

- First release.



[Unreleased]: https://github.com/giantswarm/azure-collector/compare/v2.0.2...HEAD
[2.0.2]: https://github.com/giantswarm/azure-collector/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/giantswarm/azure-collector/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/giantswarm/azure-collector/compare/v1.0.5...v2.0.0
[1.0.4]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/giantswarm/azure-collector/releases/tag/v1.0.0
