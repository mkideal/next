// Code generated by "next 0.0.4"; DO NOT EDIT.

use std::vec::Vec;
use std::boxed::Box;
use std::collections::HashMap;
use std::any::Any;

pub const A: i32 = 1;
pub const B: &'static str = "hello";
pub const C: f64 = 3.14;
pub const D: bool = true;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Color {
	Red = 0,
	Green = 1,
	Blue = 2,
}

impl Color {
	pub fn value(&self) -> i32 {
		match self {
			Color::Red => 0,
			Color::Green => 1,
			Color::Blue => 2,
		}
	}
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum LoginType {
	Username = 1,
	Email = 2,
}

impl LoginType {
	pub fn value(&self) -> i32 {
		match self {
			LoginType::Username => 1,
			LoginType::Email => 2,
		}
	}
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum UserType {
	Admin = 1,
	User = 2,
}

impl UserType {
	pub fn value(&self) -> i32 {
		match self {
			UserType::Admin => 1,
			UserType::User => 2,
		}
	}
}

pub struct User {
	pub type: UserType,
	pub id: i32,
	pub username: String,
	pub password: String,
	pub device_id: String,
	pub two_factor_token: String,
	pub roles: Vec<String>,
	pub metadata: HashMap<String, String>,
	pub scores: [i32; 4],
}

pub struct LoginRequest {
	pub type: LoginType,
	pub username: String,
	pub password: String,
	pub device_id: String,
	pub two_factor_token: String,
}

pub struct LoginResponse {
	pub success: bool,
	pub error_message: String,
	pub authentication_token: String,
	pub user: User,
}