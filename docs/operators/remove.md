## `remove` operator

The `remove` operator removes a field from a record.

It's configured by passing a field to remove.

### Configuration Fields

| Field      | Default          | Description                                                                                                                                                                                                                              |
| ---        | ---              | ---                                                                                                                                                                                                                                      |
| `id`       | `restructure`    | A unique identifier for the operator                                                                                                                                                                                                     |
| `output`   | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                                                                                                                         |
| `field`      | required       | The [field](/docs/types/field.md) to remove.
| `on_error` | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                                                                                                                          |
| `if`       |                  | An [expression](/docs/types/expression.md) that, when set, will be evaluated to determine whether this operator should be used for the given entry. This allows you to do easy conditional parsing without branching logic with routers. |

Example usage:

<hr>

Remove value from body
```yaml
- type: remove 
    field: 
    - key1
```

<table>
<tr><td> Input Entry </td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": {
    "key1": "val1",
  }
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": { }
}
```

</td>
</tr>
</table>

<hr>

Remove object from body
```yaml
- type: remove 
    field: 
    - object
```

<table>
<tr><td> Input Entry </td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": {
    "object": {
      "nestedkey": "nestedval"
    }
  },
  "key": "val"
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": { 
     "key": "val"
  }
}
```

</td>
</tr>
</table>

<hr>

Remove multiple fields from body
```yaml
- type: remove 
    field: 
    - object
    - key
```

<table>
<tr><td> Input Entry </td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": {
    "object": {
      "nestedkey": "nestedval"
    },
    "key": "val"
  },
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": { }
}
```

</td>
</tr>
</table>

<hr>

Remove Value from attributes
```yaml
- type: remove 
    field: 
    - $attributes.otherkey
```

<table>
<tr><td> Input Entry </td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { 
    "otherkey": "val"
  },  
  "body": {
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  },
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": {  },  
  "body": { 
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  }
}
```

</td>
</tr>
</table>

<hr>

Remove multiple values from attributes
```yaml
- type: remove 
    field: 
    - $attributes.otherkey
    - $attributes.otherkey2
```

<table>
<tr><td> Input Entry </td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { },
  "attributes": { 
    "otherkey": "val",
    "otherkey2": "val2"
  },  
  "body": {
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  },
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": {  },  
  "body": { 
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  }
}
```

</td>
</tr>
</table>

<hr>

Remove Value from resource
```yaml
- type: remove 
    field: 
    - $resource.otherkey
```

<table>
<tr><td> Input Entry </td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { 
    "otherkey": "val"
  },
  "attributes": {  },  
  "body": {
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  },
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": {  },  
  "body": { 
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  }
}
```

</td>
</tr>
</table>

<hr>

Remove multiple values from attributes
```yaml
- type: remove 
    field: 
    - $resource.otherkey
    - $resource.otherkey2
```

<table>
<tr><td> Input Entry </td> <td> Output Entry </td></tr>
<tr>
<td>

```json
{
  "resource": { 
    "otherkey": "val",
    "otherkey2": "val2"
  },
  "attributes": {  },  
  "body": {
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  },
}
```

</td>
<td>

```json
{
  "resource": { },
  "attributes": { },  
  "body": { 
    "object": {
      "nestedkey": "nestedval"
    }, 
    "key": "val"
  }
}
```

</td>
</tr>
</table>