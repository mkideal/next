// Code generated by "next 0.0.4"; DO NOT EDIT.

package com.example.demo;

import java.util.List;
import java.util.Map;
import java.util.Arrays;

/**
 * LoginResponse represents a login response message (type 102)
 */
public class LoginResponse {
    private String token;

    public String getToken() {
        return token;
    }

    public void setToken(String token) {
        this.token = token;
    }

    private User user;

    public User getUser() {
        return user;
    }

    public void setUser(User user) {
        this.user = user;
    }


    public static int MessageType() { return 102; }

    public LoginResponse() {}
}