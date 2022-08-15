# notion2ical

Simple web service to expose Notion database as iCalendar compatible URLs.

## Contributing

**Requirements:**
- [go](https://go.dev)
- [make](https://www.gnu.org/software/make/) (quality of life)
- [docker](https://docker.com)/[podman](https://podman.io) (for containers)
- [goreleaser](https://goreleaser.com) (to build the code)

We provide a useful `Makefile` to execute common tasks:
- Run the server locally: `make quick-run`
- Run the test & coverage suite: `make test`
- Lint the code: `make lint`
- Format the code `make format`
- Build the code: `make build`
- Clean all created files: `make clean`

## Notion integrations

Currently only the private integration is supported, so only the workspaces which the user api key is an admin for will be supported.

The **public** integration type is [currently in the works](https://github.com/fmartingr/notion2ical/pull/1).

## Configuration

Service configuration is done using environment variables. All variables should be prefixed by `NOTION2ICAL_` when using kubernetes deployments.

From [internal/config/config.go](./internal/config/config.go)

| Name                                   | Type         | Description                                                                  |
| -------------------------------------- | ------------ | ---------------------------------------------------------------------------- |
| `HOSTNAME`                             | string       | Should be automatically filled                                               |
| `LOG_LEVEL`                            | string       | [Log level](https://github.com/uber-go/zap/blob/master/zapcore/level.go#L34) |
| `HTTP_ENABLED`                         | bool         | Enable/Disable the HTTP Server                                               |
| `HTTP_PORT`                            | int          | Port for the HTTP server to listen                                           |
| `HTTP_PUBLIC_HOSTNAME`                 | string       | Hostname used publicly when the service is released                          |
| `HTTP_BODY_LIMIT`                      | int          | Body limit in length                                                         |
| `HTTP_READ_TIMEOUT`                    | duration[^1] | Request read timeout                                                         |
| `HTTP_WRITE_TIMEOUT`                   | duration[^1] | Request write timeout                                                        |
| `HTTP_IDLE_TIMEOUT`                    | duration[^1] | Request IDLE timeout                                                         |
| `HTTP_DISABLE_KEEP_ALIVE`              | bool         | Enable/Disable keep alive support                                            |
| `HTTP_DISABLE_PARSE_MULTIPART_FORM`    | bool         | Enable/Disable parsing multipart form early                                  |
| `BRANDING_THANKS_MESSAGE`              | string       | Message shown on the final configuration step                                |
| `BRANDING_FOOTER_EXTRA`                | string       | Extra footer content                                                         |
| `NOTION_INTEGRATION_TOKEN`             | string       | The Notion integration token                                                 |
| `NOTION_MAX_PAGINATION`                | int          | The maximum number of pages to retrieve from a database                      |
| `ROUTES_CACHE_EXPIRATION`              | duration[^1] | Cache TTL for the generated calendars                                        |
| `ROUTES_CACHE_CONTROL`                 | bool         | Enable cache-control header                                                  |
| `ROUTES_CALENDAR_LIMITER_MAX_REQUESTS` | int          | Maximum requests number for the calendar endpoints                           |
| `ROUTES_CALENDAR_LIMITER_DURATION`     | duration[^1] | Maximum requests interval for the calendar endpoints                         |
| `ROUTES_STATIC_PATH`                   | string       | Path prefix for the static files                                             |
| `ROUTES_STATIC_MAX_AGE`                | duration[^1] | Max age for the served static files                                          |
| `ROUTES_SYSTEM_PATH`                   | string       | Path prefix for the system endpoints                                         |

[^1]: **`*duration`** = A string containing a number and the time unit: `10s` = 10 seconds, `1h` = 1 hour, ...
