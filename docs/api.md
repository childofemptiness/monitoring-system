# API Contracts

## Create Monitor

### Use case
Client creates a new URL monitor for periodic availability checks.

### Endpoint

`POST /monitors`

### Request

```json
{
    "url": "https://example.com",
    "interval_seconds": 60
}
```

**Success response**

Status: ```201 Created```
```json
{
    "id": 1,
    "url": "https://example.com",
    "interval_seconds": 60
}
```
**Error responses**
- ```400 Bad Request``` if JSON is invalid or fields are invalid
- ```409 Conflict``` if a monitor with the same URL already exists
- ```500 Internal Server Error```  for unexpeceted internal errors

**Business rules**
- ```url``` is required
- ```url``` must be a valid ```http``` or ```https``` URL
- ```interval_seconds``` must ne a greater than ```0```
- duplicate monitors for the same URL are not alllowed

**Service contract**

```go
type CreateMonitorInput struct {
    URL              string
    InternvalSeconds int
}

func (s *Service) CreateMonitor(ctx context.Context, input CreateMonitorInput) (Monitor, Error)
```