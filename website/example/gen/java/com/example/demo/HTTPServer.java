// Code generated by "next 0.0.4"; DO NOT EDIT.

package com.example.demo;

import java.util.List;
import java.util.Map;
import java.util.Arrays;

public interface HTTPServer {
	/**
	 * @next(error) indicates that the method may return an error:
	 * - For Go: The method returns (LoginResponse, error)
	 * - For C++/Java: The method throws an exception
	 */
	void handle(String path, java.util.function.Function<com.sun.net.httpserver.HttpExchange, String> handler) throws Exception;
}
