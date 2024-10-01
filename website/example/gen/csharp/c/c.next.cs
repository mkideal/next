// Code generated by "next 0.0.4"; DO NOT EDIT.

using System;
using System.Collections.Generic;

namespace c
{
	public const int A = 1;
	public const string B = "hello";
	public const double C = 3.14;
	public const bool D = true;

	public enum Color
	{
		Red = 0,
		Green = 1,
		Blue = 2,
	}

	public enum LoginType
	{
		Username = 1,
		Email = 2,
	}

	public enum UserType
	{
		Admin = 1,
		User = 2,
	}

	public class User
	{
		public UserType type { get; set; }
		public int id { get; set; }
		public string username { get; set; }
		public string password { get; set; }
		public string deviceId { get; set; }
		public string twoFactorToken { get; set; }
		public List<string> roles { get; set; }
		public Dictionary<string, string> metadata { get; set; }
		public int[] scores { get; set; }
	}

	public class LoginRequest
	{
		public LoginType type { get; set; }
		public string username { get; set; }
		public string password { get; set; }
		public string deviceId { get; set; }
		public string twoFactorToken { get; set; }
	}

	public class LoginResponse
	{
		public bool success { get; set; }
		public string errorMessage { get; set; }
		public string authenticationToken { get; set; }
		public User user { get; set; }
	}
}