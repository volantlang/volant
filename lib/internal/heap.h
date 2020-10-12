#ifndef VO_INTERNAL_HEAP
#define VO_INTERNAL_HEAP

#ifndef VO_NO_GC
#   include "gc.h"
#   define malloc(size) GC_malloc(size)
#   define realloc(ptr, size) GC_realloc(ptr, size)
#   define calloc(size1, size2) GC_malloc(size1*size2)
#   define free(ptr) GC_free(ptr)
#else 
#   include <stdlib.h>
#endif

#endif
