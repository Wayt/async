# async

Async is a function worker scheduling framework.

This is currently a proof of concept and is prone to changes.

Any feedback are warmly welcome :)

## Glossary

* `function`: A function is defined by client, and run code. It has a `name`, arguments and retry options.
* `job`: A job is a scheduled task with a list of `function`, it has an `id`, global parameter (`data`) and a `state`.
* `worker`: And async client that register himself on a server with the list of function it is able to run.
* `server`: The central component that receive job and dispatch execution accross workers.

## Getting started

An example is available using docker, see [example](example/worker/).

Tested with:
* Docker version 17.07.0-ce
* docker-compose version 1.16.1
* [httpie](https://github.com/jakubroztocil/httpie):


You can execute it:
```bash
# go into example directory
async# cd example/

# build images
example# docker-compose build
[...]
Successfully tagged example_server:latest
[...]
Successfully tagged example_worker:latest

# Run example with 2 workers
example# docker-compose up -d --scale worker=2 --scale server=1
Creating network "example_default" with the default driver
Creating example_server_1 ...
Creating example_server_1 ... done
Creating example_worker_1 ...
Creating example_worker_1 ... done
Creating example_worker_1
Creating example_worker_2 ... done

# check logs
example# docker-compose logs -f

# schedule a job
example# http POST 127.0.0.1:8000/v1/job < worker/job_chain.json
HTTP/1.1 200 OK
[...]
    "job_id": "a2160ebb-81be-4a46-a7dd-b665ac5c839f",
}

# retrieve job information
example# http GET 127.0.0.1:8000/v1/job/a2160ebb-81be-4a46-a7dd-b665ac5c839f
HTTP/1.1 200 OK
Content-Length: 438
Content-Type: text/plain; charset=utf-8
Date: Mon, 21 May 2018 23:38:21 GMT

{
    "Job": {
        "created_at": "2018-05-21T23:37:14.288834016Z",
        "current_function": 2,
        "data": null,
        "functions": [
            {
                "name": "/v1/test-1",
                "retry_count": 1,
                "retry_options": {
                    "retry_limit": 3
                }
            },
            {
                "name": "/v1/test-2",
                "retry_count": 1,
                "retry_options": {
                    "retry_limit": 3
                }
            },
            {
                "name": "/v1/say-hello-world",
                "retry_count": 1,
                "retry_options": {
                    "retry_limit": 3
                }
            }
        ],
        "job_id": "a2160ebb-81be-4a46-a7dd-b665ac5c839f",
        "name": "test",
        "scheduled_at": "2018-05-21T23:37:15.291637675Z"
    }
}
```

## Licence

See [LICENCE](LICENCE)
