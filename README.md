# Foreman
[![Build Status](https://travis-ci.org/0xef53/foreman.svg?branch=master)](https://travis-ci.org/0xef53/foreman)

Foreman is an NSQ manager that allows handling of job-messages asynchronously in multiple threads.

On start Foreman reads the config and subscribes to all topics it describes. For each topic a number of parameters can be defined. These parameters set rules on how incoming messages should be processed: which script the message should be passed to, concurrency settings, whether to repeat the job in case of failure or not, etc.

Only messages that are JSON-encoded objects are processed, others are ignored.

Foreman works in foreground, so the most convenient way to run it is via the supervisor. See [example run script](https://github.com/0xef53/foreman/tree/master/runit) for runit or [config sample](https://github.com/0xef53/foreman/tree/master/systemd/foreman.service) for systemd.

## Installation

**From source:**

```shell
mkdir foreman && cd foreman
GOPATH=$(pwd) go get -v -tags netgo -ldflags '-s -w' github.com/0xef53/foreman
```

**From deb package:**

You can download [the pre-built package](https://github.com/0xef53/foreman/releases) or [build your own](https://github.com/0xef53/foreman-debian).

## Configuration

Configuration file has the following structure:

```ini
[common]
  ; A name/id of this instance. The default is short hostname.
  client-id = "host-01"

  ; NSQLookupd addresses separated by comma or space.
  ; This is a general list. A personal list can be defined for each topic.
  servers = "127.0.0.1:4161 192.168.1.100:4161"

  ; General channel name. Can be redefined for each topic.
  ; The default is "foreman".
  channel = "foreman"

[topic "foobar"]
  ; Personal NSQLookupd addresses for this topic.
  ; It overrides the general list of NSQLookupd addresses.
  servers = "10.0.0.10:4161"

  ; Custom channel name. It overrides the value from the "common" section.
  channel = "foobar_foreman"

  ; Directory with worker script and notify hooks.
  workdir = "/opt/foreman-workers"

  ; Worker command to run for every message this topic receives.
  ; Message body is sent to the command over stdin.
  ; Command definition supports Golang templates. Variables are expanded
  ; from the message object attributes of the same names.
  ; For example, if there is a message of the following type:
  ; {"first_name": "Alice", "last_name": "Smith"}
  ; then command definition may look like this:
  cmd = "greeting-worker --name={{.first_name}} --last-name={{.last_name}}"

  ; Number of parallel worker processes for this topic. The default is 1.
  ; Be careful: if topic workers operate on the same data structures 
  ; you will need to take care of locks.
  concurrency = 5

  ; In case the worker process finished with the special 100 exit code, 
  ; the task is retried again. This directive defines maximum number of retries.
  ; When maximum number of attempts reached and worker still fails,
  ; `notify-fault` hook is called if set for current topic.
  max-attempts = 4

  ; There are 3 hooks that can be used in different situations:
  ;
  ;    notify-start    ; runs before the main command
  ;    notify-finish   ; runs after successful exit
  ;    notify-fault    ; runs in case of an error
  ;
  ; Message body is send to the process over stdin, Golang templates work for
  ; hook definitions as well.

  ; This hook starts after recieving the message, but before the main command is called.
  ; Hook exit status will be available via `_notify_start_exit_code` name.
  notify-start = "/bin/echo 'Got a new job ( topic = {{._topic_name}} )'"

  ; This hook runs after successful exit of the main command.
  notify-finish = "/bin/echo 'Successfully completed'"

  ; This hook is called if the last attempt to process task has failed.
  ; The exit status of the worker command will be available via `_worker_exit_code` name.
  notify-fault = "/bin/echo 'Something went wrong ( exit code {{._worker_exit_code}} )'"
```

### Special fields

For each message the following special fields are set (they can be referred to via templated command definitions):

| Field                     | Description
| ------------------------- | -----------------------------------
| `_topic_name`             | A topic name
| `_worker_exit_code`       | The exit code of the worker command
| `_notify_start_exit_code` | The exit code of notify-start hook

## License

This project is under the MIT License. See the [LICENSE](https://github.com/0xef53/foreman/blob/master/LICENSE) file for the full license text.
