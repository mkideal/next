/* Code generated by "next 0.0.4"; DO NOT EDIT. */

#ifndef DEMO_B_B_H
#define DEMO_B_B_H

#include <stdint.h>
#include <string.h>

#include "../a/a.next.h"

#if !defined(__cplusplus) && !defined(__bool_true_false_are_defined)
#include <stdbool.h>
#endif

#if !defined(_Bool)
typedef unsigned char _Bool;
#endif

// Enums forward declarations
typedef enum DEMO_B_TestEnum DEMO_B_TestEnum;

// Structs forward declarations
typedef struct DEMO_B_TestStruct DEMO_B_TestStruct;

typedef enum DEMO_B_TestEnum {
	DEMO_B_TestEnum_A = 1,
	DEMO_B_TestEnum_B = 5,
	DEMO_B_TestEnum_C = 5,
	DEMO_B_TestEnum_D = 10,
	DEMO_B_TestEnum_E = 20,
	DEMO_B_TestEnum_F = 1,
	DEMO_B_TestEnum_G = 2,
} DEMO_B_TestEnum;

typedef struct DEMO_B_TestStruct {
	DEMO_A_Point2D point;
} DEMO_B_TestStruct;

#endif /* DEMO_B_B_H */
