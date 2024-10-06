<?php
namespace demo;

const VERSION = "1.0.0"; // String constant
const MAX_RETRIES = 3; // Integer constant
const TIMEOUT = 3000.0; // Float constant expression

/**
 * Color represents different color options
 * Values: Red (1), Green (2), Blue (4), Yellow (8)
 */
enum Color : int
{
	case RED = 1;
	case GREEN = 2;
	case BLUE = 4;
	case YELLOW = 8;
}

/**
 * OperatingSystem represents different operating systems
 */
enum OperatingSystem : string
{
	case WINDOWS = "windows";
	case LINUX = "linux";
	case MAC_O_S = "macos";
	case ANDROID = "android";
	case I_O_S = "ios";
}

/**
 * User represents a user in the system
 */
class User
{
	public int $id;
	public string $username;
	public array $tags;
	public array $scores;
	public array $coordinates;
	public array $matrix;
	public string $email;
	public Color $favoriteColor;
	/**
	 * @next(tokens) applies to the node name:
	 * - For snake_case: "last_login_ip"
	 * - For camelCase: "lastLoginIP"
	 * - For PascalCase: "LastLoginIP"
	 * - For kebab-case: "last-login-ip"
	 */
	public string $lastLoginIP;
	public mixed $extra;

	public function __construct()
	{
		$this->id = 0;
		$this->username = "";
		$this->tags = [];
		$this->scores = [];
		$this->coordinates = [];
		$this->matrix = [];
		$this->email = "";
		$this->favoriteColor = Color::RED;
		$this->lastLoginIP = "";
		$this->extra = null;
	}
}

/**
 * uint64 represents a 64-bit unsigned integer.
 * - In Go, it is aliased as uint64
 * - In C++, it is aliased as uint64_t
 * - In Java, it is aliased as long
 * - In Rust, it is aliased as u64
 * - In C#, it is aliased as ulong
 * - In Protobuf, it is represented as uint64
 * - In other languages, it is represented as a struct with low and high 32-bit integers.
 */
class Uint64
{
	public int $low;
	public int $high;

	public function __construct()
	{
		$this->low = 0;
		$this->high = 0;
	}
}

/**
 * uint128 represents a 128-bit unsigned integer.
 * - In rust, it is aliased as u128
 * - In other languages, it is represented as a struct with low and high 64-bit integers.
 */
class Uint128
{
	public Uint64 $low;
	public Uint64 $high;

	public function __construct()
	{
		$this->low = new Uint64();
		$this->high = new Uint64();
	}
}

/**
 * Contract represents a smart contract
 */
class Contract
{
	public Uint128 $address;
	public mixed $data;

	public function __construct()
	{
		$this->address = new Uint128();
		$this->data = null;
	}
}

/**
 * LoginRequest represents a login request message (type 101)
 * @message annotation is a custom annotation that generates message types.
 */
class LoginRequest
{
	public string $username;
	public string $password;
	/**
	 * @next(optional) annotation is a builtin annotation that marks a field as optional.
	 */
	public string $device;
	public OperatingSystem $os;
	public int $timestamp;

	public function __construct()
	{
		$this->username = "";
		$this->password = "";
		$this->device = "";
		$this->os = OperatingSystem::WINDOWS;
		$this->timestamp = 0;
	}
}

/**
 * LoginResponse represents a login response message (type 102)
 */
class LoginResponse
{
	public string $token;
	public User $user;

	public function __construct()
	{
		$this->token = "";
		$this->user = new User();
	}
}

/**
 * Reader provides reading functionality
 */
interface Reader
{
	/**
	 * @next(error) applies to the method:
	 * - For Go: The method may return an error
	 * - For C++/Java: The method throws an exception
	 *
	 * @next(mut) applies to the method:
	 * - For C++: The method is non-const
	 * - For other languages: This annotation may not have a direct effect
	 *
	 * @next(mut) applies to the parameter buffer:
	 * - For C++: The parameter is non-const, allowing modification
	 * - For other languages: This annotation may not have a direct effect,
	 *   but indicates that the buffer content may be modified
	 */
	public function read(string $buffer): int;
}

/**
 * HTTPClient provides HTTP request functionality
 */
interface HTTPClient
{
	/**
	 * Available for all languages
	 */
	public function request(string $url, string $method, string $body): string;
	public function request2(string $url, string $method, string $body): string;
}
