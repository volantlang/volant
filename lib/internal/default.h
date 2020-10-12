#ifndef VO_DEFAULT
#define VO_DEFAULT

#include <stdio.h>
#include <sys/time.h>
#include <sys/socket.h>
#include "types.h"

void _memcpy(void *dest, void *src, size_t size) {
    char *l = (char *)dest;
    char *r = (char *)src;
    for(size_t i = 0; i < size; ++i){
        l[i] = r[i];
    }
}

#ifndef VO_NO_HEAP
#   include "heap.h"
#   include "vector.h"
#   include "promise.h"

#   define new(type, size_type) (type *)malloc(sizeof(size_type))
#   define new2(type, size_type, val) (type *)({type *ptr = malloc(sizeof(size_type)); *ptr = (type)val; ptr;})
#   define new3(type, size_type, val) (type *)({type *ptr = malloc(sizeof(size_type)); _memcpy(ptr, val, sizeof(size_type)); ptr;})
#   define new4(type, val, size) ({ VECTOR_TYPE(type) vector = VECTOR_NEW(type); VECTOR_COPY(vector, (char *)val, size); (void *)vector; }) 
#   define new5(type) PROMISE_NEW(type)

#   define delete(block) free(block);
#endif

#ifndef VO_NO_BLOCKSRUNTIME
#   include "Block.h"
#else
void * _NSConcreteStackBlock[32] = { 0 };
void * _NSConcreteMallocBlock[32] = { 0 };
void * _NSConcreteAutoBlock[32] = { 0 };
void * _NSConcreteFinalizingBlock[32] = { 0 };
void * _NSConcreteGlobalBlock[32] = { 0 };
void * _NSConcreteWeakBlockVariable[32] = { 0 };
#endif

#endif