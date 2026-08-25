# Next Departures API

A REST API for UK bus stop information and real-time next departures.

## Features

- **NaPTAN Data**: Imports and serves National Public Transport Access Nodes (NaPTAN) data.
- **Bounding Box Search**: Find bus stops within a specific geographic area.
- **Real-time Departures**: Fetch live departure information for specific stops using the SIRI (Stop Monitoring) protocol via TransportAPI.
- **Background Updates**: Automated CRON jobs for periodic data maintenance.
- **Monitoring**: Built-in health checks and Prometheus metrics integration.
- **Error Tracking**: Sentry integration for monitoring and error reporting.

## Prerequisites

- Go 1.27 or higher
- SQLite3
- TransportAPI credentials (for real-time departures)

## Getting Started

### 1. Configuration

Copy the example environment file and fill in your credentials:

```bash
cp .env.example .env
```

Required environment variables:

- `TRANSPORTAPI_APP_ID`: Your TransportAPI Application ID.
- `TRANSPORTAPI_APP_KEY`: Your TransportAPI Application Key.
- `ENVIRONMENT`: Set to `development` or `production`.

### 2. Import NaPTAN Data

Before running the server, you need to import the bus stop data from GOV.UK:

```bash
go run main.go import
```

By default, this will create a SQLite database at `./data/next_departures.db`.

### 3. Run the API Server

Start the HTTP API server:

```bash
go run main.go api-server
```

The server will start on port `8080` by default.

## API Usage

### Search for stops by bounding box

Returns NaPTAN stops within the specified coordinates.

```http
GET /v1/next-departures/search?bbox=-1.565,53.961,-1.503,53.983
```

### Get next departures for a stop

Fetches real-time departures for the given NaPTAN stop ID.

```http
GET /v1/next-departures/490000235Z
```

### Reference Data: Stop Types

```http
GET /v1/next-departures/refdata/stop-types
```

### Health Check

```http
GET /healthz
```

### Metrics

```http
GET /metrics
```

## Commands

The application provides a few CLI commands:

- `api-server`: Starts the HTTP API server.
    - `--port`: Port to run on (default: 8080)
    - `--db`: Path to SQLite database (default: ./data/next_departures.db)
    - `--debug`: Enable pprof endpoints (warning: not for production)
- `import`: Performs a one-off import of bus stops from GOV.UK.
    - `--db`: Path to SQLite database.

## Development

### Running Tests

```bash
go test ./...
```

### Debugging

You can enable profiling endpoints by passing the `--debug` flag to the `api-server` command.

## References

- NaPTAN XML Schema: http://naptan.dft.gov.uk/naptan/schema/2.5/napt/NaPT_stop-v2-5.xsd
- NaPTAN Schema Guide (PDF): https://naptan.dft.gov.uk/naptan/schema/2.4/doc/NaPTANSchemaGuide-2.4-v0.57.pdf
- NaPTAN Guide for Data Managers on GOV.UK: https://www.gov.uk/government/publications/national-public-transport-access-node-schema/naptan-guide-for-data-managers
