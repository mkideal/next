# Code generated by "next 0.0.4"; DO NOT EDIT.

from typing import List, Dict, Any
from abc import ABC, abstractmethod

VERSION = "1.0.0" # String constant
MAX_RETRIES = 3 # Integer constant
TIMEOUT = 3000.0 # Float constant expression

"""
Color represents different color options
Values: Red (1), Green (2), Blue (4), Yellow (8)
"""
class Color:
	RED = 1
	GREEN = 2
	BLUE = 4
	YELLOW = 8

"""
MathConstants represents mathematical constants
"""
class MathConstants:
	PI = 3.14159265358979323846
	E = 2.71828182845904523536

"""
OperatingSystem represents different operating systems
"""
class OperatingSystem:
	WINDOWS = "windows"
	LINUX = "linux"
	MAC_O_S = "macos"
	ANDROID = "android"
	I_O_S = "ios"

"""
User represents a user in the system
"""
class User:
	def __init__(self):
		self.id = 0
		self.username = ""
		self.tags = []
		self.scores = {}
		self.coordinates = [0 for _ in range(3)]
		self.matrix = [[0 for _ in range(2)] for _ in range(3)]
		self.email = ""
		self.favorite_color = Color(0)
		self.extra = None

"""
uint64 represents a 64-bit unsigned integer.
- In Go, it is aliased as uint64
- In C++, it is aliased as uint64_t
- In Java, it is aliased as long
- In Rust, it is aliased as u64
- In C#, it is aliased as ulong
- In Protobuf, it is represented as uint64
- In other languages, it is represented as a struct with low and high 32-bit integers.
"""
class Uint64:
	def __init__(self):
		self.low = 0
		self.high = 0

"""
uint128 represents a 128-bit unsigned integer.
- In rust, it is aliased as u128
- In other languages, it is represented as a struct with low and high 64-bit integers.
"""
class Uint128:
	def __init__(self):
		self.low = Uint64()
		self.high = Uint64()

"""
Contract represents a smart contract
"""
class Contract:
	def __init__(self):
		self.address = Uint128()
		self.data = None

"""
LoginRequest represents a login request message (type 101)
@message annotation is a custom annotation that generates message types.
"""
class LoginRequest:
	def __init__(self):
		self.username = ""
		self.password = ""
		"""
		@optional annotation is a builtin annotation that marks a field as optional.
		"""
		self.device = ""
		self.os = OperatingSystem("")
		self.timestamp = 0

"""
LoginResponse represents a login response message (type 102)
"""
class LoginResponse:
	def __init__(self):
		self.token = ""
		self.user = User()

"""
Reader provides reading functionality
"""
class Reader(ABC):
	"""
	@next(error) applies to the method:
	- For Go: The method may return an error
	- For C++/Java: The method throws an exception

	@next(mut) applies to the method:
	- For C++: The method is non-const
	- For other languages: This annotation may not have a direct effect

	@next(mut) applies to the parameter buffer:
	- For C++: The parameter is non-const, allowing modification
	- For other languages: This annotation may not have a direct effect,
	  but indicates that the buffer content may be modified
	Args:
		buffer (bytes)
	Returns:
		int
	"""
	@abstractmethod
	def read(self, buffer: bytes) -> int:
		pass

"""
HTTPClient provides HTTP request functionality
"""
class HTTPClient(ABC):
	"""
	Available for all languages
	Args:
		url (str)
		method (str)
		body (str)
	Returns:
		str
	"""
	@abstractmethod
	def request(self, url: str, method: str, body: str) -> str:
		pass

	"""

	Args:
		url (str)
		method (str)
		body (str)
	Returns:
		str
	"""
	@abstractmethod
	def request_2(self, url: str, method: str, body: str) -> str:
		pass