# Code generated by "next 0.0.4"; DO NOT EDIT.

from typing import List, Dict, Any

"""
XX constant
XX value 2
"""
X_X = 1 # XX value
"""
Constants
"""
SERVER_NAME = "Comprehensive Test Server"
VERSION = "1.0.0"
MAX_CONNECTIONS = 1000
PI = 3.14159265358979323846
MAX_INT_64 = 9223372036854775807 # 2^63 - 1
MIN_INT_64 = -9223372036854775808 # -2^63
"""
Constants with complex expressions
"""
A = 1
B = 3
C = 9
D = 7
E = 28
F = 1052
G = 1052
H = 5672
I = 5673.618 # Approximation of golden ratio
J = 47.28015 # 120 is 5!
"""
Constants with function calls
"""
STRING_LENGTH = 13
MIN_VALUE = 1
MAX_VALUE = 5673
"""
Constants using built-in functions
"""
INT_FROM_BOOL = 1
INT_FROM_FLOAT = 3
FLOAT_FROM_INT = 42.0
FLOAT_FROM_BOOL = 0.0
BOOL_FROM_INT = True
BOOL_FROM_STRING = True
FORMATTED_STRING_1 = "The answer is 42"
FORMATTED_STRING_2 = "Pi is approximately 3.14"
FORMATTED_STRING_3 = "Hello World\n"
"""
Constants for testing complex expressions and bitwise operations
"""
COMPLEX_1 = 5673
COMPLEX_2 = 78547
COMPLEX_3 = 31
COMPLEX_4 = 31
COMPLEX_5 = 31

"""
Enum with iota
"""
class Color:
    RED = 1
    GREEN = 2
    BLUE = 4
    ALPHA = 8
    YELLOW = 3
    CYAN = 6
    MAGENTA = 5
    WHITE = 7

"""
Enum with complex iota usage
"""
class FilePermission:
    NONE = 0
    EXECUTE = 1
    WRITE = 2
    READ = 4
    USER_READ = 4
    USER_WRITE = 32
    USER_EXECUTE = 256
    GROUP_READ = 2048
    GROUP_WRITE = 16384
    GROUP_EXECUTE = 131072
    OTHERS_READ = 1048576
    OTHERS_WRITE = 8388608
    OTHERS_EXECUTE = 67108864
    """
    4|32|256|2048|16384|131072|1048576|8388608|67108864
    4 + 32 + 256 + 2048 + 16384 + 131072 + 1048576 + 8388608 + 67108864
    """
    ALL = 76695844

class Day:
    MONDAY = 1
    TUESDAY = 2
    WEDNESDAY = 4
    THURSDAY = 8
    FRIDAY = 16
    SATURDAY = 32
    SUNDAY = 64
    WEEKDAY = 31
    WEEKEND = 96

class Month:
    JANUARY = 1
    FEBRUARY = 2
    MARCH = 4
    APRIL = 8
    MAY = 16
    JUNE = 32
    JULY = 64
    AUGUST = 128
    SEPTEMBER = 256
    OCTOBER = 512
    NOVEMBER = 1024
    DECEMBER = 2048
    Q_1 = 7
    Q_2 = 56
    Q_3 = 448
    Q_4 = 3584

"""
Test cases for iota
"""
class IotatestEnum:
    A = 0 # 0
    B = 1 # 1
    C = 0 # 0
    D = 2 # 2
    E = 0 # 0
    F = 1 # 1
    G = 0 # 0

"""
Struct types
"""
class Point2D:
    def __init__(self):
        self.x = 0
        self.y = 0

class Point3D:
    def __init__(self):
        self.point = Point2D()
        self.z = 0

class Rectangle:
    def __init__(self):
        self.top_left = Point2D()
        self.bottom_right = Point2D()

"""
Struct with various field types
"""
class ComplexStruct:
    def __init__(self):
        self.flag = False
        self.tiny_int = 0
        self.small_int = 0
        self.medium_int = 0
        self.big_int = 0
        self.default_int = 0
        self.single_precision = 0
        self.double_precision = 0
        self.text = ""
        self.single_byte = 0
        self.byte_array = b""
        self.fixed_array = [0 for _ in range(5)]
        self.dynamic_array = []
        self.int_array = []
        self.dictionary = {}

class User:
    def __init__(self):
        self.id = 0
        self.username = ""
        self.email = ""
        self.preferred_day = Day(0)
        self.birth_month = Month(0)

class UserProfile:
    def __init__(self):
        self.user = User()
        self.first_name = ""
        self.last_name = ""
        self.age = 0
        self.interests = []

"""
message types
"""
class LoginRequest:
    def __init__(self):
        self.username = ""
        self.password = ""
        self.device_id = ""
        self.two_factor_token = ""

class LoginResponse:
    def __init__(self):
        self.success = False
        self.error_message = ""
        self.authentication_token = ""
        self.user = User()

class GenericRequest:
    def __init__(self):
        self.request_id = ""
        self.timestamp = 0

class GenericResponse:
    def __init__(self):
        self.request_id = ""
        self.timestamp = 0
        self.success = False
        self.error_code = ""
        self.error_message = ""
