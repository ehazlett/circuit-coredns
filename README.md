# CoreDNS Circuit Plugin
Experimental CoreDNS plugin to perform lookup and resolution via Circuit GRPC client.

# Build
To build use the following:

## Docker
The following will build CoreDNS with the Circuit plugin as a Docker image:

`make`

## Manual
To build, include this in the `plugin.cfg` in CoreDNS and `go build`.  For example:

```
...
circuit:github.com/ehazlett/circuit-coredns
```

Then you should be able to `go build` to build the custom `coredns`.

See [here](https://coredns.io/2017/03/01/how-to-add-plugins-to-coredns/) for details on plugins.

# Usage
Place this in your `corefile`:

```
.:1053 {
  log
  debug
  circuit 127.0.0.1:8080
  forward . 1.1.1.1 {
    except internal
  }
}
```

Where `127.0.0.1:8080` is the addr to the Circuit socket.  This will configure
 the Circuit plugin to service all ".internal" domains and forward all others to `1.1.1.1`.
