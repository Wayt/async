# async

## Glossary

* `function`: A function is defined by client, and run code. It has a `name`, arguments and retry options.
* `job`: A job is a scheduled task with a list of `function`, it has an `id`, global parameter (`data`) and a `state`.

## Getting started

A sample server is available [main.md](examples/server/main.go)

You can execute it:
```bash
$ go run main.go
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /v1/job/:job_id           --> github.com/wayt/async/api.(*HttpJobHandler).Get-fm (3 handlers)
[GIN-debug] POST   /v1/job                   --> github.com/wayt/async/api.(*HttpJobHandler).Create-fm (3 handlers)
[GIN-debug] Listening and serving HTTP on :8080
[async] 2018/02/15 22:49:00 Having 4 function(s):
[async] 2018/02/15 22:49:00 	- /v1/say-hello-world
[async] 2018/02/15 22:49:00 	- /v1/test-1
[async] 2018/02/15 22:49:00 	- /v1/test-2
[async] 2018/02/15 22:49:00 	- /v1/test-fail
[async] 2018/02/15 22:49:00 Listening and serving HTTP on :8179
```

And create a job using [httpie](https://github.com/jakubroztocil/httpie):
```bash
$ http POST localhost:8080/v1/job name=test functions:='[{"name":"/v1/test-1"},{"name":"/v1/test-fail","args":[1,"2",3], "retry_count":0, "retry_options":{"retry_limit": 3}}, {"name":"/v1/test-2"}]'
```

## Job API

### API Reference

Both input parameters and output result are in json.

#### Types

StateEnum:
```
[
    "pending",
    "doing",
    "done",
    "failed"
]
```

Job:
```
{
    "id": "uuid",
    "name": "string",
    "functions": "[]Function",
    "state": "StateEnum"
    "data": "[string]object",
    "created_at": "time",
    "updated_at": "time"
}
```

Function:
```
{
    "name": "string",
    "args": "[]object",
    "retry_count": "int32",
    "retry_options": "RetryOption"
}
```

RetryOption:
```
{
    "retry_limit": "int32"
}
```

#### Create a job

`POST /v1/job`

Parameters:
* `name` (string): the job name
* `functions` (list of `Function`): Ordered list of function for this job

Result: `Job`

#### Get a job

`GET /v1/job/:job_id`

Parameters:
* `job_id` (uuid): job uuid

Result: `Job`

## Licence

See [LICENCE](LICENCE)
