
#ifndef VO_DEFAULT
#define VO_DEFAULT

#include <stdio.h>
#include <string.h>
#include "gc.h"
#include "uv.h"
#include "Block.h"

#define malloc(x) GC_MALLOC(x)
#define free(x) GC_FREE(x)

#define new(type, size_type) (type *)malloc(sizeof(size_type))
#define new2(type, size_type, val) (type *)({type *ptr = malloc(sizeof(size_type)); *ptr = (type)val; ptr;})
#define new3(type, size_type, val) (type *)({type *ptr = malloc(sizeof(size_type)); memcpy(ptr, val, sizeof(size_type)); ptr;})

#define delete(block) free(block);

#define len(block, base_type) (size(block)/sizeof(base_type))
#define len2(array, base_type) sizeof(array)/sizeof(base_type)
#define len3(var) sizeof(var)

#define size(block) *((size_t *)block-1)
#define size2(var) sizeof(var)

#define cast(val, type) ((type)val)

typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long u64;

typedef char i8;
typedef short i16;
typedef int i32;
typedef long i64;

typedef struct {
    size_t size;
    char mem[];
} __mem_block;

/*
void *_new(size_t type) {
    __mem_block *block = (__mem_block *)malloc(type+sizeof(size_t));
    block->size = type;
    return block->mem;
}
*/

typedef union uv_any_handle uv_any_handle;
typedef union uv_any_req uv_any_req;
#endif