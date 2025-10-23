# INNGen

![Go](https://github.com/z0rr0/inngen/workflows/Go/badge.svg)
![Version](https://img.shields.io/github/tag/z0rr0/inngen.svg)
![License](https://img.shields.io/github/license/z0rr0/inngen.svg)

Taxpayer Identification Number (INN) generator and validator

## Description

This is a mixed application that can be run as:

1. A console tool to generate and validate Taxpayer Identification Numbers (INN)
2. A web application for the same purposes

## Features

- **Validate INN**: Check if an INN is valid (supports both 10-digit juridical and 12-digit physical person INNs)
- **Generate Physical Person INN**: Create valid 12-digit INNs for physical persons
- **Generate Juridical Person INN**: Create valid 10-digit INNs for juridical persons
- **Web Interface**: User-friendly web application with all functionality

## Installation

```bash
make build
# result file: inngen
```

## Usage

### Console Tool

#### Validate INN

```bash
./inngen -c <INN>
```

Example:
```bash
./inngen -c 7707083893
# Output: INN 7707083893 is valid (juridical person)

./inngen -c 500100732259
# Output: INN 500100732259 is valid (physical person)

./inngen -c 500100732250
# Output: INN 500100732250 invalid: invalid INN checksum: invalid physical inn, 12th digit is 0, expected 9
```

#### Generate INNs

```bash
./inngen -f [count] -j [count]
```

If no count is specified, generates 5 INNs by default.

Example:
```bash
./inngen -f 0
# No INNs generated for physical persons

./inngen -f 3
# Generates 3 INNs for physical persons and 5 for juridical ones

./inngen -f 2 -j 3
# Generates 2 INNs for physical persons and 3 for juridical ones
```

#### Run as Web Application

In development yet!

```bash
./inngen -w
```

This starts a web server on `127.0.0.1:2288` by default.
Open your browser and navigate to `http://127.0.0.1:2288` to use the web interface.

### Web Application

The web interface provides:

- **Validation form**: Enter an INN to check if it's valid
- **Physical person generator**: Generate multiple valid 12-digit INNs
- **Juridical person generator**: Generate multiple valid 10-digit INNs

## INN Format

- **Physical Person (12 digits)**: Uses two checksum digits (positions 11 and 12)
- **Juridical Person (10 digits)**: Uses one checksum digit (position 10)

The checksums are calculated using specific coefficients according to Russian INN validation rules.

## Testing

Run tests:

```bash
make test
```

Run benchmarks:

```bash
make bench
```

## License

This source code is governed by a [BSD 3-Clause](https://opensource.org/licenses/BSD-3-Clause)
license that can be found in the [LICENSE](https://github.com/z0rr0/inngen/blob/main/LICENSE) file.
