# Server API docs

------

### Submitting new job

<details>
    <summary><code>POST</code> <code><b>/submit</b></code> <code>(submits a new job to the scheduler)</code></summary>

##### Parameters

> | name | required | data type | description |
> | ---- |----------| --------- | ----------- |
> | body | yes      | object (JSON) | Job submission payload with Lua script and priority level (optional) |

##### JSON Body Fields

> | field     | type     | required | description                                            |
> |-----------|----------|----------|--------------------------------------------------------|
> | script    | string   | yes      | Lua code to execute on a worker                        |
> | priority  | integer  | no       | Job priority (low = 0, mid = 1, high = 2), Default: 0. |

##### Example Request body

```json
{
  "script": "print('wassup!')",
  "priority": 1
}
```

##### Responses

See `/status/{id}` endpoint response section.

</details>

### Get status of the job

<details>
    <summary><code>GET</code> <code><b>/status{id}</b></code> <code>(returns the current status of a submitted job)</code></summary>

##### Parameters

> | name | required | data type | description     |
> |------|----------|-----------|-----------------|
> | id   | yes      | integer   | Existing job id |

##### Example Request URL

```text
 http://localhost:8080/status/123
```

##### Responses

> | http code | content-type       | response    |
> |-----------|--------------------|-------------|
> | `200`     | `application/json` | JSON object |
> | `400`     | `application/json` | JSON object |
> | `500`     | `application/json` | JSON object |

##### Response JSON fields

###### Success Response (200)

> | Field   | Type    | Description                                                  |
> |---------|---------|--------------------------------------------------------------|
> | job_id  | integer | Unique ID assigned to the job                                |
> | status  | string  | Initial job status (PENDING, RUNNING, COMPLETED, FAILED)     |
> | result  | string  | Lua script's result, absent if job status is not "COMPLETED" |

```json
{
  "job_id": 123,
  "status": "PENDING",
  "result": "wassup!"
}
```

###### Error Response (400, 500)

> | Field | Type   | Description   |
> |-------|--------|---------------|
> | error | string | Error message |

```json
{
  "error": "clear error message"
}
```

</details>