// Code generated by "next 0.0.4"; DO NOT EDIT.

package com.example.demo;

import java.util.List;
import java.util.Map;
import java.util.Arrays;

/**
 * Reader provides reading functionality
 */
public interface Reader {
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
	 */
	int read(byte[] buffer) throws Exception;
}