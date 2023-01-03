# pbconv

[Protocol Buffers](https://developers.google.com/protocol-buffers) data-format conversion utility.

This supports:

- `pbconv from-proto` : Convert JSON or [Protocol Buffers Text-format](https://developers.google.com/protocol-buffers/docs/text-format-spec) into binary of Protocol Buffers wire format
- `pbconv to-proto` : Convert binary of Protocol Buffers wire format into JSON or Text-format

## Installation

```sh
$ go install github.com/thara/pbconv@latest
```

## Example

sample proto file

```protobuf
syntax = "proto3";

package dev.thara.book;

message Book {
    int64 isbn = 1;
    string title = 2;
    string author = 3;
    bool published = 4;
}
```

### Generate binary of Protocol Buffers wire format from JSON

```sh
$ cat book.json
{
  "isbn": 123,
  "title": "sample",
  "author": "Tom",
  "published": true
}
$ cat book.json | pbconv to-proto --from json --out book.bin Book book.proto
```

### Show contents of Protocol Buffers wire format by [Text-format](https://developers.google.com/protocol-buffers/docs/text-format-spec)

```sh
$ pbconv from-proto --in book.bin --to text Book book.proto
isbn: 123
title: "sample"
author: "Tom"
published: true
```

## Author

Tomochika Hara https://thara.dev

## License

MIT
