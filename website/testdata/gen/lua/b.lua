-- Code generated by "next v0.0.4"; DO NOT EDIT.

-- Package: b


local a_a = require("a.a")

local _M_ = {}

_M_.TestEnum = {
    A = 1,
    B = 5,
    C = 5,
    D = 10,
    E = 20,
    F = 1,
    G = 2
}

local TestStruct = {}
_M_.TestStruct = TestStruct
TestStruct.__index = TestStruct

function TestStruct:new()
    local obj = {
        point = a_a.Point2D:new()
    }
    setmetatable(obj, self)
    return obj
end
return _M_