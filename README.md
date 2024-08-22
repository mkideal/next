# Next 语言规范

## 1. 介绍

Next 是一种用于生成其他语言代码或各种文件的语言。本文档定义了 Next 语言并列举了其语法和语义。

## 2. 词法元素

### 2.1 注释

Next 支持两种形式的注释：

1. 行注释以 `//` 开始，直到行末。
2. 通用注释以 `/*` 开始，以 `*/` 结束。

### 2.2 标识符

标识符命名程序中的实体，如变量和类型。标识符是一个或多个字母和数字的序列。标识符必须以字母开头。

```
identifier = letter { letter | unicode_digit } .
```

### 2.3 关键字

以下关键字被保留，不能用作标识符：

```
package   import    const     enum      struct    protocol
```

### 2.4 操作符和标点符号

```
+    &    &&   =    !=   (    )
-    |    ||   <<   <=   [    ]
*    !    ==   >>   >=   {    }
/    ,    ^    ;    !    <    >
%    .
```

**注**: 中括号 `[]` 目前尚未使用，但保留以被后续可能扩展语法时使用。

## 3. 源代码表示

### 3.1 源文件

源文件以 UTF-8 编码。文件扩展名为 `.next`。

### 3.2 包声明

每个 Next 源文件都以包声明开始：

```
PackageClause = "package" PackageName ";" .
PackageName    = identifier .
```

示例：
```next
package demo;
```

### 3.3 导入声明

导入声明用于引入其他文件：

```
ImportDecl       = "import" ImportSpec ";" .
ImportSpec       = string_lit .
```

示例：

```next
import "./a.next";
```

引入文件后，可以使用该文件中定义的包内的常量、枚举、结构体、协议等。

## 4. 声明和作用域

### 4.1 常量

常量声明可以使用 `const` 关键字：

```
ConstDecl      = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
ConstSpec      = identifier "=" Expression ";" .
```

常量可以是数字、字符串、布尔值、或者任何常量表达式。

示例：

```next
const V0 = 1;
const V1 = 100.5;
const V2 = 1000_000; // 等同于 1000000
const V3 = V1 + V2; // 表达式并且引用其他常量

enum Errno {
    OK = 0,
}

const (
    A = 1;
    B = 2.0;
    C = false;
    D = "hello";
    E = Errno.OK;  // 枚举字段引用
)
```

### 4.2 类型

#### 4.2.1 枚举类型

枚举声明使用 `enum` 关键字：

```
EnumDecl     = "enum" ( EnumSpec | "(" { EnumSpec ";" } ")" ) .
EnumSpec     = identifier "{" { identifier [ "=" Expression ] "," } "}" .
```

枚举可以使用含 `iota` 的表达式进行推导。**注意只有枚举可以使用 iota 推导，常量定义 const 不可以。**

示例：

```next
enum Color {
    Red = 1,
    Green = 2,
    Blue = 3,
}

enum Errno {
    OK = iota,  // 0
    Internal,   // 1
    BadRequest, // 2

    UserNotFound = iota + 100, // 100
    ProviderNotFound,          // 101
}

enum (
    EnumA {
        A1 = 0,
        A2 = 1,
    }

    EnumB {
        B1 = 0,
        B2 = 1,
    }
)
```

枚举字段可以通过 `EnumName.FieldName` 的方式引用，可用于常量定义和常量表达式中。

#### 4.2.2 结构体类型

结构体声明使用 `struct` 关键字：

```
StructDecl     = "struct" ( StructSpec | "(" { StructSpec ";" } ")" ) .
StructSpec     = identifier "{" { FieldDecl ";" } "}" .
FieldDecl      = identifier Type .
```

示例：

```next
struct Location {
    string country;
    string city;
    int zipCode;
}

struct (
    StructA {
        int field1;
        bool field2;
    }

    StructB {
        StructA a;
        vector<StructA> as;
        string field1;
        int field2;
    }
)
```

#### 4.2.3 协议类型

协议声明使用 `protocol` 关键字，语法与结构体类似：

```
ProtocolDecl     = "protocol" ( ProtocolSpec | "(" { ProtocolSpec ";" } ")" ) .
ProtocolSpec     = identifier "{" { FieldDecl ";" } "}" .
```

协议和结构体的主要区别是，协议不能作为任何结构体和协议的字段，必须是顶级的（即协议不能出现在任何结构体和协议的字段中）。

示例：

```next
protocol User {
    int64 id;
    string nickname;
    string avatar;
    Location location;
    vector<string> fields;
    map<string, int> scores;
    array<int, 6> history;
}

protocol (
    ProtocolA {}
    ProtocolB {
        string id;
    }
)
```

以下是协议类型的错误使用示例：

```
protocol Login {
    User user; // ERROR: User 在前面已经定义为 protocol，所以这里不可以使用
    vector<User> users; // ERROR: 容器中也不可使用协议
}

struct Logout {
    User user; // ERROR: User 在前面已经定义为 protocol，所以这里不可以使用
    map<int, User> users; // ERROR: 容器中也不可使用协议
}
```

### 4.3 注解

注解可以添加到包、任意的声明、常量、枚举（及其任意字段）、结构体（及其任意字段）、协议（及其任意字段）。注解使用 `@` 符号开头：

