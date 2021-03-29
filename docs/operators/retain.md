## `Retain` operator

The `retain` operator keeps the specified list of fields, and removes the rest.

The operator is configured with a list fields to be kept.

### Configuration Fields

| Field      | Default          | Description                                                                                                                                                                                                                              |
| ---        | ---              | ---                                                                                                                                                                                                                                      |
| `id`       | `restructure`    | A unique identifier for the operator                                                                                                                                                                                                     |
| `output`   | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                                                                                                                         |
| `fields`      | required         | A list of fields to be kept.                                                                                                                                                     |
| `on_error` | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                                                                                                                          |
| `if`       |                  | An [expression](/docs/types/expression.md) that, when set, will be evaluated to determine whether this operator should be used for the given entry. This allows you to do easy conditional parsing without branching logic with routers. |

Example usage:
```yaml
- type: retain
    fields:
      - "key1"
      - "key2"
```

<table>
<tr><td> Input record </td> <td> Output record </td></tr>
<tr>
<td>

```json
{
  "key1": "val1",
  "key2": "val2",
  "key3": "val3",
  "key4": "val4"
}
```

</td>
<td>

```json
{
  "key1": "val1",
  "key2": "val2"
}
```

</td>
</tr>
</table>