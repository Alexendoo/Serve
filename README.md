HTTP server for development - serve several specified directories
simulataniously

## Install

```
go get github.com/Alexendoo/Serve
```

[Or download a binary](https://github.com/Alexendoo/serve/releases)

## Usage

```
USAGE:
   serve [OPTION]... [DIR]...

OPTIONS:
       --host     --  bind to host (default: localhost)
   -i, --index    --  serve all paths to index if file not found
       --no-list  --  disable directory listings
   -p, --port     --  bind to port (default: 8080)
   -v, --verbose  --  display requests and responses
```


## Examples

Serve files from the current directory

```
serve
```

----

Utilise npm packages without having to type `../node_modules`

```
serve client node_modules
```

```ANTLR
client/
  index.html
node_modules/
  whatwg-fetch/
    fetch.js
```

```html
<!-- in client/index.html -->
<script src="whatwg-fetch/fetch.js">
```

---

Use a specified file for any non matching requests, for instance for HTML5 routing

```
serve -i index.html
```
