## x8scan

Mass-scale hidden parameter discovery using x8. A high-performance wrapper that parallelizes x8 for efficient parameter discovery across multiple targets.

## Overview

x8scan supercharges [x8](https://github.com/Sh1Yo/x8) by enabling concurrent scanning of multiple URLs, dramatically reducing execution time while maintaining the powerful detection capabilities of the original tool.

## Features
- **🚀 Mass Parallelization**: Scan hundreds of URLs concurrently with configurable concurrency
- **🔧 Automatic Setup**: Self-downloads default wordlist on first run
- **🛠️ Tool Integration**: Clean output formats for seamless pipeline integration
- **📊 Multiple Output Formats**: Standard, JSON, and filtered URL outputs
- **⚡ Performance Optimized**: Significant speed improvements over sequential processing

## Prerequisites

x8 must be installed and accessible in your PATH. Install x8 from [https://github.com/Sh1Yo/x8](https://github.com/Sh1Yo/x8)

## Installation
```
go install github.com/rix4uni/x8scan@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/x8scan/releases/download/v0.0.2/x8scan-linux-amd64-0.0.2.tgz
tar -xvzf x8scan-linux-amd64-0.0.2.tgz
rm -rf x8scan-linux-amd64-0.0.2.tgz
mv x8scan ~/go/bin/x8scan
```
Or download [binary release](https://github.com/rix4uni/x8scan/releases) for your platform.

## Compile from source
```
git clone --depth 1 https://github.com/rix4uni/x8scan.git
cd x8scan; go install
```

## Usage

```yaml
Usage of x8scan:
  -a, --append-output string   File to append the output instead of overwriting.
  -c, --concurrency int        Number of concurrent URL scans (default 10)
  -j, --json                   Output results in JSON format
  -m, --methods string         HTTP methods to test (comma-separated) (default "GET")
  -o, --output string          File to save the output.
      --silent                 Silent mode.
  -t, --tfilter                Print only transformed URLs for tool integration.
      --verbose                Enable verbose output for debugging purposes.
      --version                Print the version of the tool and exit.
  -w, --wordlist string        Custom wordlist (default "~/.config/x8scan/params.txt")
```

## Usage Examples

### Basic Scanning
```yaml
# Single URL with default settings
echo "https://wstp.kzlabs.store/6.php" | x8scan

# Single URL with multiple methods
echo "https://wstp.kzlabs.store/6.php" | x8scan -m GET,POST

# Custom wordlist
echo "https://wstp.kzlabs.store/6.php" | x8scan -w /path/to/wordlist.txt
```

### Output Formats
```yaml
# JSON output
echo "https://wstp.kzlabs.store/6.php" | x8scan -j

# Only transformed URLs (for tool integration)
echo "https://wstp.kzlabs.store/6.php" | x8scan -t

# Save to file
echo "https://wstp.kzlabs.store/6.php" | x8scan -o results.txt
```

## Performance Comparison

| Scenario | Time | Command |
|----------|------|---------|
| Sequential (5 URLs) | ~2m | `cat urls.txt \| x8scan -m GET,POST` |
| **Concurrent (5 URLs)** | **~25s** | `cat urls.txt \| x8scan -m GET,POST -c 10` |

**→ 85% faster execution with concurrency**

## Best Practices
1. **Domain Distribution**: Use `-concurrency` primarily for scanning different domains/subdomains
2. **Rate Limiting**: Randomize URLs with `shuf` when scanning same-domain endpoints
3. **Progressive Scanning**: Start with lower concurrency (`-c 10`) and increase based on target responsiveness
4. **Output Management**: Use `-tfilter` for tool pipelines and `-j` for automated processing

## Troubleshooting
- **x8 not found**: Ensure x8 is installed and accessible in your PATH
- **Wordlist issues**: Delete `~/.config/x8scan/params.txt` to trigger redownload
- **Performance problems**: Reduce concurrency with `-c` for rate-limited targets
- **Verbose debugging**: Use `--verbose` flag to identify specific issues

## Acknowledgments
- [Sh1Yo](https://github.com/Sh1Yo) for the original [x8](https://github.com/Sh1Yo/x8) tool