```
Annotation     = "@" identifier [ "(" [ Parameters ] ")" ] ";" .
Parameters     = PositionalParams | NamedParams .
PositionalParams = Expression { "," Expression } .
NamedParams    = NamedParam { "," NamedParam } .
NamedParam     = identifier "=" Expression .
```

注解支持无参数、有参数（包括具名参数和匿名参数）的形式。

示例：

```next
@next(
    go_package = "github.com/username/repo/a",
    cpp_namespace = "repo::a", // 最后一个参数后面的逗号可有可无
)
package demo;

@type(100)
protocol LoginRequest {
    @required
    string token;
    string ip;
}

@json(omitempty)
struct User {
    @key
    int id;

    @json(name="nick_name")
    string nickname;

    @json(ignore)
    string password;
}
```

**@next 是 Next 内置注解，其使用遵循明确的参数定义，其作用被 Next 内部解释，不应该使用 @next 作为自定义的注解**

## 5. 类型

### 5.1 基本类型

- 布尔类型：`bool`
- 整数类型：`int`、`int8`、`int16`、`int32`、`int64`
- 浮点数类型：`float32`、`float64`
- 字符串类型：`string`、`byte`、`bytes`

**注意: 不支持无符号整数类型。**

### 5.2 复合类型

- 数组类型：`array<T, N>`，其中 T 是元素类型，N 是数组长度
- 向量类型：`vector<T>`，其中 T 是元素类型
- 映射类型：`map<K, V>`，其中 K 是键类型，V 是值类型

## 6. 属性和字段

结构体和协议中的字段声明遵循以下语法：

```
FieldDecl = identifier Type ";" .
```

字段可以有注解。

## 7. 表达式

在 Next 语言中，所有表达式都是常量表达式，在编译时求值。表达式用于计算值，遵循以下语法：

```
Expression     = UnaryExpr | Expression binary_op Expression | FunctionCall .
UnaryExpr      = PrimaryExpr | unary_op UnaryExpr .
PrimaryExpr    = Operand | PrimaryExpr Selector .
Operand        = Literal | identifier | EnumFieldRef | "(" Expression ")" .
Selector       = "." identifier .
EnumFieldRef   = identifier "." identifier .
FunctionCall   = identifier "(" [ ArgumentList ] ")" .
ArgumentList   = Expression { "," Expression } .

binary_op     = "+" | "-" | "*" | "/" | "%" |
                "&" | "|" | "^" | ">>" | "<<" | "&^" |
                "<" | ">" | "<=" | ">=" | "==" | "!=" | "&&" | "||" .
unary_op      = "+" | "-" | "!" | "^" .
```

表达式可以包括：

1. 字面量（数字、字符串、布尔值）
2. 标识符（常量名、枚举名等）
3. 枚举字段引用
4. 二元操作（算术、位运算、逻辑，比较操作等）
5. 一元操作（正负号、逻辑非、按位取反等）
6. 圆括号分组
7. 字段选择（仅用于枚举）
8. 函数调用（目前仅支持内置函数: len, min, max）

所有表达式都是在编译时求值的常量表达式。它们可以包含字面量、常量标识符、枚举字段引用，以及对这些的操作。

Next 语言的表达式不支持索引操作。

示例：

```next
42                  // 数字字面量
"hello"             // 字符串字面量
true                // 布尔字面量
X                   // 常量标识符
Color.Red           // 枚举字段引用
x + y               // 二元加法
-z                  // 一元负号
(x + y) * z         // 带括号的表达式
min(x, y, z)        // 函数调用
max(a, b)           // 函数调用
len("hello")        // 函数调用
```

在常量声明中使用的所有表达式都必须是有效的常量表达式，可以在编译时求值。

## 8. 内置变量和函数

| 函数或变量 | 用法说明 |
|----------|---------|
| **iota** | 用于枚举声明中声明内置递增变量 |
| **int**(`x: bool\|int\|float`) | 将 bool，int 或 float 转换成整数 |
| **float**(`x: bool\|int\|float`) | 将 bool，int 或 float 转换成浮点数 |
| **bool**(`x: any`) | 将 bool，int，float 或 string 转换成bool |
| **min**(`x: any`, `y: any...`) | 获取一个或多个值的最小值 |
| **max**(`x: any`, `y: any...`) | 获取一个或多个值的最大值 |
| **len**(`s: string`) | 计算字符串的长度 |
| **sprint**(`args: any...`) | 用于表达式中，返回格式化后字符串 |
| **sprintf**(`fmt: string`, `args: any...`) | 用于表达式中，返回格式化后字符串 |
| **sprintln**(`args: any...`) | 用于表达式中，返回格式化后字符串 |
| **print**(`args: any...`) | 用于调试输出信息，**不可**用于表达式中 |
| **printf**(`fmt: string`, `args: any...`) | 用于调试输出信息，**不可**用于表达式中 |
| **println**(`args: any...`) | 用于调试输出信息，**不可**用于表达式中 |
| **assert**(`cond: bool`, `args: any...`) | 用于断言，**不可**用于表达式中 |

## 9. 词法约定

- 包名使用小驼峰式命名（建议使用全小写）。
- 常量，枚举成员，类型名（结构体、枚举、协议）使用大驼峰式命名。