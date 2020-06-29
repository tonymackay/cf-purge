# Overview
`cf-purge` is a command line utility written in Go that purges URLs from Cloudflare's edge cache.

## Building
Clone the repo then run the following commands:

```
go mod download
go build
```

To assign a version when building run:

```
go build -ldflags=-X=main.version=v1.0.0-beta1
```


## Using
Assign the Cloudflare API Token and Zone ID to the environment variables as shown below:

```
export CF_API_TOKEN=<api_token> 
export CF_ZONE_ID=<zone_id>
```

*Note: The API Token needs to have the Zone.Cache Purge permission for the corresponding Zone ID.*

### Purge a single URL
Run the command with the `-url` option specifying the full URL:
```
cf-purge -url https://example.com
Purging URLs from Cloudflare's edge cache
purged:  [https://example.com]
```

### Purge multiple URLS
Create a file called `urls.txt` with multiple URLS, each on a new line:

```
https://example.com
https://example.com/hello-world/
https://example.com/hello-world/img.jpg
```

Now run the command with the `-file` option:

```
cf-purge -file urls.txt
Purging URLs from Cloudflare's edge cache
purged:  [https://example.com https://example.com/hello-world/ https://example.com/hello-world/img.jpg]
```

## License
[MIT License](LICENSE)
