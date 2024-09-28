// Code generated by "next v0.0.4"; DO NOT EDIT.

package demo

import "strconv"
import "fmt"
import "net/http"

var _ = strconv.FormatInt
var _ = (*fmt.Stringer)(nil)
var _ = http.HandlerFunc
const Version = "1.0.0" // String constant
const MaxRetries = 3 // Integer constant
const Timeout = 3000.0 // Float constant expression

// Color represents different color options
// Values: Red (1), Green (2), Blue (4), Yellow (8)
type Color int8

const (
    ColorRed = 1
    ColorGreen = 2
    ColorBlue = 4
    ColorYellow = 8
)

func (x Color) String() string {
    switch x {
    case ColorRed:
        return "Red"
    case ColorGreen:
        return "Green"
    case ColorBlue:
        return "Blue"
    case ColorYellow:
        return "Yellow"
    }
    return "Color(" + strconv.FormatInt(int64(x), 10) + ")"
}

// MathConstants represents mathematical constants
type MathConstants float64

const (
    MathConstantsPi = 3.14159265358979323846
    MathConstantsE = 2.71828182845904523536
)

func (x MathConstants) String() string {
    switch x {
    case MathConstantsPi:
        return "Pi"
    case MathConstantsE:
        return "E"
    }
    return fmt.Sprintf("MathConstants(%v)", x)
}

// OperatingSystem represents different operating systems
type OperatingSystem string

const (
    OperatingSystemWindows = "windows"
    OperatingSystemLinux = "linux"
    OperatingSystemMacOS = "macos"
    OperatingSystemAndroid = "android"
    OperatingSystemIOS = "ios"
)

func (x OperatingSystem) String() string {
    return string(x)
}

// User represents a user in the system
type User struct {
    Id int64
    Username string
    Tags []string
    Scores map[string]int
    Coordinates [3]float64
    Matrix [3][2]int
    FavoriteColor Color
    Email string
    Extra any
}

// uint128 represents a 128-bit unsigned integer.
// - In rust, it is aliased as u128
// - In other languages, it is represented as a struct with low and high 64-bit integers.
type Uint128 struct {
    Low uint64
    High uint64
}

// Contract represents a smart contract
type Contract struct {
    Address Uint128
    Data any
}

// LoginRequest represents a login request message (type 101)
// @message annotation is a custom annotation that generates message types.
type LoginRequest struct {
    Username string
    Password string
    // @optional annotation is a custom annotation that marks a field as optional.
    Device string
    Os OperatingSystem
    Timestamp int64
}

func (LoginRequest) MessageType() int { return 101 }

// LoginResponse represents a login response message (type 102)
type LoginResponse struct {
    Token string
    User User
}

func (LoginResponse) MessageType() int { return 102 }

// Reader provides reading functionality
type Reader interface {
    // @next(error) applies to the method:
    // - For Go: The method may return an error
    // - For C++/Java: The method throws an exception
    // 
    // @next(mut) applies to the method:
    // - For C++: The method is non-const
    // - For other languages: This annotation may not have a direct effect
    // 
    // @next(mut) applies to the parameter buffer:
    // - For C++: The parameter is non-const, allowing modification
    // - For other languages: This annotation may not have a direct effect,
    //   but indicates that the buffer content may be modified
    read(buffer []byte) (int, error)
}

type HTTPServer interface {
    // @next(error) indicates that the method may return an error:
    // - For Go: The method returns (LoginResponse, error)
    // - For C++/Java: The method throws an exception
    handle(path string, handler http.Handler) error
}

// HTTPClient provides HTTP request functionality
type HTTPClient interface {
    // Available for all languages
    request(url string, method string, body string) string
    request2(url string, method string, body string) string
    // Available for Go and Java
    get(url string) (string, error)
}
