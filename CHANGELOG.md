# Changelog

All notable changes to FFmpego will be documented in this file.

## [1.4.0] - 2025-01-03

### Added
- **Beginner-friendly presets** for silence detection
  - `SilenceDurationVeryShort`, `SilenceDurationShort`, `SilenceDurationMedium`, `SilenceDurationLong`, `SilenceDurationVeryLong`
  - `SilenceThresholdVeryStrict`, `SilenceThresholdStrict`, `SilenceThresholdModerate`, `SilenceThresholdRelaxed`, `SilenceThresholdVeryRelaxed`
  - No need to remember dB or millisecond values anymore!

### Changed
- **Better function naming**: `DetectSilence()` → `GetNonSilentSegments()`
  - Old name was confusing (returned non-silent segments, not silence)
  - New name clearly describes what it returns
- **Improved concurrency safety**: Temporary files now use unique names
  - Safe to run multiple processes simultaneously
  - Uses `os.MkdirTemp()` and `os.CreateTemp()` for guaranteed uniqueness

### Improved
- All examples now use the new presets
- README updated with clear preset documentation
- Better comments explaining thread-safety

## [1.3.0] - 2025-01-03

### Added
- Initial parallel processing support for video segments
- Example showing how to remove silence from videos

## Earlier Versions

See git history for changes before v1.3.0

---

## How to Read This Changelog

- **Added**: New features
- **Changed**: Changes to existing features
- **Improved**: Performance or usability improvements
- **Fixed**: Bug fixes
- **Removed**: Features that were removed

Each version follows [Semantic Versioning](https://semver.org/):
- **Major** (1.x.x): Breaking changes
- **Minor** (x.1.x): New features, backward compatible
- **Patch** (x.x.1): Bug fixes, backward compatible
