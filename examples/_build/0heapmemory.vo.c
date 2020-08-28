#include "internal/default.h"
#include "1heap.vo.h"
#include "1mem.vo.h"
i32 (^v0_main)(void) = ^i32 (void){
	i32 (*v0_ptr) = (i32*)(v1_malloc((sizeof(i32))*100));
	{
		size_t v0_i = 0;
		while(v0_i<100){
			printf("%i ", (*(v0_ptr+v0_i)));
			(++v0_i);
		}
	}
	printf("\n\n");
	v2_set(v0_ptr, 0, (sizeof(i32))*100);
	(*(v0_ptr+10)) = 100;
	{
		i32 v0_i = 0;
		while(v0_i<100){
			printf("%i ", (*(v0_ptr+v0_i)));
			(++v0_i);
		}
	}
	printf("\n");
	v1_free(v0_ptr);
	i32 (*v0_ptr2)[10] = new3(i32,i32[10],((i32[10]){0, 0, 0, 0, 0, 0, 0, 0, 0, 0, }));
	delete(v0_ptr2);

	return 0;
};

int main() {
	return v0_main();
}