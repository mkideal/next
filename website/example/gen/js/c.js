// Code generated by "next 0.0.5"; DO NOT EDIT.

import * as a from './a.js';
import * as b from './b.js';

export const A = 1;
export const B = "hello";
export const C = 3.14;
export const D = true;

export const Color = Object.freeze({
	Red: 0,
	Green: 1,
	Blue: 2
});

export const LoginType = Object.freeze({
	Username: 1,
	Email: 2
});

export const UserType = Object.freeze({
	Admin: 1,
	User: 2
});

export class User {
	constructor() {
		/** @type { 'number' } */
		this.type = UserType[Object.keys(UserType)[0]];
		/** @type { Number } */
		this.id = 0;
		/** @type { String } */
		this.username = "";
		/** @type { String } */
		this.password = "";
		/** @type { String } */
		this.deviceId = "";
		/** @type { String } */
		this.twoFactorToken = "";
		/** @type { 'Array<String>' } */
		this.roles = [];
		/** @type { 'Map<String, String>' } */
		this.metadata = new Map();
		/** @type { 'Array<Number>' } */
		this.scores = [];
	}
}

export class LoginRequest {
	constructor() {
		/** @type { 'number' } */
		this.type = LoginType[Object.keys(LoginType)[0]];
		/** @type { String } */
		this.username = "";
		/** @type { String } */
		this.password = "";
		/** @type { String } */
		this.deviceId = "";
		/** @type { String } */
		this.twoFactorToken = "";
	}
}

export class LoginResponse {
	constructor() {
		/** @type { Boolean } */
		this.success = false;
		/** @type { String } */
		this.errorMessage = "";
		/** @type { String } */
		this.authenticationToken = "";
		/** @type { User } */
		this.user = new User;
	}
}
