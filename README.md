### mitum-currency
*mitum-currency* is the cryptocurrency case of mitum model, based on
[*mitum*](https://github.com/spikeekips/mitum). This project was started for
creating the first model case of *mitum*, but it can be used for simple
cryptocurrency blockchain network (at your own risk).

~~For more details, see the [documentation](https://mitum-currency-doc.readthedocs.io/en/latest/?badge=master).~~

#### Features,

* account: account address and keypair is not same.
* simple transaction: creating account, transfer balance.
* *mongodb*: as mitum does, *mongodb* is the api server storage.
* supports multiple currencies

#### Installation

> NOTE: at this time, *mitum* and *mitum-currency* is actively developed, so
before building mitum-currency, you will be better with building the latest
mitum source.
> `$ git clone https://github.com/spikeekips/mitum`
>
> and then, add `replace github.com/spikeekips/mitum => <your mitum source directory>` to `go.mod` of *mitum-currency*.

Build it from source
```sh
$ cd mitum-currency
$ go build -ldflags="-X 'main.Version=v0.0.1'" -o ./mc ./main.go
```

#### Run

At the first time, you can simply start node with example configuration.

> To start, you need to run *mongodb* on localhost(port, 27017).

```
$ ./mc init --design=./standalone.yml genesis-design.yml
$ ./mc run --design=./standalone.yml
```

> Please check `$ ./mc --help` for detailed usage.