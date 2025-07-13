# Boardshapes Serialization Specification

This is the specification for the formats that Boardshapes data can be serialized to and serialized from.

## Binary

The binary format is the most compact format for Boardshapes data. This means it should take the least amount of storage space and the least time to process. However, this also means that the data is unreadable to a human without a special tool.

The data in the binary format is organized into "chunks." Each chunk is prefixed by a single byte identifying which type of chunk it is. The remainder of this specification on the binary format will show the different types of chunks, what they do, and how their data is structured.

The header of each section below is in the format [`chunk_number`] `chunk_name`. The `chunk_number` indicates what the value of the chunk's prefixing byte should be, in order to identify what type that chunk is.

---

### [0] Boardshapes Version

The version of the Boardshapes package used to generate this data.

This chunk should not appear more than once. It is very inadvisable to not have this chunk, as it may make the data impossible to deserialize (with the latest package version) in the future.

#### Structure

Should always be in [semver](https://semver.org/) format as a null-terminated UTF-8 string.

Examples:

- 0.5.2
- 1.0.3-beta

---

### [1] Color Table

Lists all possible colors that shapes may be identified by, and their names.

This chunk should not appear more than once.

#### Structure

The value of the first byte of the chunk should be the number of colors in this color table. Proceeding that, each color in the table should be represented by a 32-bit RGBA color followed by the name of the color a null-terminated UTF-8 string.

Example Color Table chunk:

```
0x04FF0000FF5265540000FF00FF477265656E000000FFFF426C756500000000FF426C61636B00
```

![Color Table Chunk Diagram](./spec_img/color_table.png)

### [8] Shape Path

WIP

### [9] Shape Color

WIP

## JSON

WIP
