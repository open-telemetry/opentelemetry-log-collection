## `recombine` operator

The `recombine` operator combines consecutive logs into single logs based on simple expression rules.

### Configuration Fields

| Field            | Default          | Description |
| ---              | ---              | ---         |
| `id`             | `recombine`      | A unique identifier for the operator. |
| `output`         | Next in pipeline | The connected operator(s) that will receive all outbound entries. |
| `on_error`       | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md). |
| `is_first_entry` |                  | An [expression](/docs/types/expression.md) that returns true if the entry being processed is the first entry in a multiline series. |
| `is_last_entry`  |                  | An [expression](/docs/types/expression.md) that returns true if the entry being processed is the last entry in a multiline series. |
| `combine_field`  | required         | The [field](/docs/types/field.md) from all the entries that will recombined. |
| `combine_with`   | `"\n"`           | The string that is put between the combined entries. This can be an empty string as well. When using special characters like `\n`, be sure to enclose the value in double quotes: `"\n"`. |
| `max_batch_size` | 1000             | The maximum number of consecutive entries that will be combined into a single entry. |
| `overwrite_with` | `oldest`         | Whether to use the fields from the `oldest` or the `newest` entry for all the fields that are not combined. |

Exactly one of `is_first_entry` and `is_last_entry` must be specified.

NOTE: this operator is only designed to work with a single input. It does not keep track of what operator entries are coming from, so it can't combine based on source.

### Example Configurations

#### Recombine Kubernetes logs in the CRI format

Kubernetes logs in the CRI format have a tag that indicates whether the log entry is part of a longer log line (P) or the final entry (F). Using this tag, we can recombine the CRI logs back into complete log lines.

Configuration:

```yaml
- type: file_input
  include:
    - ./input.log
- type: regex_parser
  regex: '^(?P<timestamp>[^\s]+) (?P<stream>\w+) (?P<logtag>\w) (?P<message>.*)'
- type: recombine
  combine_field: message
  combine_with: ""
  is_last_entry: "$body.logtag == 'F'"
  overwrite_with: "newest"
```

Input file:

```
2016-10-06T00:17:09.669794202Z stdout F Single entry log 1
2016-10-06T00:17:10.113242941Z stdout P This is a very very long line th
2016-10-06T00:17:10.113242941Z stdout P at is really really long and spa
2016-10-06T00:17:10.113242941Z stdout F ns across multiple log entries
```

Output logs:

```json
[
  {
    "timestamp": "2020-12-04T13:03:38.41149-05:00",
    "severity": 0,
    "body": {
      "message": "Single entry log 1",
      "logtag": "F",
      "stream": "stdout",
      "timestamp": "2016-10-06T00:17:09.669794202Z"
    }
  },
  {
    "timestamp": "2020-12-04T13:03:38.411664-05:00",
    "severity": 0,
    "body": {
      "message": "This is a very very long line that is really really long and spans across multiple log entries",
      "logtag": "F",
      "stream": "stdout",
      "timestamp": "2016-10-06T00:17:10.113242941Z"
    }
  }
]
```
