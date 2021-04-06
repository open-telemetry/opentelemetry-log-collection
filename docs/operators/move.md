## `move` operator

The `move` operator moves (or renames) a field from one location to another.

It's configured by passing 'to' and 'from' fields.

### Configuration Fields

| Field      | Default          | Description                                                                                                                                                                                                                              |
| ---        | ---              | ---                                                                                                                                                                                                                                      |
| `id`       | `restructure`    | A unique identifier for the operator                                                                                                                                                                                                     |
| `output`   | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                                                                                                                         |
| `from`      | required       | The field to move the value out of.   
| `to`      | required       | The field to move the value into.
| `on_error` | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                                                                                                                          |
| `if`       |                  | An [expression](/docs/types/expression.md) that, when set, will be evaluated to determine whether this operator should be used for the given entry. This allows you to do easy conditional parsing without branching logic with routers. |

Example usage:
```yaml
- type: move
    from: key1
    to: key3
```

<table>
<tr><td> Input Body </td> <td> Output Body </td></tr>
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
  "key3": "val1",
  "key2": "val2"
}
```

</td>
</tr>
</table>