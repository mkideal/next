// Code generated by "next 0.0.4"; DO NOT EDIT.

package com.example.demo;

/**
 * MathConstants represents mathematical constants
 */
public enum MathConstants {
    Pi((double) 3.14159265358979323846),
    E((double) 2.71828182845904523536);

    private final double value;

    MathConstants(double value) {
        this.value = value;
    }

    public double getValue() {
        return value;
    }
}