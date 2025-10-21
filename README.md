# inngen

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
go build -o inngen
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
```

#### Generate INNs for Physical Persons
```bash
./inngen -f [count]
```

If no count is specified, generates 5 INNs by default.

Example:
```bash
./inngen -f 0
# Generates 5 INNs (default)

./inngen -f 3
# Generates 3 INNs
```

#### Generate INNs for Juridical Persons
```bash
./inngen -j [count]
```

If no count is specified, generates 5 INNs by default.

Example:
```bash
./inngen -j 0
# Generates 5 INNs (default)

./inngen -j 2
# Generates 2 INNs
```

#### Run as Web Application
```bash
./inngen -w
```

This starts a web server on `127.0.0.1:2288` by default. Open your browser and navigate to `http://127.0.0.1:2288` to use the web interface.

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
go test -v
```

Run benchmarks:
```bash
go test -bench=.
```

## License

See LICENSE file for details.
