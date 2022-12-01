# Overview

Hush Lightwalletd is a fork of [lightwalletd](https://github.com/adityapk00/lightwalletd) original from Zcash (ZEC).

It is a backend service that provides a bandwidth-efficient interface to the Hush blockchain for [SilentDragonLite cli](https://git.hush.is/hush/silentdragonlite-light-cli) and [SilentDragonLite](https://git.hush.is/hush/SilentDragonLite).

## Changes from upstream lightwalletd
This version of lightwalletd extends lightwalletd and:

* Adds support for HUSH
* Adds support for transparent addresses
* Adds several new RPC calls for lightclients
* Lots of perf improvements
  * Replaces SQLite with in-memory cache for Compact Blocks
  * Replace local Txstore, delegating Tx lookups to hushd
  * Remove the need for a separate ingestor

## Running your own SDL lightwalletd

#### 0. First, install Go
You will need Go >= 1.17 which you can download from the official [download page](https://golang.org/dl/) or install via your OS package manager.
Most OS package managers will not have such a new version, but you might get lucky.

This [installation](https://golang.org/doc/install) document shows how to do it on various OS's.

If you're using Ubuntu or Debian, try:

```
$ sudo apt install golang
```

#### 1. Run a Hush node.
Either compile or build the [Hush Daemon (hushd)](https://git.hush.is/hush/hush3).

Next, change your HUSH3.conf file to something like the following:

```
rpcuser=user-CHANGETHIS
rpcpassword=pass-CHANGETHIS
rpcport=18031 # this if for HUSH, change it for other HSC's
server=1
txindex=1
addressindex=1 # required for the newest lightwalletd code
rpcworkqueue=256
rpcallowip=127.0.0.1
rpcbind=127.0.0.1
```

Then start `hushd` in your command window. You might need to run with `-reindex` the first time if you are enabling `-addressindex` option for the first time. The reindex might take a while. A fresh sync is usually the fastest way to enable `-addressindex`, instead of doing a reindex.

#### 2. Compile lightwalletd
Run the build script.

```
make
```

#### 3. Get a TLS certificate and run the Lightwalletd frontend
First, get a TLS certificate:

On Ubuntu Linux, **I SUGGEST YOU DO NOT USE SNAPD** and just ```sudo apt install certbot``` and then start on [Step 7 of these instructions by the EFF](https://certbot.eff.org/instructions)

Next you decide how you want to setup lightwalletd - with (Option A) or without NGINX (Option B).

##### Option A: "Let's Encrypt" certificate using NGINX as a reverse proxy
If you running a public-facing server, the easiest way to obtain a certificate is to use a NGINX reverse proxy and get a Let's Encrypt certificate.

Create a new section for the NGINX reverse proxy:
```
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    
    server_name your_host.net;

    ssl_certificate /etc/letsencrypt/live/your_host.net/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/your_host.net/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
        
    location / {
        # Replace 9067 with the port of your gRPC server if using a custom port
        # Hush Smart Chains should use a different port than 9067 so it doesn't conflict with HUSH lightwalletd
        # NOTE: it's only safe to use --no-tls on lightwalletd if this is on localhost
    	grpc_pass grpc://localhost:9067;
    }
}
```

Then run the lightwalletd frontend with the following:

```
./start.sh
```

Note: we use the "--no-tls" option as we are using NGINX as a reverse proxy and letting it handle the TLS authentication for us instead. If you want to do TLS directly with lightwalletd with no reverse proxy, see the next section.


##### Option B: "Let's Encrypt" certificate just using lightwalletd without NGINX
The other option is to configure lightwalletd to handle its own TLS authentication. Once you have a certificate that you want to use (from a certificate authority), pass the certificate to the frontend as follows:

```
./start-tls.sh -tls-cert /etc/letsencrypt/live/YOURWEBSITE/fullchain.pem -tls-key /etc/letsencrypt/live/YOURWEBSITE/privkey.pem
```

#### 4. Point the `silentdragonlite-cli` to this server
You should start seeing the frontend ingest and cache the Hush blocks after ~15 seconds. 

Now, connect to your server! (Substitute with your own below)
```
git clone https://git.hush.is/hush/silentdragonlite-cli
cd silentdragonlite-cli
cargo build --release
./target/release/silentdragonlite-cli --server https://lite.example.org
```

* If you have trouble compiling silentdragonlite-cli, then [please refer to it's separate documentation here](https://git.hush.is/hush/silentdragonlite-cli) on how to build it and what pre-requisites need to be installed.

## Lightwalletd Command-line Options

These are some of the most used command line options for lightwalletd:


| CLI option | Default                   | What it does                  |
|------------|:--------------:|------------------------------:|
| --grpc-bind-addr | 127.0.0.1:9067 | address and port to listen on |
| --tls-cert  | blank             | the path to a TLS certificate |
| --tls-key   | blank             | the path to a TLS key file    |
| --no-tls    | false             | Disable TLS, serve un-encrypted traffic |
| --log-file  | blank             | log file to write to                  |
| --log-level | logrus.InfoLevel | log level 1 thru 7 (something from logrus)|
| --hush-conf-path | blank             | conf file to pull RPC creds from |
| --cache-size| 40000             | number of blocks to hold in the cache |


Run `./lightwalletd --help` to see all options.

## Developing

To create a `foo.pb.go` file from a `foo.proto` file:

```
protoc --go_out=paths=source_relative:. foo.proto
```

Or do `make protobuf`

## Support
For support or other questions, join us on [Telegram](https://hush.is/telegram), or tweet at [@HushIsPrivacy](https://twitter.com/HushIsPrivacy), or toot at our [Mastodon](https://fosstodon.org/@myhushteam) or join [Telegram Support](https://hush.is/telegram_support).

## License
GPLv3 or later

# Copyright

2016-2022 The Hush Developers
