### Description

Polkadothub Indexer project is responsible for fetching and indexing Polkadot data.

### Internal dependencies:
This package connects via gRPC to a polkadothub-proxy which in turn connects to Polkadot node.

### External Packages:
* `polkadothub-proxy` - Go proxy to Polkadot node
* `indexing-engine` - A backbone for indexing process
* `gin` - Http server
* `gorm` - ORM with PostgreSQL interface
* `cron` - Cron jobs runner
* `zap` - logging 

### Environmental variables:

* `APP_ENV` - application environment (development | production) 
* `PROXY_URL` - url to polkadothub-proxy
* `SERVER_ADDR` - address to use for API
* `SERVER_PORT` - port to use for API
* `FIRST_BLOCK_HEIGHT` - height of first block in chain
* `INDEX_WORKER_INTERVAL` - index interval for worker
* `SUMMARIZE_WORKER_INTERVAL` - summary interval for worker
* `PURGE_WORKER_INTERVAL` - purge interval for worker
* `DEFAULT_BATCH_SIZE` - syncing batch size. Setting this value to 0 means no batch size
* `DATABASE_DSN` - PostgreSQL database URL
* `DEBUG` - turn on db debugging mode
* `LOG_LEVEL` - level of log
* `LOG_OUTPUT` - log output (ie. stdout or /tmp/logs.json)
* `ROLLBAR_ACCESS_TOKEN` - Rollbar access token for error reporting
* `ROLLBAR_SERVER_ROOT` - Rollbar server root for error reporting
* `INDEXER_METRIC_ADDR` - Prometheus server address for indexer metrics 
* `SERVER_METRIC_ADDR` - Prometheus server address for server metrics 
* `METRIC_SERVER_URL` - Url at which metrics will be accessible (for both indexer and server)
* `PURGE_BLOCK_INTERVAL` - Block sequence older than given interval will be purged
* `PURGE_BLOCK_HOURLY_SUMMARY_INTERVAL` - Block hourly summary records older than given interval will be purged
* `PURGE_BLOCK_DAILY_SUMMARY_INTERVAL` - Block daily summary records older than given interval will be purged
* `PURGE_VALIDATOR_INTERVAL` - Validator sequence older than given interval will be purged
* `PURGE_VALIDATOR_HOURLY_SUMMARY_INTERVAL` - Validator hourly summary records older than given interval will be purged
* `PURGE_VALIDATOR_DAILY_SUMMARY_INTERVAL` - Validator daily summary records older than given interval will be purged
* `INDEXER_TARGETS_FILE` - JSON file with targets and its task names 

### Available endpoints:

| Method | Path                                 | Description                                                 | Params                                                                                                                                                |
|--------|------------------------------------  |-------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| GET    | `/health`                            | health endpoint                                             | -                                                                                                                                                     |
| GET    | `/status`                            | status of the application and chain                         | include_chain (bool, optional) -   when true, returns chain status                                                                                                                                             |
| GET    | `/block`                             | return block by height                                      | height (optional) - height [Default: 0 = last]                                                                                                        |
| GET    | `/block_times/:limit`                | get last x block times                                      | limit (required) - limit of blocks                                                                                                                    |
| GET    | `/blocks_summary`                    | get block summary                                           | interval (required) - time interval [hourly or daily] period (required) - summary period [ie. 24 hours]                                               |
| GET    | `/transactions`                      | get list of transactions                                    | height (optional) - height [Default: 0 = last]                                                                                                        |
| GET    | `/account/:stash_account`            | get account information for height                          | stash_account (required) - stash account  height (optional) - height [Default: 0 = last]                                                                  |
| GET    | `/account_details/:stash_account`    | get account details                                         | stash_account (required) - stash account                                                                                                                  |
| GET    | `/rewards/:stash_account`            | get daily rewards for account                               | stash_account (required), start (optional) - the starting era [Default: 1 = first], end (optional) - the ending era (if unspecified, returns latest)(optional)                                                                                                               |
| GET    | `/validators`                        | get list of validators                                      | height (optional) - height [Default: 0 = last]                                                                                                        |
| GET    | `/validators/for_min_height/:height` | get the list of validators for height greater than provided | height (required) - height [Default: 0 = last]                                                                                                        |
| GET    | `/validator/:stash_account`          | get validator by address                                    | stash_account (required) - validator's stash account    sessions_limit (required) - number of last sessions to include    eras_limit (required) - number of last eras to include                                                                                                      |
| GET    | `/validators_summary`                | validator summary                                           | interval (required) - time interval [hourly or daily] period (required) - summary period [ie. 24 hours]  stash_account (optional) - validator's stash account |
| GET    | `/system_events`                | get system events for validator                                  | after (optional) - height kind (optional) - system event kind [eg. "joined_set"]  |

### Running app

Once you have created a database and specified all configuration options, you
need to migrate the database. You can do that by running the command below:

```bash
polkadothub-indexer -config path/to/config.json -cmd=migrate
```

Start the data indexer:

```bash
polkadothub-indexer -config path/to/config.json -cmd=worker
```

Start the API server:

```bash
polkadothub-indexer -config path/to/config.json -cmd=server
```

IMPORTANT!!! Make sure that you have polkadothub-proxy running and connected to Polkadot node.

### Running one-off commands

Start indexer:
```bash
polkadothub-indexer -config path/to/config.json -cmd=indexer_start
```

Create summary tables for sequences:
```bash
polkadothub-indexer -config path/to/config.json -cmd=indexer_summarize
```

Purge old data:
```bash
polkadothub-indexer -config path/to/config.json -cmd=indexer_purge
```

### Running tests

To run tests with coverage you can use `test` Makefile target:
```shell script
make test
```

### Exporting metrics for scrapping
We use Prometheus for exposing metrics for indexer and for server.
Check environmental variables section on what variables to use to setup connection details to metrics scrapper.
We currently expose below metrics:
* `figment_indexer_height_success` (counter) - total number of successfully indexed heights
* `figment_indexer_height_error` (counter) - total number of failed indexed heights
* `figment_indexer_height_duration` (gauge) - total time required to index one height
* `figment_indexer_height_task_duration` (gauge) - total time required to process indexing task 
* `figment_indexer_use_case_duration` (gauge) - total time required to execute use case 
* `figment_database_query_duration` (gauge) - total time required to execute database query 
* `figment_server_request_duration` (gauge) - total time required to executre http request 


