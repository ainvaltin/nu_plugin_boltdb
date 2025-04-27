# Supported actions

- buckets - list buckets (output is stream);
- keys - list keys in a bucket (output is stream);
- get - get value of a key;
- set - set value of a key (either adds or overrides, value is given either as command input or argument);
- add - create bucket, will create all the buckets that do not exist in the given path ("bucket" flag);
- delete - deletes either bucket (flag "key" is not given) or key inside given bucket;
- stat - performance stat of the database (flag "bucket" not given) or given bucket;
- info - structure of the bucket;

# Flags "bucket" & "key"

The bucket and key names in Bolt are byte arrays so Nu Binary values is used to represent them, ie

    boltdb /db/file.name get -b [0x[010203], 0x[040506]] -k 0x[070809]

The command above returns value of the three byte key (byte values 7, 8, 9) in the bucket {4-5-6} which is nested inside bucket {1-2-3}.

In addition to raw byte values strings can be used, ie `-k "foobar"` is equal to `-k 0x[666F6F626172]`, IOW utf8 representation of the string is used.

Strings and Binary can be mixed, ie `-b [[bucket, 0x[0001]]]` is the same as `-b 0x[6275636b65740001]`. Note how nested list is used to concat the items into single array before it is used as item in the "bucket path" (without the outer List it would be path with two buckets).

The values returned by the 'buckets' and 'keys' actions are formatted (by Nu) by default as List of integers (ie `[102, 111, 111]`), use `boltdb ... | each { encode hex }` to format as hex strings, `boltdb ... | each { decode utf8 }` as text etc.