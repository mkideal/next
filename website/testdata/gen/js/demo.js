// Code generated by "next v0.0.4"; DO NOT EDIT.

export const Version = "1.0.0"; // String constant
export const MaxRetries = 3; // Integer constant
export const Timeout = 3000.0; // Float constant expression

/**
 * Color represents different color options
 * Values: Red (1), Green (2), Blue (4), Yellow (8)
 */
export const Color = Object.freeze({
    Red: 1,
    Green: 2,
    Blue: 4,
    Yellow: 8
});

/**
 * MathConstants represents mathematical constants
 */
export const MathConstants = Object.freeze({
    Pi: 3.14159265358979323846,
    E: 2.71828182845904523536
});

/**
 * OperatingSystem represents different operating systems
 */
export const OperatingSystem = Object.freeze({
    Windows: "windows",
    Linux: "linux",
    MacOS: "macos",
    Android: "android",
    IOS: "ios"
});

/**
 * User represents a user in the system
 */
export class User {
    constructor() {
        /** @type { BigInt } */
        this.id = 0;
        /** @type { String } */
        this.username = "";
        /** @type { 'Array<String>' } */
        this.tags = [];
        /** @type { 'Map<String, Number>' } */
        this.scores = new Map();
        /** @type { 'Array<Number>' } */
        this.coordinates = [];
        /** @type { 'Array<'Array<Number>'>' } */
        this.matrix = [];
        /** @type { 'number' } */
        this.favoriteColor = Color[Object.keys(Color)[0]];
        /** @type { String } */
        this.email = "";
        /** @type { 'Object' } */
        this.extra = null;
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
export class Uint64 {
    constructor() {
        /** @type { Number } */
        this.low = 0;
        /** @type { Number } */
        this.high = 0;
    }
}

/**
 * uint128 represents a 128-bit unsigned integer.
 * - In rust, it is aliased as u128
 * - In other languages, it is represented as a struct with low and high 64-bit integers.
 */
export class Uint128 {
    constructor() {
        /** @type { Uint64 } */
        this.low = new Uint64;
        /** @type { Uint64 } */
        this.high = new Uint64;
    }
}

/**
 * Contract represents a smart contract
 */
export class Contract {
    constructor() {
        /** @type { Uint128 } */
        this.address = new Uint128;
        /** @type { 'Object' } */
        this.data = null;
    }
}

/**
 * LoginRequest represents a login request message (type 101)
 * @message annotation is a custom annotation that generates message types.
 */
export class LoginRequest {
    constructor() {
        /** @type { String } */
        this.username = "";
        /** @type { String } */
        this.password = "";
        /**
         * @optional annotation is a custom annotation that marks a field as optional.
         * @type { String }
         */
        this.device = "";
        /** @type { 'string' } */
        this.os = OperatingSystem[Object.keys(OperatingSystem)[0]];
        /** @type { BigInt } */
        this.timestamp = 0;
    }
}

/**
 * LoginResponse represents a login response message (type 102)
 */
export class LoginResponse {
    constructor() {
        /** @type { String } */
        this.token = "";
        /** @type { User } */
        this.user = new User;
    }
}

/**
 * Reader provides reading functionality
 */
export class Reader {
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
     * @param { 'Uint8Array' } buffer
     * @returns { Number }
     */
    read(buffer) {
        throw new Error('Method not implemented.');
    }
}

/**
 * HTTPClient provides HTTP request functionality
 */
export class HTTPClient {
    /**
     * Available for all languages
     *
     * @param { String } url * @param { String } method * @param { String } body
     * @returns { String }
     */
    request(url, method, body) {
        throw new Error('Method not implemented.');
    }
    /**
     * @param { String } url * @param { String } method * @param { String } body
     * @returns { String }
     */
    request2(url, method, body) {
        throw new Error('Method not implemented.');
    }
}
