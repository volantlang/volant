#ifndef VO_INTERNAL_FUNCTION
#define VO_INTERNAL_FUNCTION

#include <callback.h>

struct Block_layout {
    void *isa;
    int flags;
    int reserved; 
    void (*invoke)(void *, ...);
    struct Block_descriptor *descriptor;
};

void *_to_c_func(void *block) {
    return alloc_callback((callback_function_t)((struct Block_layout *)block)->invoke, block); 
}

#endif