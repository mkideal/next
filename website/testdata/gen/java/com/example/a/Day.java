// Code generated by "next 0.0.4"; DO NOT EDIT.

package com.example.a;

public enum Day {
    Monday((int) 1),
    Tuesday((int) 2),
    Wednesday((int) 4),
    Thursday((int) 8),
    Friday((int) 16),
    Saturday((int) 32),
    Sunday((int) 64),
    Weekday((int) 31),
    Weekend((int) 96);

    private final int value;

    Day(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }
}