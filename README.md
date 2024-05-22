# slr: SikaLabs Random Utils

`slr` is very similar tool to `slu` ([sikalabs/slu](https://github.com/sikalabs/slu)) but it's for random stuff that doesn't fit into `slu`. We dont want to put random hacks to `slu` so we created `slr` to be able to easily distribute our random scripts, tools and utils.

If something is useful for more cases we will move it to `slu`.

## Install

Linux AMD64

```
curl -fsSL https://raw.githubusercontent.com/sikalabs/slr/master/install.sh | sudo sh
```

Using [slu](https://github.com/sikalabs/slu)

```
slu install-bin slr
```

Install on Mac

```
brew install sikalabs/tap/slr
```

## Contributing

You can create new command from [cmd/example](./cmd/example/example.go).
