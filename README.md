# gron

Make JSON and YAML greppable!

`gron` transforms JSON or YAML into discrete assignments to make it easier to `grep` for what you want and see the absolute "path" to it.

```console
$ gron "https://api.github.com/repos/tomnomnom/gron/commits?per_page=1" | fgrep "commit.author"
json[0].commit.author = {};
json[0].commit.author.date = "2016-07-02T10:51:21Z";
json[0].commit.author.email = "mail@tomnomnom.com";
json[0].commit.author.name = "Tom Hudson";
```

`gron` can work backwards too, enabling you to turn your filtered data back into JSON.
To illustrate, add `gron --ungron` to the end of the above pipeline:

```console
$ gron "https://api.github.com/repos/tomnomnom/gron/commits?per_page=1" | fgrep "commit.author" | gron --ungron
[
  {
    "commit": {
      "author": {
        "date": "2016-07-02T10:51:21Z",
        "email": "mail@tomnomnom.com",
        "name": "Tom Hudson"
      }
    }
  }
]
```

## Installation

`gron` has no runtime dependencies. You can just [download a binary for Linux, Mac, Windows or FreeBSD and run it](https://github.com/lafrenierejm/gron/releases).
Put the binary in your `PATH` (e.g. in `/usr/local/bin`) to make it easy to use:

```shell
tar xzf gron-linux-amd64-0.1.5.tgz
sudo mv gron /usr/local/bin/
```

Or if you're a Go user you can use `go install`:

```shell
go install github.com/tomnomnom/gron@latest
```

It's recommended that you alias `ungron` or `norg` (or both!) to `gron --ungron`.
Put something like this in your shell profile (e.g. in `~/.bashrc`):

```shell
alias norg="gron --ungron"
alias ungron="gron --ungron"
```

Or you could create a shell script in your `PATH` named `ungron` or `norg` to affect all users:

```shell
gron --ungron "$@"
```

## Usage

<details open>
<summary><code>gron</code> supports reading JSON from three different input methods.</summary>

1. Local file

   ```console
   $ gron ./testdata/two.json
   json = {};
   json.contact = {};
   json.contact.email = "mail@tomnomnom.com";
   json.contact.twitter = "@TomNomNom";
   json.github = "https://github.com/tomnomnom/";
   json.likes = [];
   json.likes[0] = "code";
   json.likes[1] = "cheese";
   json.likes[2] = "meat";
   json.name = "Tom";
   ```

1. URL

   ```console
   $ gron http://headers.jsontest.com/
   json = {};
   json.Host = "headers.jsontest.com";
   json["User-Agent"] = "gron/0.1";
   json["X-Cloud-Trace-Context"] = "6917a823919477919dbc1523584ba25d/11970839830843610056";
   ```

1. Standard input (<code>stdin</code>)

   ```console
   $ curl -s http://headers.jsontest.com/ | gron
   json = {};
   json.Accept = "*/*";
   json.Host = "headers.jsontest.com";
   json["User-Agent"] = "curl/7.43.0";
   json["X-Cloud-Trace-Context"] = "c70f7bf26661c67d0b9f2cde6f295319/13941186890243645147";
   ```

</details>

<details open>
<summary>Grep for something and easily see the path to it.</summary>

```console
$ gron testdata/two.json | grep twitter
json.contact.twitter = "@TomNomNom";
```

</details>

<details open>
<summary>Easily diff JSON.</summary>

```console
$ diff <(gron two.json) <(gron two-b.json)
3c3
< json.contact.email = "mail@tomnomnom.com";
---
> json.contact.email = "contact@tomnomnom.com";
```

</details>

<details open>
<summary>The output of <code>gron</code> is valid JavaScript.</summary>

```console
$ gron ./testdata/two.json > tmp.js && echo "console.log(json);" >> tmp.js && nodejs tmp.js
{ contact: { email: 'mail@tomnomnom.com', twitter: '@TomNomNom' },
  github: 'https://github.com/tomnomnom/',
  likes: [ 'code', 'cheese', 'meat' ],
  name: 'Tom' }
```

</details>

<details open>
<summary>Obtain <code>gron</code>'s output as a JSON stream via the <code>--json</code> switch.</summary>

```console
$ curl -s http://headers.jsontest.com/ | gron --json
[[],{}]
[["Accept"],"*/*"]
[["Host"],"headers.jsontest.com"]
[["User-Agent"],"curl/7.43.0"]
[["X-Cloud-Trace-Context"],"c70f7bf26661c67d0b9f2cde6f295319/13941186890243645147"]
```

</details>

## ungronning

<details open>
<summary><code>gron</code> can also turn its output back into JSON.</summary>

```console
$ gron testdata/two.json | gron -u
{
  "contact": {
    "email": "mail@tomnomnom.com",
    "twitter": "@TomNomNom"
  },
  "github": "https://github.com/tomnomnom/",
  "likes": [
    "code",
    "cheese",
    "meat"
  ],
  "name": "Tom"
}
```

</details>

<details open>
<summary>This means you use can use <code>gron</code> with <code>grep</code> and other tools to modify JSON.</summary>

```console
$ gron testdata/two.json | grep likes | gron --ungron
{
  "likes": [
    "code",
    "cheese",
    "meat"
  ]
}
```

</details>

<details open>
<summary>To preserve array keys, arrays are padded with <code>null</code> when values are missing.</summary>

To demonstrate, first create a shell pipeline that filters out an array entry:

```console
$ gron testdata/two.json | grep likes | grep -v cheese
json.likes = [];
json.likes[0] = "code";
json.likes[2] = "meat";
```

> **Note** that the array indices jump directly from 0 to 2 because we're excluding the `"cheese"` value via `grep -v`.

Now add `gron --ungron` to the end of the above pipeline to get JSON output:

```console
$ gron testdata/two.json | grep likes | grep -v cheese | gron --ungron
{
  "likes": [
    "code",
    null,
    "meat"
  ]
}
```

> **Note** the `null` placeholder has been inserted to account for the excluded `"cheese"` value.

</details>

If you get creative you can do [some pretty neat tricks with gron](ADVANCED.mkd), and then ungron the output back into JSON.

## Get Help

Run `gron --help` for help:

<!-- `$ gron --help` as txt -->
```txt
gron transforms JSON or YAML (from a file, URL, or stdin) into discrete assignments to make it easier to grep for what you want and see the absolute "path" to it.

Examples:
  gron /tmp/apiresponse.json
  gron http://jsonplaceholder.typicode.com/users/1
  curl -s http://jsonplaceholder.typicode.com/users/1 | gron
  gron http://jsonplaceholder.typicode.com/users/1 | grep company | gron --ungron

Usage:
  gron [flags]

Flags:
  -c, --colorize     Colorize output (default on TTY)
  -h, --help         help for gron
  -k, --insecure     Disable certificate validation when reading from a URL
  -j, --json         Represent gron data as JSON stream
  -m, --monochrome   Do not colorize output
      --sort         Sort output
  -s, --stream       Treat each line of input as a separate JSON object
  -u, --ungron       Reverse the operation (turn assignments back into JSON)
  -v, --values       Print just the values of provided assignments
      --version      Print version information
  -y, --yaml         Treat input as YAML instead of JSON
```

## FAQ

### Wasn't this written in PHP before?

Yes it was!
The original version is [preserved here for posterity](https://github.com/tomnomnom/gron/blob/master/original-gron.php).

### Why the change to Go?

Mostly to remove PHP as a dependency.
There are a lot of people who work with JSON who don't have PHP installed.

### Why shouldn't I just use `jq`?

[`jq`](https://stedolan.github.io/jq/) is _awesome_, and a lot more powerful than gron, but with that power comes complexity.
`gron` aims to make it easier to use the tools you already know, like `grep` and `sed`.

`gron`'s primary purpose is to make it easy to find the path to a value in a deeply nested JSON blob when you don't already know the structure;
much of `jq`'s power is unlocked only once you know that structure.
