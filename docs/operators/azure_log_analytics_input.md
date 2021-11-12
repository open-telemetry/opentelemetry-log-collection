## `azure_log_analytics_input` operator

The `azure_log_analytics_input` operator reads Azure Log Analytics logs from Azure Event Hub.

The `azure_log_analytics_input` operator will use the `timegenerated` field as the parsed entry's timestamp. The label `azure_log_analytics_table` is derived from the log's `type` field.

## Prerequisites

You must define a Log Analytics Export Rule using Azure CLI. Microsoft has documentation [here](https://docs.microsoft.com/en-us/azure/azure-monitor/logs/logs-data-export?tabs=portal)

### Configuration Fields

| Field               | Default                     | Description                                                                                   |
| ---                 | ---                         | ---                                                                                           |
| `id`                | `azure_log_analytics_input` | A unique identifier for the operator                                                          |
| `output`            | Next in pipeline            | The connected operator(s) that will receive all outbound entries                              |
| `namespace`         | required                    | The Event Hub Namespace                                                                       |
| `name`              | required                    | The Event Hub Name                                                                            |
| `group`             | required                    | The Event Hub Consumer Group                                                                  |
| `connection_string` | required                    | The Event Hub [connection string](https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-get-connection-string) |
| `prefetch_count`    | `1000`                      | Desired number of events to read at one time                                                  |
| `start_at`          | `end`                       | At startup, where to start reading events. Options are `beginning` or `end`                   |

### Example Configurations

#### Simple Azure Event Hub input

Configuration:
```yaml
pipeline:
- type: azure_log_analytics_input
  namespace: otel
  name: devel
  group: Default
  connection_string: 'Endpoint=sb://otel.servicebus.windows.net/;SharedAccessKeyName=dev;SharedAccessKey=supersecretkey;EntityPath=devel'
  start_at: end
```

### Example Output

A list of potential fields for each Azure Log Analytics table can be found [here](https://docs.microsoft.com/en-us/azure/azure-monitor/reference/tables/tables-category).
