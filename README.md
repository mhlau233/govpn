# govpn

a simple implementation of vpn in 250 lines of code.

## Installation

govpn requires waterto setup tun interface.

```bash
git clone https://github.com/dllexport/govpn
cd govpn
go get github.com/songgao/water
go build
```

## Usage

govpn server runs on linux and client runs on macos, change the setup shell script if you want to run on other os.

```
  -cip string
        client tun ip (default "10.0.0.100")
  -k string
        encryption key (default "abcdefgqywuwyw")
  -m string
        run mode
  -se string
        server endpoint
  -sip string
        server tun ip (default "10.0.0.1")
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)