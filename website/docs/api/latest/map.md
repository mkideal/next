# Map {#user-content-Map}

`Map` represents a language map file. Usually, it is a text file that contains key-value pairs separated by a equal sign. The key is the name of the language feature. The built-in keys are:

- ext: the file extension for the language
- comment: the line comment pattern for the language
- int: the integer type for the language
- int8: the 8-bit integer type for the language
- int16: the 16-bit integer type for the language
- int32: the 32-bit integer type for the language
- int64: the 64-bit integer type for the language
- float32: the 32-bit floating-point type for the language
- float64: the 64-bit floating-point type for the language
- bool: the boolean type for the language
- string: the string type for the language
- byte: the byte type for the language
- bytes: the byte slice type for the language
- any: the any type for the language
- array: the array type for the language
- vector: the slice type for the language
- map: the map type for the language
- box: the box replacement for the language (for java, e.g., box(int) = Integer)

Here is an example of a language map file for the Java language:

```plain title="java.map"
ext=.java
comment=// %T%

# primitive types
int=int
int8=byte
int16=short
int32=int
int64=long
float32=float
float64=double
bool=boolean
string=String
byte=byte
bytes=byte[]
any=Object
map=Map<box(%K%),box(%V%)>
vector=List<box(%T%)>
array=%T%[]

# box types
box(int)=Integer
box(byte)=Byte
box(short)=Short
box(long)=Long
box(float)=Float
box(double)=Double
box(boolean)=Boolean
```

:::note

- The `%T%`, `%K%`, `%V%` and `%N%` are placeholders for the type, key type, value type and name type.
- Line comments are started with the `#` character and it must be the first character of the line.

```plain title="java.map"
# This will error
ext=.java # This is an invalid comment
# Comment must be the first character of the line (leading spaces are allowed)

	# This is a valid comment
```

See [Builtin Map Files](https://github.com/gopherd/next/tree/main/builtin) for more information.

:::

