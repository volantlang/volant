#ifndef VO_INTERNAL_TYPES
#define VO_INTERNAL_TYPES

typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long long u64;

typedef char i8;
typedef short i16;
typedef int i32;
typedef long long i64;

typedef float f32;
typedef double f64;

typedef u64 uptr;
typedef u8 bool;

#define true 1
#define false 0
#define null ((void *)0)
#endif