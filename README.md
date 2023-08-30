# lensed
Edit complex files in place

By default edits JSON/YAML files, but supports more formats with "lenses".

Lenses can be nested. You can in-place "edit" a json field encoded as a string in a TOML file
that in turn is embedded in a string literal of a YAML file.

## Install

### Homebrew

TODO

### From sources

```bash
go install github.com/kubecfg/lensed@latest
```

## Usage

Let's start with a field replacement in YAML files.

### Basics

It can operate in a pipeline:

```console
$ cat >example.yaml <<EOF
foo:
  bar: old
EOF
$ lensed set /foo/bar=new <example.yaml
foo:
  bar: new
```

or modify files in place (like `sed -i``):

```console
$ cat >example.yaml <<EOF
foo:
  bar: old
EOF
$ lensed set /foo/bar=new -f example.yaml
$ cat example.yaml
foo:
  bar: new
```

### Preserves formatting

It preserves comments:

```console
$ cat >example.yaml <<EOF
foo:
  bar: old # comment preserved
EOF
$ lensed set /foo/bar=new <example.yaml
foo:
  bar: new # comment preserved
```

and other stylistic choices from the source file:

```console
$ cat >example.yaml <<EOF
foo:
  bar: 'old'
EOF
$ lensed set /foo/bar=new <example.yaml
foo:
  bar: 'new'
```

```console
$ cat >example.yaml <<EOF
foo:
  bar: "old"
EOF
$ lensed set /foo/bar=new <example.yaml
foo:
  bar: "new"
```

### Lenses (file formats)

Once you reached a string field, you can "dig deeper" and interpret the content of that string
as another hierarchical text format, and address other fields inside that inner string:

```console
$ cat >example.yaml <<EOF                  
foo: |
  [bar]
  baz = "old"
EOF
$ lensed set "/foo/~(toml)/bar/baz=new" <example.yaml  
foo: |
  [bar]
  baz = "new"
```

The edits in the inner text are properly escaped back.

```console
$ cat >example.yaml <<EOF                      
foo: "bar = \"I'm old\""
EOF
$ lensed set "/foo/~(toml)/bar=I'm \"quoted\"" -f example.yaml  
$ cat example.yaml
foo: "bar = \"I'm \\\"quoted\\\"\""
$ lensed get "/foo/~(toml)/bar" < example.yaml
I'm "quoted"
```

Supported lenses:

* yaml (also json)
* yamls (yaml stream)
* toml
* base64
* line
* regexp
* oci (alias ociImageRef)
* jsonnet

#### Base64

Decoding a base64 field from a YAML file is not hard:

```console
$ cat >example.yaml <<EOF
apiVersion: v1
kind: Secret
data:
  foo: eyJwYXNzd29yZCI6ICAgICAiaHVudGVyMTIifQ==
EOF
$ lensed get "/data/foo/~(base64)" <example.yaml                               
{"password":     "hunter12"}
```

But editing a field inside of a base64 encoded JSON inside a YAML file, it's a breeze with `lensed`:

```console
$ lensed set "/data/foo/~(base64)/~(yaml)/password=newpassword" -f example.yaml
$ cat example.yaml
apiVersion: v1
kind: Secret
data:
  foo: eyJwYXNzd29yZCI6ICAgICAibmV3cGFzc3dvcmQifQ==
$ lensed get "/data/foo/~(base64)" <example.yaml                               
{"password":     "newpassword"}
```

(notice how the original formatting of the embedded value is preserved, despite the various decoding/encoding roundtrips)

#### Jsonnet

This example also shows that you can edit a top-level format that is not YAML.
All you need to do is to start the JSONPointer with `~(<lensname>)`:

```console
$ cat >example.yaml <<EOF
{
  foo+: {
    bar: "old",
  }
}
EOF
$ lensed get "~(jsonnet)/foo/bar" <example.yaml 
old
$ lensed set "~(jsonnet)/foo/bar=new" <example.yaml
{
  foo+: {
    bar: "new",
  }
}
```

(support for editing `import` and `importstr` expressions would be very useful but it's not yet implemented)
