-- Code generated by "next"; DO NOT EDIT

-- Package: demo

local _M_ = {}

_M_.Version = "1.0.0" -- String constant
_M_.MaxRetries = 3 -- Integer constant
_M_.Timeout = 3000.0 -- Float constant expression

-- Color represents different color options
-- Values: Red (1), Green (2), Blue (4), Yellow (8)
local Color = {
	RED = 1,
	GREEN = 2,
	BLUE = 4,
	YELLOW = 8
}
_M_.Color = Color

-- MathConstants represents mathematical constants
local MathConstants = {
	PI = 3.14159265358979323846,
	E = 2.71828182845904523536
}
_M_.MathConstants = MathConstants

-- OperatingSystem represents different operating systems
local OperatingSystem = {
	WINDOWS = "windows",
	LINUX = "linux",
	MAC_OS = "macos",
	ANDROID = "android",
	IOS = "ios"
}
_M_.OperatingSystem = OperatingSystem

-- User represents a user in the system
local User = {}
_M_.User = User
User.__index = User

function User:new()
	local obj = {
		id = 0,
		username = "",
		tags = {},
		scores = {},
		coordinates = {},
		matrix = {},
		email = "",
		favoriteColor = nil,
		-- @next(tokens) applies to the node name:
		-- - For snake_case: "last_login_ip"
		-- - For camelCase: "lastLoginIP"
		-- - For PascalCase: "LastLoginIP"
		-- - For kebab-case: "last-login-ip"
		lastLoginIP = "",
		extra = nil
	}
	setmetatable(obj, self)
	return obj
end

-- uint64 represents a 64-bit unsigned integer.
-- - In Go, it is aliased as uint64
-- - In C++, it is aliased as uint64_t
-- - In Java, it is aliased as long
-- - In Rust, it is aliased as u64
-- - In C#, it is aliased as ulong
-- - In Protobuf, it is represented as uint64
-- - In other languages, it is represented as a struct with low and high 32-bit integers.
local Uint64 = {}
_M_.Uint64 = Uint64
Uint64.__index = Uint64

function Uint64:new()
	local obj = {
		low = 0,
		high = 0
	}
	setmetatable(obj, self)
	return obj
end

-- uint128 represents a 128-bit unsigned integer.
-- - In rust, it is aliased as u128
-- - In other languages, it is represented as a struct with low and high 64-bit integers.
local Uint128 = {}
_M_.Uint128 = Uint128
Uint128.__index = Uint128

function Uint128:new()
	local obj = {
		low = Uint64:new(),
		high = Uint64:new()
	}
	setmetatable(obj, self)
	return obj
end

-- Contract represents a smart contract
local Contract = {}
_M_.Contract = Contract
Contract.__index = Contract

function Contract:new()
	local obj = {
		address = Uint128:new(),
		data = nil
	}
	setmetatable(obj, self)
	return obj
end

-- LoginRequest represents a login request message (type 101)
-- @message annotation is a custom annotation that generates message types.
local LoginRequest = {}
_M_.LoginRequest = LoginRequest
LoginRequest.__index = LoginRequest

function LoginRequest:new()
	local obj = {
		username = "",
		password = "",
		-- @next(optional) annotation is a builtin annotation that marks a field as optional.
		device = "",
		os = nil,
		timestamp = 0
	}
	setmetatable(obj, self)
	return obj
end

-- LoginResponse represents a login response message (type 102)
local LoginResponse = {}
_M_.LoginResponse = LoginResponse
LoginResponse.__index = LoginResponse

function LoginResponse:new()
	local obj = {
		token = "",
		user = User:new()
	}
	setmetatable(obj, self)
	return obj
end

-- Reader provides reading functionality
local Reader = {}
_M_.Reader = Reader
Reader.__index = Reader

-- @next(error) applies to the method:
-- - For Go: The method may return an error
-- - For C++/Java: The method throws an exception
--
-- @next(mut) applies to the method:
-- - For C++: The method is non-const
-- - For other languages: This annotation may not have a direct effect
--
-- @next(mut) applies to the parameter buffer:
-- - For C++: The parameter is non-const, allowing modification
-- - For other languages: This annotation may not have a direct effect,
--   but indicates that the buffer content may be modified
function Reader:read(buffer)
	error("Method not implemented")
end

-- HTTPClient provides HTTP request functionality
local HTTPClient = {}
_M_.HTTPClient = HTTPClient
HTTPClient.__index = HTTPClient

-- Available for all languages
function HTTPClient:request(url, method, body)
	error("Method not implemented")
end

function HTTPClient:request2(url, method, body)
	error("Method not implemented")
end

return _M_
