// Code generated by "next 0.0.4"; DO NOT EDIT.

package com.example.a;

/**
 * Enum with iota
 */
public enum Color {
	Red((int) 1),
	Green((int) 2),
	Blue((int) 4),
	Alpha((int) 8),
	Yellow((int) 3),
	Cyan((int) 6),
	Magenta((int) 5),
	White((int) 7);

	private final int value;

	Color(int value) {
		this.value = value;
	}

	public int getValue() {
		return value;
	}
}
