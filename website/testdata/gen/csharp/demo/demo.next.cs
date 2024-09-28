// Code generated by "next v0.0.4"; DO NOT EDIT.

using System;
using System.Collections.Generic;

namespace demo
{
    public const string Version = "1.0.0"; // String constant
    public const int MaxRetries = 3; // Integer constant
    public const float Timeout = 3000.0; // Float constant expression

    // Color represents different color options
    // Values: Red (1), Green (2), Blue (4), Yellow (8)
    public enum Color
    {
        Red = 1,
        Green = 2,
        Blue = 4,
        Yellow = 8,
    }

    // MathConstants represents mathematical constants
    public enum MathConstants
    {
        Pi = 3.14159265358979323846,
        E = 2.71828182845904523536,
    }

    // OperatingSystem represents different operating systems
    public enum OperatingSystem
    {
        Windows = "windows",
        Linux = "linux",
        MacOS = "macos",
        Android = "android",
        IOS = "ios",
    }

    // User represents a user in the system
    public class User
    {
        public long id { get; set; }
        public string username { get; set; }
        public List<string> tags { get; set; }
        public Dictionary<string, int> scores { get; set; }
        public double[] coordinates { get; set; }
        public int[][] matrix { get; set; }
        public Color favoriteColor { get; set; }
        public string email { get; set; }
        public object extra { get; set; }
    }

    // uint128 represents a 128-bit unsigned integer.
    // - In rust, it is aliased as u128
    // - In other languages, it is represented as a struct with low and high 64-bit integers.
    public class Uint128
    {
        public ulong low { get; set; }
        public ulong high { get; set; }
    }

    // Contract represents a smart contract
    public class Contract
    {
        public Uint128 address { get; set; }
        public object data { get; set; }
    }

    // LoginRequest represents a login request message (type 101)
    // @message annotation is a custom annotation that generates message types.
    public class LoginRequest
    {
        public string username { get; set; }
        public string password { get; set; }
        // @optional annotation is a custom annotation that marks a field as optional.
        public string device { get; set; }
        public OperatingSystem os { get; set; }
        public long timestamp { get; set; }

        public static int MessageType() { return 101; }
    }

    // LoginResponse represents a login response message (type 102)
    public class LoginResponse
    {
        public string token { get; set; }
        public User user { get; set; }

        public static int MessageType() { return 102; }
    }

    // Reader provides reading functionality
    public interface Reader
    {
        // @next(error) applies to the method:
        // - For Go: The method may return an error
        // - For C++/Java: The method throws an exception
        // 
        // @next(mut) applies to the method:
        // - For C++: The method is non-const
        // - For other languages: This annotation may not have a direct effect
        // 
        // @next(mut) applies to the parameter buffer:
        // - For C++: The parameter is non-const, allowing modification
        // - For other languages: This annotation may not have a direct effect,
        //   but indicates that the buffer content may be modified
        int read(ref byte[] buffer);
    }

    // HTTPClient provides HTTP request functionality
    public interface HTTPClient
    {
        // Available for all languages
        string request(string url, string method, string body);
        string request2(string url, string method, string body);
    }
    
}