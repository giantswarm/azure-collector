# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

### Changed

- Collector doesn't use credential paramaters, rather reads credentials from secrets in the control plane.
- Removed `Validate()` method for azureClientSetConfig. Use factory method instead.
- Make collectors use utility methods in the `credential` package.

## [1.0.1] 2020-04-17

### Fixed

- Fixed label selector to filter secrets from control plane.

## [1.0.0]

### Added

- First release.



[Unreleased]: https://github.com/giantswarm/azure-collector/compare/v1.0.0...HEAD

[1.0.0]: https://github.com/giantswarm/azure-collector/releases/tag/v1.0.0
