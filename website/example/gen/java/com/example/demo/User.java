// Code generated by "next"; DO NOT EDIT

package com.example.demo;

import java.util.List;
import java.util.Map;
import java.util.Arrays;

/**
 * User represents a user in the system
 */
public class User {
	private long id;

	public long getId() {
		return id;
	}

	public void setId(long id) {
		this.id = id;
	}

	private String username;

	public String getUsername() {
		return username;
	}

	public void setUsername(String username) {
		this.username = username;
	}

	private List<String> tags;

	public List<String> getTags() {
		return tags;
	}

	public void setTags(List<String> tags) {
		this.tags = tags;
	}

	private Map<String, Integer> scores;

	public Map<String, Integer> getScores() {
		return scores;
	}

	public void setScores(Map<String, Integer> scores) {
		this.scores = scores;
	}

	private double[] coordinates;

	public double[] getCoordinates() {
		return coordinates;
	}

	public void setCoordinates(double[] coordinates) {
		this.coordinates = coordinates;
	}

	private int[][] matrix;

	public int[][] getMatrix() {
		return matrix;
	}

	public void setMatrix(int[][] matrix) {
		this.matrix = matrix;
	}

	private String email;

	public String getEmail() {
		return email;
	}

	public void setEmail(String email) {
		this.email = email;
	}

	private Color favoriteColor;

	public Color getFavoriteColor() {
		return favoriteColor;
	}

	public void setFavoriteColor(Color favoriteColor) {
		this.favoriteColor = favoriteColor;
	}

	/**
	 * @next(tokens) applies to the node name:
	 * - For snake_case: "last_login_ip"
	 * - For camelCase: "lastLoginIP"
	 * - For PascalCase: "LastLoginIP"
	 * - For kebab-case: "last-login-ip"
	 */
	private String lastLoginIP;

	public String getLastLoginIP() {
		return lastLoginIP;
	}

	public void setLastLoginIP(String lastLoginIP) {
		this.lastLoginIP = lastLoginIP;
	}

	private Object extra;

	public Object getExtra() {
		return extra;
	}

	public void setExtra(Object extra) {
		this.extra = extra;
	}

	public User() {
	}
}
