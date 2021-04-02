## `add` operator

The `add` operator adds a field to a record. It must have a `field` key and exactly one of `value`.

`field` is a [field](/docs/types/field.md) that will be set to `value` or the evaluated expression.

`value` is a static string that will be added to each entry at the field defined by `field`. 
 Expressions can be used in this field through surrounding it with `EXPR()`



### Configuration Fields

| Field      | Default          | Description                                                                                                                                                                                                                              |
| ---        | ---              | ---                                                                                                                                                                                                                                      |
| `id`       | `restructure`    | A unique identifier for the operator                                                                                                                                                                                                     |
| `output`   | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                                                                                                                         |
| `field`      | required       | The field to be added.    
| `value`      | required       | The value of the field to be added.
| `on_error` | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                                                                                                                          |
| `if`       |                  | An [expression](/docs/types/expression.md) that, when set, will be evaluated to determine whether this operator should be used for the given entry. This allows you to do easy conditional parsing without branching logic with routers. |


Example usage:
```yaml
- type: add
    field: "key"
    value: "val"
```

<table>
<tr><td> Input record </td> <td> Output record </td></tr>
<tr>
<td>

```json
{}
```

</td>
<td>

```json
{
  "key1": "val1",
  "key2": "val1-suffix"
}
```

</td>
</tr>
</table>