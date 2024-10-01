// Code generated by "next 0.0.4"; DO NOT EDIT.

export const Version: string = "1.0.0"; // String constant
export const MaxRetries: number = 3; // Integer constant
export const Timeout: number = 3000.0; // Float constant expression

/**
 * Color represents different color options
 * Values: Red (1), Green (2), Blue (4), Yellow (8)
 */
export enum Color {
	Red = 1,
	Green = 2,
	Blue = 4,
	Yellow = 8
}

/**
 * MathConstants represents mathematical constants
 */
export enum MathConstants {
	Pi = 3.14159265358979323846,
	E = 2.71828182845904523536
}

/**
 * OperatingSystem represents different operating systems
 */
export enum OperatingSystem {
	Windows = "windows",
	Linux = "linux",
	MacOS = "macos",
	Android = "android",
	IOS = "ios"
}

/**
 * User represents a user in the system
 */
export class User {
	id: bigint = 0n;
	username: string = "";
	tags: Array<string> = [];
	scores: Map<string, number> = new Map();
	coordinates: Array<number> = [];
	matrix: Array<Array<number>> = [];
	email: string = "";
	favoriteColor: Color = 0 as Color;
	extra: any = null;
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
export class Uint64 {
	low: number = 0;
	high: number = 0;
}

/**
 * uint128 represents a 128-bit unsigned integer.
 * - In rust, it is aliased as u128
 * - In other languages, it is represented as a struct with low and high 64-bit integers.
 */
export class Uint128 {
	low: Uint64 = new Uint64;
	high: Uint64 = new Uint64;
}

/**
 * Contract represents a smart contract
 */
export class Contract {
	address: Uint128 = new Uint128;
	data: any = null;
}

/**
 * LoginRequest represents a login request message (type 101)
 * @message annotation is a custom annotation that generates message types.
 */
export class LoginRequest {
	username: string = "";
	password: string = "";
	/**
	 * @optional annotation is a builtin annotation that marks a field as optional.
	 */
	device: string = "";
	os: OperatingSystem = "" as OperatingSystem;
	timestamp: bigint = 0n;
}

/**
 * LoginResponse represents a login response message (type 102)
 */
export class LoginResponse {
	token: string = "";
	user: User = new User;
}

/**
 * Reader provides reading functionality
 */
export interface Reader {
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
	 *
	 * @param buffer { Uint8Array }
	 * @returns { number }
	 */
	read(buffer: Uint8Array): number;
}

/**
 * HTTPClient provides HTTP request functionality
 */
export interface HTTPClient {
	/**
	 * Available for all languages
	 *
	 * @param url { string }
	 * @param method { string }
	 * @param body { string }
	 * @returns { string }
	 */
	request(url: string, method: string, body: string): string;
	request2(url: string, method: string, body: string): string;
}
