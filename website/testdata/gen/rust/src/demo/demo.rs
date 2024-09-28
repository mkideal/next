// Code generated by "next 0.0.4"; DO NOT EDIT.

use std::vec::Vec;
use std::boxed::Box;
use std::collections::HashMap;
use std::any::Any;
pub const VERSION: &'static str = "1.0.0"; // String constant
pub const MAX_RETRIES: i32 = 3; // Integer constant
pub const TIMEOUT: f32 = 3000.0; // Float constant expression

/// Color represents different color options
/// Values: Red (1), Green (2), Blue (4), Yellow (8)
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Color {
    Red = 1,
    Green = 2,
    Blue = 4,
    Yellow = 8,
}
impl Color {
    pub fn value(&self) -> i32 {
        match self {
            Color::Red => 1,
            Color::Green => 2,
            Color::Blue => 4,
            Color::Yellow => 8,
        }
    }
}

/// MathConstants represents mathematical constants
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum MathConstants {
    Pi,
    E,
}
impl MathConstants {
    pub fn value(&self) -> f64 {
        match self {
            MathConstants::Pi => 3.14159265358979323846,
            MathConstants::E => 2.71828182845904523536,
        }
    }
}

/// OperatingSystem represents different operating systems
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OperatingSystem {
    Windows,
    Linux,
    MacOS,
    Android,
    IOS,
}
impl OperatingSystem {
    pub fn value(&self) -> &'static str {
        match self {
            OperatingSystem::Windows => "windows",
            OperatingSystem::Linux => "linux",
            OperatingSystem::MacOS => "macos",
            OperatingSystem::Android => "android",
            OperatingSystem::IOS => "ios",
        }
    }
}

/// User represents a user in the system
pub struct User {
    pub id: i64,
    pub username: String,
    pub tags: Vec<String>,
    pub scores: HashMap<String, i32>,
    pub coordinates: [f64; 3],
    pub matrix: [[i32; 2]; 3],
    pub favorite_color: Color,
    pub email: String,
    pub extra: Box<dyn Any>,
}

/// Contract represents a smart contract
pub struct Contract {
    pub address: u128,
    pub data: Box<dyn Any>,
}

/// LoginRequest represents a login request message (type 101)
/// @message annotation is a custom annotation that generates message types.
pub struct LoginRequest {
    pub username: String,
    pub password: String,
    /// @optional annotation is a custom annotation that marks a field as optional.
    pub device: String,
    pub os: OperatingSystem,
    pub timestamp: i64,
}

impl LoginRequest {
    pub fn message_type() -> i32 {
        101
    }
}

/// LoginResponse represents a login response message (type 102)
pub struct LoginResponse {
    pub token: String,
    pub user: User,
}

impl LoginResponse {
    pub fn message_type() -> i32 {
        102
    }
}

/// Reader provides reading functionality
pub trait Reader {
    /// @next(error) applies to the method:
    /// - For Go: The method may return an error
    /// - For C++/Java: The method throws an exception
    /// 
    /// @next(mut) applies to the method:
    /// - For C++: The method is non-const
    /// - For other languages: This annotation may not have a direct effect
    /// 
    /// @next(mut) applies to the parameter buffer:
    /// - For C++: The parameter is non-const, allowing modification
    /// - For other languages: This annotation may not have a direct effect,
    ///   but indicates that the buffer content may be modified
    fn read(&mut self, buffer: Vec<u8>)-> i32;
}

/// HTTPClient provides HTTP request functionality
pub trait HTTPClient {
    /// Available for all languages
    fn request(&self, url: String, method: String, body: String)-> String;
    fn request_2(&self, url: String, method: String, body: String)-> String;
}
