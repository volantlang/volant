#ifndef VO_DEFAULT
#define VO_DEFAULT

#include <stdio.h>
#include <string.h>

#include "types.h"
#include "heap.h"
#include "vector.h"
// #include "libuv.h"
#include "promise.h"
#include "function.h"

#include "Block.h"

#define new(type, size_type) (type *)malloc(sizeof(size_type))
#define new2(type, size_type, val) (type *)({type *ptr = malloc(sizeof(size_type)); *ptr = (type)val; ptr;})
#define new3(type, size_type, val) (type *)({type *ptr = malloc(sizeof(size_type)); memcpy(ptr, val, sizeof(size_type)); ptr;})
#define new4(type, val, size) (VECTOR_COPY(VECTOR_NEW(type), (char *)val, size))
#define new5(type) PROMISE_NEW(type)

#define delete(block) free(block);

#define len(array, base_type) (sizeof(array)/sizeof(base_type))
#define len2(array, base_type) sizeof(array)/sizeof(base_type)
#define len3(var) sizeof(var)

#endif