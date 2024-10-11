/* Code generated by "next 0.0.5"; DO NOT EDIT. */

#ifndef DEMO_C_C_H
#define DEMO_C_C_H

#include <stdint.h>
#include <string.h>

#include "../a/a.next.h"
#include "../b/b.next.h"

#if !defined(__cplusplus) && !defined(__bool_true_false_are_defined)
#include <stdbool.h>
#endif

#if !defined(_Bool)
typedef unsigned char _Bool;
#endif

// Enums forward declarations
typedef enum DEMO_C_Color DEMO_C_Color;
typedef enum DEMO_C_LoginType DEMO_C_LoginType;
typedef enum DEMO_C_UserType DEMO_C_UserType;

// Structs forward declarations
typedef struct DEMO_C_User DEMO_C_User;
typedef struct DEMO_C_LoginRequest DEMO_C_LoginRequest;
typedef struct DEMO_C_LoginResponse DEMO_C_LoginResponse;

#define DEMO_C_A 1
#define DEMO_C_B "hello"
#define DEMO_C_C 3.14
#define DEMO_C_D 1

typedef enum DEMO_C_Color {
	DEMO_C_Color_Red = 0,
	DEMO_C_Color_Green = 1,
	DEMO_C_Color_Blue = 2,
} DEMO_C_Color;

typedef enum DEMO_C_LoginType {
	DEMO_C_LoginType_Username = 1,
	DEMO_C_LoginType_Email = 2,
} DEMO_C_LoginType;

typedef enum DEMO_C_UserType {
	DEMO_C_UserType_Admin = 1,
	DEMO_C_UserType_User = 2,
} DEMO_C_UserType;

typedef struct DEMO_C_User {
	DEMO_C_UserType type;
	int32_t id;
	char* username;
	char* password;
	char* device_id;
	char* two_factor_token;
	void* roles;
	void* metadata;
	int32_t scores[4];
} DEMO_C_User;

typedef struct DEMO_C_LoginRequest {
	DEMO_C_LoginType type;
	char* username;
	char* password;
	char* device_id;
	char* two_factor_token;
} DEMO_C_LoginRequest;

typedef struct DEMO_C_LoginResponse {
	_Bool success;
	char* error_message;
	char* authentication_token;
	DEMO_C_User user;
} DEMO_C_LoginResponse;

#endif /* DEMO_C_C_H */
