// Code generated by "next 0.0.4"; DO NOT EDIT.

package a


// XX constant
// XX value 2
const XX = 1 // XX value
// Constants
const ServerName = "Comprehensive Test Server"
const Version = "1.0.0"
const MaxConnections = 1000
const Pi = 3.14159265358979323846
const MaxInt64 = 9223372036854775807 // 2^63 - 1
const MinInt64 = -9223372036854775808 // -2^63
// Constants with complex expressions
const A = 1
const B = 3
const C = 9
const D = 7
const E = 28
const F = 1052
const G = 1052
const H = 5672
const I = 5673.618 // Approximation of golden ratio
const J = 47.28015 // 120 is 5!
// Constants with function calls
const StringLength = 13
const MinValue = 1
const MaxValue = 5673
// Constants using built-in functions
const IntFromBool = 1
const IntFromFloat = 3
const FloatFromInt = 42.0
const FloatFromBool = 0.0
const BoolFromInt = true
const BoolFromString = true
const FormattedString1 = "The answer is 42"
const FormattedString2 = "Pi is approximately 3.14"
const FormattedString3 = "Hello World\n"
// Constants for testing complex expressions and bitwise operations
const Complex1 = 5673
const Complex2 = 78547
const Complex3 = 31
const Complex4 = 31
const Complex5 = 31

// Enum with iota
type Color int32

const (
    ColorRed = 1
    ColorGreen = 2
    ColorBlue = 4
    ColorAlpha = 8
    ColorYellow = 3
    ColorCyan = 6
    ColorMagenta = 5
    ColorWhite = 7
)

// Enum with complex iota usage
type FilePermission int32

const (
    FilePermissionNone = 0
    FilePermissionExecute = 1
    FilePermissionWrite = 2
    FilePermissionRead = 4
    FilePermissionUserRead = 4
    FilePermissionUserWrite = 32
    FilePermissionUserExecute = 256
    FilePermissionGroupRead = 2048
    FilePermissionGroupWrite = 16384
    FilePermissionGroupExecute = 131072
    FilePermissionOthersRead = 1048576
    FilePermissionOthersWrite = 8388608
    FilePermissionOthersExecute = 67108864
    // 4|32|256|2048|16384|131072|1048576|8388608|67108864
    // 4 + 32 + 256 + 2048 + 16384 + 131072 + 1048576 + 8388608 + 67108864
    FilePermissionAll = 76695844
)

type Day int32

const (
    DayMonday = 1
    DayTuesday = 2
    DayWednesday = 4
    DayThursday = 8
    DayFriday = 16
    DaySaturday = 32
    DaySunday = 64
    DayWeekday = 31
    DayWeekend = 96
)

type Month int32

const (
    MonthJanuary = 1
    MonthFebruary = 2
    MonthMarch = 4
    MonthApril = 8
    MonthMay = 16
    MonthJune = 32
    MonthJuly = 64
    MonthAugust = 128
    MonthSeptember = 256
    MonthOctober = 512
    MonthNovember = 1024
    MonthDecember = 2048
    MonthQ1 = 7
    MonthQ2 = 56
    MonthQ3 = 448
    MonthQ4 = 3584
)

// Test cases for iota
type IotatestEnum int32

const (
    IotatestEnumA = 0 // 0
    IotatestEnumB = 1 // 1
    IotatestEnumC = 0 // 0
    IotatestEnumD = 2 // 2
    IotatestEnumE = 0 // 0
    IotatestEnumF = 1 // 1
    IotatestEnumG = 0 // 0
)

// Struct types
type Point2D struct {
    X float64
    Y float64
}

type Point3D struct {
    Point Point2D
    Z float64
}

type Rectangle struct {
    TopLeft Point2D
    BottomRight Point2D
}

// Struct with various field types
type ComplexStruct struct {
    Flag bool
    TinyInt int8
    SmallInt int16
    MediumInt int32
    BigInt int64
    DefaultInt int
    SinglePrecision float32
    DoublePrecision float64
    Text string
    SingleByte byte
    ByteArray []byte
    FixedArray [5]int
    DynamicArray []string
    IntArray []int
    Dictionary map[string]int
}

type User struct {
    Id int64
    Username string
    Email string
    PreferredDay Day
    BirthMonth Month
}

type UserProfile struct {
    User User
    FirstName string
    LastName string
    Age int
    Interests []string
}

// message types
type LoginRequest struct {
    Username string
    Password string
    DeviceId string
    TwoFactorToken string
}

func (LoginRequest) MessageType() int { return 201 }

type LoginResponse struct {
    Success bool
    ErrorMessage string
    AuthenticationToken string
    User User
}

func (LoginResponse) MessageType() int { return 202 }

type GenericRequest struct {
    RequestId string
    Timestamp int64
}

type GenericResponse struct {
    RequestId string
    Timestamp int64
    Success bool
    ErrorCode string
    ErrorMessage string
}
