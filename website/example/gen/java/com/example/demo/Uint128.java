// Code generated by "next"; DO NOT EDIT

package com.example.demo;

import java.util.List;
import java.util.Map;
import java.util.Arrays;

/**
 * uint128 represents a 128-bit unsigned integer.
 * - In rust, it is aliased as u128
 * - In other languages, it is represented as a struct with low and high 64-bit integers.
 */
public class Uint128 {
	private long low;

	public long getLow() {
		return low;
	}

	public void setLow(long low) {
		this.low = low;
	}

	private long high;

	public long getHigh() {
		return high;
	}

	public void setHigh(long high) {
		this.high = high;
	}

	public Uint128() {
	}
}
