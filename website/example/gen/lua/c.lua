-- Code generated by "next"; DO NOT EDIT

-- Package: c

local a_a = require("a.a")
local b_b = require("b.b")

local _M_ = {}

_M_.A = 1
_M_.B = "hello"
_M_.C = 3.14
_M_.D = true

local Color = {
	RED = 0,
	GREEN = 1,
	BLUE = 2
}
_M_.Color = Color

local LoginType = {
	USERNAME = 1,
	EMAIL = 2
}
_M_.LoginType = LoginType

local UserType = {
	ADMIN = 1,
	USER = 2
}
_M_.UserType = UserType

local User = {}
_M_.User = User
User.__index = User

function User:new()
	local obj = {
		type = nil,
		id = 0,
		username = "",
		password = "",
		deviceId = "",
		twoFactorToken = "",
		roles = {},
		metadata = {},
		scores = {}
	}
	setmetatable(obj, self)
	return obj
end

local LoginRequest = {}
_M_.LoginRequest = LoginRequest
LoginRequest.__index = LoginRequest

function LoginRequest:new()
	local obj = {
		type = nil,
		username = "",
		password = "",
		deviceId = "",
		twoFactorToken = ""
	}
	setmetatable(obj, self)
	return obj
end

local LoginResponse = {}
_M_.LoginResponse = LoginResponse
LoginResponse.__index = LoginResponse

function LoginResponse:new()
	local obj = {
		success = false,
		errorMessage = "",
		authenticationToken = "",
		user = User:new()
	}
	setmetatable(obj, self)
	return obj
end

return _M_
