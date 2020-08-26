#ifndef VO_INTERNAL_VECTOR
#define VO_INTERNAL_VECTOR

#include "heap.h"

typedef struct BasicVector {
    size_t length;
    size_t capacity;
    char mem[];
} BasicVector;

BasicVector *_vector_new(size_t);
BasicVector *_vector_copy(BasicVector *, char *, size_t, size_t);
BasicVector *_vector_resize(BasicVector *, size_t, size_t);
BasicVector *_vector_concat(BasicVector *, BasicVector *, size_t);
void _vector_free(BasicVector *);

#define VECTOR_TYPE(type)   \
    struct {                \
        size_t length;      \
        size_t capacity;    \
        type mem[];         \
    } *  

#define VECTOR_NEW(type) (void *)_vector_new(sizeof(type))
#define VECTOR_RESIZE(vector, num) _vector_resize((BasicVector*)vector, sizeof(BasicVector) + (num + vector->capacity), sizeof(*vector->mem)) 
#define VECTOR_COPY(vector, val, len) (void *)_vector_copy((BasicVector *)vector, val, len, sizeof(*vector->mem))

#define VECTOR_PUSH(vector, value)              \
    ({ if(vector->length == vector->capacity){  \
        VECTOR_RESIZE(vector, 8);               \
    }                                           \
    vector->mem[vector->length++] = value; })     

#define VECTOR_POP(vector) ({ vector->length--; vector->mem[vector->length];})  
#define VECTOR_CONCAT(vector1, vector2) (void *)_vector_concat((BasicVector *)vector1, (BasicVector *)vector2, sizeof(*vector1->mem))
#define VECTOR_FREE(vector) _vector_free((BasicVector *)vector)
#define VECTOR_CLONE(vector) (void *)_vector_clone((BasicVector *)vector, sizeof(*vector->mem))

#define VECTOR_FOREACH(vector, block)                 \
    for(size_t i = 0; i < vector->length; ++i){       \
        typeof(vector->mem[0]) it = vector->mem[i];   \
        block;                                        \
    }

BasicVector *_vector_new(size_t size_of_each_element){
    BasicVector *vector = malloc(sizeof(BasicVector)+size_of_each_element*8);
    vector->length = 0;
    vector->capacity = 8;
    return vector;
}

BasicVector *_vector_copy(BasicVector *vec, char *mem, size_t len, size_t el_size) {
    if(vec->capacity < len){
        _vector_resize(vec, len, el_size);
    }
    size_t size = len*el_size;
    for(size_t i = 0; i < size; ++i) {
        vec->mem[i] = mem[i];
    }
    vec->length = len;
    return vec;
}

BasicVector *_vector_resize(BasicVector *vector, size_t new_length, size_t el_size){
    vector = realloc(vector, sizeof(BasicVector)+new_length*el_size);
    vector->capacity = new_length;
    return vector;
}

BasicVector *_vector_concat(BasicVector *first, BasicVector *second, size_t el_size) {
    size_t newLength = first->length + second->length;
    if(first->capacity < newLength){
        _vector_resize(first, newLength, el_size);
    } 
    size_t size1 = first->length*el_size;
    size_t size2 = second->length*el_size;
    
    for(size_t i = 0; i < size2; ++i) {
        first->mem[i+size1] = second->mem[i];
    }
    first->length = newLength;
    return first;
}

void _vector_free(BasicVector *vector) {
    free(vector->mem);
    free(vector);
}

BasicVector *_vector_clone(BasicVector *vec, size_t el_size) {
    BasicVector *newVec = _vector_new(el_size);
    return _vector_copy(newVec, vec->mem, vec->length, el_size);
}
#endif