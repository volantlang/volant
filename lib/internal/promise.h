#ifndef VO_INTERNAL_PROMISE
#define VO_INTERNAL_PROMISE

#include "heap.h"
#include "vector.h"

#define PROMISE_TYPE(type)                       \
    struct {                                     \
        u8 state;                                \
        VECTOR_TYPE(void *) listeners;           \
        type val;                                \
    } *

#define PROMISE_NEW(type) (void *)({PROMISE_TYPE(type) prom = malloc(sizeof(PROMISE_TYPE(type))); prom->listeners = VECTOR_NEW(void *); prom->state = 0; prom; })
#define PROMISE_THEN(promise, callback) VECTOR_PUSH(promise->listeners, (void *)callback)
#define PROMISE_RESOLVE(promise, value) ({ promise->val = value; promise->state = 1; VECTOR_FOREACH(promise->listeners, ({ ((void (^)(typeof(promise->val)))it)(promise->val); })); })

#endif