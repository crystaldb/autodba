# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.3.0] - 2024-10-01

### Added
- **Support for multiple databases** allowing better management of database fleets, and also improving scalability and flexibility.
- **Collector API Server stubs** to enable a smoother collector integration.
- **Activity cube filters** to refine data views within the Activity Cube.
- **Longer timeframes in time bar** for improved historical data analysis in the UI.
- **UI retry logic on app start**, ensuring a seamless user experience even when the backend is delayed.

### Changed
- **Prometheus query timeout increased** for more stable data retrieval during complex queries.
- **HTML meta title updated per page** to improve user navigation.

### Fixed
- **Flaky test caused by `time.now`** to ensure test stability.
- **Handling of `dbidentifier`** between the front-end, backend, and Prometheus for consistent identification.
- **Stale-marker handling** for multi-database settings, preventing erroneous data from being displayed.
- **UI eslint errors** cleaned up for better code quality.
- **Collector populating `cc_pg_stat_activity` time-series** in Prometheus for accurate activity tracking.

### Security
- **PromQL input validation** added to improve backend security for query handling.

## [0.2.0] - 2024-09-19

### Added
- Time bar for access to historical data
- Allow operation without AWS credentials

### Changed
- Simplified packaging and installation

### Fixed
- Ensure consistent UI performance with limits and throttling

## [0.1.0] - 2024-09-06

### Added
- **Initial release** of AutoDBA for PostgreSQL.
- **Prometheus database** integration for capturing time-series metrics.
- **BFF API** for backend-for-frontend architecture.
- **RDS exporter** to gather AWS RDS-specific metrics.
- **Activity Cube UI** providing a multi-dimensional visualization of database activity.
- **Database metrics UI** to display detailed PostgreSQL database statistics.
- **System metrics UI** for tracking hardware and OS performance.
- **Basic installation scripts** for `.tar.gz`, `.deb`, and `.rpm` formats.
- **Dockerized installation** for easier deployment and testing.
