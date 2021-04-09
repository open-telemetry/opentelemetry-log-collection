## `copy` operator

The `copy` operator copies a value from one [field](/docs/types/field.md) to another.

### Configuration Fields

| Field      | Default          | Description                                                                                                                                                                                                                              |
| ---        | ---              | ---                                                                                                                                                                                                                                      |
| `id`       | `copy`    | A unique identifier for the operator                                                                                                                                                                                                     |
| `output`   | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                                                                                                                         |
| `from`      | required       | The [field](/docs/types/field.md)  to copy the value of.   
| `to`      | required       | The [field](/docs/types/field.md)  to copy the value into.
| `on_error` | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                                                                                                                          |
| `if`       |                  | An [expression](/docs/types/expression.md) that, when set, will be evaluated to determine whether this operator should be used for the given entry. This allows you to do easy conditional parsing without branching logic with routers. |

Example usage:

Copy a value from the body to attributes
```yaml
- type: copy
    from: key2
    to: $attributes.newkey
```

<table>
<tr><td> Input Entry</td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": {
    "key1": "val1",
    "key2": "val2"
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": { 
      "newkey": "val2"
  },  
  "body": {
    "key3": "val1",
    "key2": "val2"
  }
}
```

</td>
</tr>
</table>

<hr>

Copy a value from attributes to the body
```yaml
- type: copy
    from: $attributes.key
    to: newkey
```

<table>
<tr><td> Input Entry</td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { 
      "key": "newval"
  },  
  "body": {
    "key1": "val1",
    "key2": "val2"
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": { 
      "key": "newval"
  },  
  "body": {
    "key3": "val1",
    "key2": "val2",
    "newkey": "newval"
  }
}
```

</td>
</tr>
</table>

<hr>

Copy a value from an object to the body
```yaml
- type: copy
    from: obj.nested
    to: newkey
```

<table>
<tr><td> Input Entry</td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": {
      "obj": {
        "nested":"nestedvalue"
    }
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": {
    "obj": {
        "nested":"nestedvalue"
    },
    "newkey":"nestedvalue"
  }
}
```

</td>
</tr>
</table>

Copy a value from resource to the body
```yaml
- type: copy
    from: $resource.key
    to: newkey
```

<table>
<tr><td> Input Entry</td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { 
    "key":"value"
  },
  "attributes": { },  
  "body": { }
}
```

</td>
<td>

```json
{
  "resource": { 
       "key":"value"
  },
  "attributes": { },  
  "body": {
    "newkey":"value"
  }
}
```

</td>
</tr>
</table>
