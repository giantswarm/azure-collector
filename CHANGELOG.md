# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

## [2.5.1] - 2021-06-22

### Changed

- Avoid scraping service principal expiration on customer installations.

## [2.5.0] - 2021-06-22

### Changed

- Always use `tenant ID` and ignore `GS tenant ID` flag.

## [2.4.0] - 2020-12-16

### Added

- Add `azure_operator_cluster_release` metric.

## [2.3.0] - 2020-12-11

### Added

- Add new collector to expose `Cluster` CR conditions as metrics to be used as inhibitions.

## [2.2.0] - 2020-11-05

### Added

- Add collector to expose cluster creation time.

## [2.1.2] - 2020-11-04

### Fixed

- Do not export data about customer's VPN gateways.

## [2.1.1] - 2020-10-27

### Fixed

- Try to find credential secret in default namespace if it's not found using organization.

## [2.1.0] - 2020-10-23

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



[Unreleased]: https://github.com/giantswarm/azure-collector/compare/v2.5.1...HEAD
[2.5.1]: https://github.com/giantswarm/azure-collector/compare/v2.5.0...v2.5.1
[2.5.0]: https://github.com/giantswarm/azure-collector/compare/v2.4.0...v2.5.0
[2.4.0]: https://github.com/giantswarm/azure-collector/compare/v2.3.0...v2.4.0
[2.3.0]: https://github.com/giantswarm/azure-collector/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/giantswarm/azure-collector/compare/v2.1.2...v2.2.0
[2.1.2]: https://github.com/giantswarm/azure-collector/compare/v2.1.1...v2.1.2
[2.1.1]: https://github.com/giantswarm/azure-collector/compare/v2.1.0...v2.1.1
[2.1.0]: https://github.com/giantswarm/azure-collector/compare/v2.0.2...v2.1.0
[2.0.2]: https://github.com/giantswarm/azure-collector/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/giantswarm/azure-collector/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/giantswarm/azure-collector/compare/v1.0.5...v2.0.0
[1.0.4]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/giantswarm/kubernetes-node-exporter/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/giantswarm/azure-collector/releases/tag/v1.0.0
