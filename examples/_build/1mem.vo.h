#ifndef H_2
#define H_2
#include "internal/default.h"
void* (^v2_copy)(void*, void*, size_t);
i8 (^v2_compare)(void*, void*, size_t);
void* (^v2_set)(void*, u8, size_t);
void* (^v2_copy)(void*, void*, size_t) =^void* (void (*v2_dest), void (*v2_src), size_t v2_length){
	u8 (*v2_l) = (u8*)(v2_dest);	u8 (*v2_r) = (u8*)(v2_src);
	{
		size_t v2_i = 0;
		while(v2_i<v2_length){
			v2_l[v2_i] = v2_r[v2_i];
			(++v2_i);
		}
	}
	return v2_dest;
};i8 (^v2_compare)(void*, void*, size_t) =^i8 (void (*v2_first), void (*v2_second), size_t v2_length){
	u8 (*v2_l) = (u8*)(v2_first);	u8 (*v2_r) = (u8*)(v2_second);

	while((v2_length!=0)&&(((*v2_l))==((*v2_r)))){
		(--v2_length);
		(++v2_l);
		(++v2_r);
	}
	return v2_length ? (i8)(((*v2_l))-((*v2_r))) : 0;
};void* (^v2_set)(void*, u8, size_t) =^void* (void (*v2_ptr), u8 v2_char, size_t v2_length){
	u8 (*v2_l) = (u8*)(v2_ptr);
	{
		size_t v2_i = 0;
		while(v2_i<v2_length){
			v2_l[v2_i] = v2_char;
			(++v2_i);
		}
	}
	return v2_ptr;
};
#endif