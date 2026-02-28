# zumba Design

## Overview
CLI tool to search conda packages with enhanced metadata (description, homepage, etc.) that mamba doesn't provide.

## Tech Stack
- **Language:** Go
- **CLI Style:** Git-style subcommands (`zumba search TERM`)
- **CLI Library:** cobra

## Commands

### `zumba search TERM [--channel CHANNEL]`
Search for packages matching TERM.
- Default channel: conda-forge
- Output: table with name, version, summary, homepage

### `zumba info PACKAGE [--channel CHANNEL]`
Show detailed package information.
- All versions available
- Full description
- Dependencies
- Links (homepage, docs, source)

## Data Sources

### Primary: repodata.json
- URL: `https://conda.anaconda.org/{channel}/{platform}/repodata.json`
- Contains: package names, versions, dependencies, hashes
- Cached locally for speed

### Secondary: channeldata.json
- URL: `https://conda.anaconda.org/{channel}/channeldata.json`
- Contains: descriptions, homepages, summaries per package
- Cached locally

## Caching Strategy
- Cache in `~/.cache/zumba/` (or XDG cache dir)
- repodata.json: refresh after 1 hour
- channeldata.json: refresh after 1 hour
- Implement `--refresh` flag to force update

## Match Algorithm
1. Exact name match (score: 100)
2. Name starts with term (score: 80)
3. Name contains term (score: 60)
4. Summary contains term (score: 40)
5. Description contains term (score: 20)

## Output Format
Default: table view
- `--json` flag for JSON output
- `--wide` flag for more columns

## Project Structure
```
cmd/zumba/            - CLI entry points
  main.go
  search.go
  info.go
internal/conda/       - Conda data types and fetching
  types.go
  repodata.go
  channel.go
  cache.go
```

## Architecture Goals
- Fast: leverage caching, parallel requests
- Offline-capable: use cached data when available
- Extensible: easy to add new commands