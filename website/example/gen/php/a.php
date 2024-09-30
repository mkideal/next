<?php
namespace a;

/**
 * XX constant
 * XX value 2
 */
const X_X = 1; // XX value
/**
 * Constants
 */
const SERVER_NAME = "Comprehensive Test Server";
const VERSION = "1.0.0";
const MAX_CONNECTIONS = 1000;
const PI = 3.14159265358979323846;
const MAX_INT_64 = 9223372036854775807; // 2^63 - 1
const MIN_INT_64 = -9223372036854775808; // -2^63
/**
 * Constants with complex expressions
 */
const A = 1;
const B = 3;
const C = 9;
const D = 7;
const E = 28;
const F = 1052;
const G = 1052;
const H = 5672;
const I = 5673.618; // Approximation of golden ratio
const J = 47.28015; // 120 is 5!
/**
 * Constants with function calls
 */
const STRING_LENGTH = 13;
const MIN_VALUE = 1;
const MAX_VALUE = 5673;
/**
 * Constants using built-in functions
 */
const INT_FROM_BOOL = 1;
const INT_FROM_FLOAT = 3;
const FLOAT_FROM_INT = 42.0;
const FLOAT_FROM_BOOL = 0;
const BOOL_FROM_INT = true;
const BOOL_FROM_STRING = true;
const FORMATTED_STRING_1 = "The answer is 42";
const FORMATTED_STRING_2 = "Pi is approximately 3.14";
const FORMATTED_STRING_3 = "Hello World\n";
/**
 * Constants for testing complex expressions and bitwise operations
 */
const COMPLEX_1 = 5673;
const COMPLEX_2 = 78547;
const COMPLEX_3 = 31;
const COMPLEX_4 = 31;
const COMPLEX_5 = 31;

/**
 * Enum with iota
 */
enum Color : int
{
    case RED = 1;
    case GREEN = 2;
    case BLUE = 4;
    case ALPHA = 8;
    case YELLOW = 3;
    case CYAN = 6;
    case MAGENTA = 5;
    case WHITE = 7;
}

/**
 * Enum with complex iota usage
 */
enum FilePermission : int
{
    case NONE = 0;
    case EXECUTE = 1;
    case WRITE = 2;
    case READ = 4;
    case USER_READ = 4;
    case USER_WRITE = 32;
    case USER_EXECUTE = 256;
    case GROUP_READ = 2048;
    case GROUP_WRITE = 16384;
    case GROUP_EXECUTE = 131072;
    case OTHERS_READ = 1048576;
    case OTHERS_WRITE = 8388608;
    case OTHERS_EXECUTE = 67108864;
    /**
     * 4|32|256|2048|16384|131072|1048576|8388608|67108864
     * 4 + 32 + 256 + 2048 + 16384 + 131072 + 1048576 + 8388608 + 67108864
     */
    case ALL = 76695844;
}

enum Day : int
{
    case MONDAY = 1;
    case TUESDAY = 2;
    case WEDNESDAY = 4;
    case THURSDAY = 8;
    case FRIDAY = 16;
    case SATURDAY = 32;
    case SUNDAY = 64;
    case WEEKDAY = 31;
    case WEEKEND = 96;
}

enum Month : int
{
    case JANUARY = 1;
    case FEBRUARY = 2;
    case MARCH = 4;
    case APRIL = 8;
    case MAY = 16;
    case JUNE = 32;
    case JULY = 64;
    case AUGUST = 128;
    case SEPTEMBER = 256;
    case OCTOBER = 512;
    case NOVEMBER = 1024;
    case DECEMBER = 2048;
    case Q_1 = 7;
    case Q_2 = 56;
    case Q_3 = 448;
    case Q_4 = 3584;
}

/**
 * Test cases for iota
 */
enum IotatestEnum : int
{
    case A = 0; // 0
    case B = 1; // 1
    case C = 0; // 0
    case D = 2; // 2
    case E = 0; // 0
    case F = 1; // 1
    case G = 0; // 0
}

/**
 * Struct types
 */
class Point2D
{
    public float $x;
    public float $y;

    public function __construct()
    {
        $this->x = 0;
        $this->y = 0;
    }
}

class Point3D
{
    public Point2D $point;
    public float $z;

    public function __construct()
    {
        $this->point = new Point2D();
        $this->z = 0;
    }
}

class Rectangle
{
    public Point2D $topLeft;
    public Point2D $bottomRight;

    public function __construct()
    {
        $this->topLeft = new Point2D();
        $this->bottomRight = new Point2D();
    }
}

/**
 * Struct with various field types
 */
class ComplexStruct
{
    public bool $flag;
    public int $tinyInt;
    public int $smallInt;
    public int $mediumInt;
    public int $bigInt;
    public int $defaultInt;
    public float $singlePrecision;
    public float $doublePrecision;
    public string $text;
    public int $singleByte;
    public string $byteArray;
    public array $fixedArray;
    public array $dynamicArray;
    public array $intArray;
    public array $dictionary;

    public function __construct()
    {
        $this->flag = false;
        $this->tinyInt = 0;
        $this->smallInt = 0;
        $this->mediumInt = 0;
        $this->bigInt = 0;
        $this->defaultInt = 0;
        $this->singlePrecision = 0;
        $this->doublePrecision = 0;
        $this->text = "";
        $this->singleByte = 0;
        $this->byteArray = "";
        $this->fixedArray = [];
        $this->dynamicArray = [];
        $this->intArray = [];
        $this->dictionary = [];
    }
}

class User
{
    public int $id;
    public string $username;
    public string $email;
    public Day $preferredDay;
    public Month $birthMonth;

    public function __construct()
    {
        $this->id = 0;
        $this->username = "";
        $this->email = "";
        $this->preferredDay = Day::MONDAY;
        $this->birthMonth = Month::JANUARY;
    }
}

class UserProfile
{
    public User $user;
    public string $firstName;
    public string $lastName;
    public int $age;
    public array $interests;

    public function __construct()
    {
        $this->user = new User();
        $this->firstName = "";
        $this->lastName = "";
        $this->age = 0;
        $this->interests = [];
    }
}

/**
 * message types
 */
class LoginRequest
{
    public string $username;
    public string $password;
    public string $deviceId;
    public string $twoFactorToken;

    public function __construct()
    {
        $this->username = "";
        $this->password = "";
        $this->deviceId = "";
        $this->twoFactorToken = "";
    }
}

class LoginResponse
{
    public bool $success;
    public string $errorMessage;
    public string $authenticationToken;
    public User $user;

    public function __construct()
    {
        $this->success = false;
        $this->errorMessage = "";
        $this->authenticationToken = "";
        $this->user = new User();
    }
}

class GenericRequest
{
    public string $requestId;
    public int $timestamp;

    public function __construct()
    {
        $this->requestId = "";
        $this->timestamp = 0;
    }
}

class GenericResponse
{
    public string $requestId;
    public int $timestamp;
    public bool $success;
    public string $errorCode;
    public string $errorMessage;

    public function __construct()
    {
        $this->requestId = "";
        $this->timestamp = 0;
        $this->success = false;
        $this->errorCode = "";
        $this->errorMessage = "";
    }
}
