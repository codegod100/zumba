# zumba

A CLI tool to search conda packages with enhanced metadata that mamba doesn't provide.

## Why?

`mamba search` is fast but only shows basic package info. `zumba` fetches additional metadata like descriptions, homepages, and licenses from `channeldata.json`.

## Installation

```bash
go install github.com/nandi/zumba/cmd/zumba@latest
```

Or build from source:

```bash
git clone https://github.com/nandi/zumba.git
cd zumba
go build -o zumba ./cmd/zumba
```

## Usage

### Search for packages

```bash
# Basic search
zumba search numpy

# Wide output (includes homepage)
zumba search pandas --wide

# JSON output
zumba search scipy --json

# Search a different channel
zumba search star --channel bioconda

# Force refresh cached data
zumba search matplotlib --refresh
```

### Get detailed package info

```bash
# Show package details
zumba info numpy

# JSON output
zumba info pandas --json

# Different channel
zumba info star --channel bioconda
```

## Output Examples

### Search

```
$ zumba search numpy --wide
NAME                    VERSION    LICENSE          HOMEPAGE                              SUMMARY
numpy                   2.4.2      BSD-3-Clause      http://numpy.org/                     The fundamental package for scientific computing...
numpy-financial         1.0.0      BSD-3-Clause      https://github.com/numpy-numpy/...    Simple financial functions
numpydoc                 1.10.0     BSD-3-Clause      https://github.com/numpy/numpydoc     Numpy's Sphinx extensions
```

### Info

```
$ zumba info pandas
Name:       pandas
Version:    3.0.1
License:    BSD-3-Clause
Homepage:   http://pandas.pydata.org/
Dev URL:    https://github.com/pandas-dev/pandas
Doc URL:    https://pandas.pydata.org/docs/
Platforms:  [linux-64 linux-aarch64 linux-ppc64le osx-64 osx-arm64 win-32 win-64]

Summary:  Powerful data structures for data analysis, time series, and statistics
```

## Data Sources

| Source | Content | CacheTTL |
|--------|---------|----------|
| `repodata.json` | Package names, versions, dependencies | 1 hour |
| `channeldata.json` | Descriptions, homepages, licenses | 1 hour |

Cache location: `~/.cache/zumba/` (override with `ZUMBA_CACHE_DIR`)

## Comparison

| Feature | mamba search | zumba search |
|---------|--------------|--------------|
| Package names | ✅ | ✅ |
| Versions | ✅ | ✅ |
| Description | ❌ | ✅ |
| Homepage | ❌ | ✅ |
| License | ❌ | ✅ |
| Supported platforms | ❌ | ✅ |

## Commands

| Command | Description |
|---------|-------------|
| `zumba search TERM` | Search packages by name/summary |
| `zumba info PACKAGE` | Show detailed package info |

## Flags

| Flag | Description |
|------|-------------|
| `-c, --channel` | Conda channel (default: conda-forge) |
| `-p, --platform` | Platform (default: noarch) |
| `-j, --json` | Output as JSON |
| `-w, --wide` | Show more columns |
| `-r, --refresh` | Force refresh cached data |

## License

MIT