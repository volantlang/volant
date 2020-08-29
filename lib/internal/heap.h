#ifndef VO_INTERNAL_HEAP
#define VO_INTERNAL_HEAP

#include "gc.h"

#define malloc(size) GC_malloc(size)
#define realloc(ptr, size) GC_realloc(ptr, size)
#define calloc(size1, size2) GC_malloc(size1*size2)
#define free(ptr) GC_free(ptr)

#endif
