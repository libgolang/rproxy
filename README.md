# Simple Reverse Proxy


## Install
```
go get -u github.com:libgolang/rproxy
```


# Run

- Make a copy of the sample config file `config.example.toml`.
  E.g.:
  ```
  	cp config.example.toml config.toml
  ```
- Make changes appropriate changes to `[[proxy]]` sections.

- Run with `--config` flag:
  ```
  rproxy --config config.toml
  ```

