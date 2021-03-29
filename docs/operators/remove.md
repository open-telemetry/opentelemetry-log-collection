## `remove` operator

The `remove` operator removes a field from a record.

It's configured by passing a field to remove.

### Configuration Fields

| Field      | Default          | Description                                                                                                                                                                                                                              |
| ---        | ---              | ---                                                                                                                                                                                                                                      |
| `id`       | `restructure`    | A unique identifier for the operator                                                                                                                                                                                                     |
| `output`   | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                                                                                                                         |
| `field`      | required       | The field to remove.
| `on_error` | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                                                                                                                          |
| `if`       |                  | An [expression](/docs/types/expression.md) that, when set, will be evaluated to determine whether this operator should be used for the given entry. This allows you to do easy conditional parsing without branching logic with routers. |

Example usage:
```yaml
- type: remove 
    field: "key1"
```

<table>
<tr><td> Input record </td> <td> Output record </td></tr>
<tr>
<td>

```json
{
  "key1": "val1",
  "key2": "val2"
}
```

</td>
<td>

```json
{
  "key2": "val2"
}
```

</td>
</tr>
</table>