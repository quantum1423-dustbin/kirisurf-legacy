#include <stdint.h>
 
void tea_encrypt_block (uint32_t* v, uint32_t* k) {
    uint32_t v0=v[0], v1=v[1], sum=0, i;           /* set up */
    uint32_t delta=0x9e3779b9;                     /* a key schedule constant */
    uint32_t k0=k[0], k1=k[1], k2=k[2], k3=k[3];   /* cache key */
    for (i=0; i < 32; i++) {                       /* basic cycle start */
        sum += delta;
        v0 += ((v1<<4) + k0) ^ (v1 + sum) ^ ((v1>>5) + k1);
        v1 += ((v0<<4) + k2) ^ (v0 + sum) ^ ((v0>>5) + k3);
    }                                              /* end cycle */
    v[0]=v0; v[1]=v1;
}

// We implement the CTR in C too, since FFI is slow
void tea_ctr_stream (uint64_t start_idx, uint64_t end_idx, char *outbuf, uint64_t outlen) {
	// We first make the rounded version (with unnecessary cruft at beginning and end)
	uint64_t rsi = (start_idx / 8);
	uint64_t rei = ((end_idx + 7) / 8);
	
}