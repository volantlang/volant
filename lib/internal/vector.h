#ifndef VO_INTERNAL_VECTOR
#define VO_INTERNAL_VECTOR

#include "heap.h"

typedef struct VectorLayout {
    bool is_owner;
    size_t length;
    size_t capacity;
    char *mem;
} VectorLayout;

VectorLayout *_vector_new(size_t);
VectorLayout *_vector_copy(VectorLayout *, char *, size_t, size_t);
VectorLayout *_vector_resize(VectorLayout *, size_t, size_t);
VectorLayout *_vector_concat(VectorLayout *, VectorLayout *, size_t);
VectorLayout *_vector_slice(VectorLayout *, size_t, size_t, size_t);
void _vector_free(VectorLayout *);

#define VECTOR_TYPE(type)   \
    struct {                \
        bool is_owner;      \
        size_t length;      \
        size_t capacity;    \
        type *mem;          \
    } *  

#define VECTOR_NEW(type) ((void *)_vector_new(sizeof(type)))
#define VECTOR_RESIZE(vector, num) ((void *)_vector_resize((VectorLayout*)vector, sizeof(VectorLayout) + (num + vector->capacity), sizeof(*vector->mem)))
#define VECTOR_COPY(vector, val, len) ((void *)_vector_copy((VectorLayout *)vector, val, len, sizeof(*vector->mem)))
#define VECTOR_RAW(vector) (vector->mem)

#define VECTOR_PUSH(vector, value)              \
    ({ if(vector->length == vector->capacity){  \
        VECTOR_RESIZE(vector, 8);               \
    }                                           \
    vector->mem[vector->length++] = value; })    

#define VECTOR_APPEND(vector, ptr, length) ((void *)_vector_append((VectorLayout *)vector, (char *)ptr, length*(sizeof(*ptr)), sizeof(*vector->mem)))
#define VECTOR_POP(vector) (vector->mem[--vector->length])  
#define VECTOR_CONCAT(vector1, vector2) ((void *)_vector_concat((VectorLayout *)vector1, (VectorLayout *)vector2, sizeof(*vector1->mem)))
#define VECTOR_FREE(vector) (_vector_free((VectorLayout *)vector))
#define VECTOR_CLONE(vector) ((void *)_vector_clone((VectorLayout *)vector, sizeof(*vector->mem)))
#define VECTOR_SLICE(vector, start, len) ((void *)_vector_slice((VectorLayout *)vector, start, len, sizeof(*vector->mem)))

#define VECTOR_FOREACH(vector, block)                 \
    for(size_t i = 0; i < vector->length; ++i){       \
        typeof(vector->mem[0]) it = vector->mem[i];   \
        block;                                        \
    }

VectorLayout *_vector_new(size_t size_of_each_element){
    VectorLayout *vector = malloc(sizeof(VectorLayout));
    vector->mem = malloc(size_of_each_element*8);
    vector->length = 0;
    vector->capacity = 8;
    vector->is_owner = true;
    return vector;
}

VectorLayout *_vector_copy(VectorLayout *vec, char *mem, size_t len, size_t el_size) {
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

VectorLayout *_vector_resize(VectorLayout *vector, size_t new_length, size_t el_size){
    char *temp = (char *)vector->mem;
    vector->capacity = new_length;
    vector->mem = malloc(new_length*el_size);
    
    _memcpy(vector->mem, temp, vector->length*el_size);
    
    if(vector->is_owner){
        free(temp);
    }
    return vector;
}

VectorLayout *_vector_concat(VectorLayout *first, VectorLayout *second, size_t el_size) {
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

VectorLayout *_vector_append(VectorLayout *vector, char *mem, size_t size, size_t el_size) {
    size_t newSize = vector->length + size;
    if(vector->capacity < newSize){
        _vector_resize(vector, newSize, el_size);
    }
    size_t size1 = vector->length*el_size;

    for(size_t i = 0; i < size; ++i) {
        vector->mem[i+size1] = *(mem+i);
    }
    return vector;
}

VectorLayout *_vector_clone(VectorLayout *vec, size_t el_size) {
    VectorLayout *newVec = _vector_new(el_size);
    return _vector_copy(newVec, vec->mem, vec->length, el_size);
}

VectorLayout *_vector_slice(VectorLayout *vec, size_t start, size_t length, size_t el_size){
    VectorLayout *newVec = malloc(sizeof(VectorLayout));
    newVec->mem = vec->mem+(start*el_size);
    newVec->is_owner = false;
    newVec->length = length;
    newVec->capacity = vec->capacity-start;
    return newVec;
}

void _vector_free(VectorLayout *vector) {
    free(vector->mem);
    free(vector);
}

#endif