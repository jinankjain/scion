#pragma once

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "aes.h"
#include "siphash.h"

#define EPHID_HOST_LEN 8
#define EPHID_MAC_LEN 4
#define EPHID_IV_LEN 4

typedef struct {
   uint8_t iv[AES_BLOCKLEN];
   uint8_t encrypted_id[EPHID_HOST_LEN];
   uint8_t mac[EPHID_MAC_LEN];
} ephid;

ephid* parse_ephid(const uint8_t *buf) {
    ephid* e = (ephid *) malloc(sizeof(ephid));
    memset(e, 0, sizeof(ephid));
    memcpy(e->iv, buf, EPHID_IV_LEN);
    memcpy(e->encrypted_id, buf+EPHID_IV_LEN, EPHID_HOST_LEN);
    memcpy(e->mac, buf+EPHID_IV_LEN+EPHID_HOST_LEN, EPHID_MAC_LEN);
    return e;
}

const uint8_t aes_key[16] = {0x7d, 0x97, 0xc4, 0x9c, 0x33, 0xbd, 0xc5, 0xb1,
                             0x44, 0xe2, 0x26, 0x21, 0xcd, 0x6a, 0xd2, 0x01};

uint32_t decrypt_host_id(const uint8_t *buf) {
    ephid *e = parse_ephid(buf);
    struct AES_ctx ctx;
    AES_init_ctx_iv(&ctx, aes_key, e->iv);
    uint8_t plaintext[] = {0, 0, 0, 0, 0, 0, 0, 0,
                           0, 0, 0, 0, 0, 0, 0, 0};
    AES_CBC_encrypt_buffer(&ctx, plaintext, AES_BLOCKLEN);
    uint8_t temp[EPHID_HOST_LEN];
    for (int i = 0; i < EPHID_HOST_LEN; ++i) {
        temp[i] = plaintext[i] ^ e->encrypted_id[i];
    }
    uint32_t ans = (uint32_t)temp[1] | (uint32_t)(temp[2] << 8) | (uint32_t)(temp[3] << 16);
    return ans;
}

const char siphash_key[16] = {0xda, 0xe7, 0xac, 0xe5, 0xb7, 0x72, 0x3b, 0xd4,
                              0xec, 0x59, 0x86, 0xa8, 0xd2, 0x5f, 0x12, 0xc6};

uint32_t service_addr_to_siphash(const uint8_t *src, unsigned long src_sz) {
    return siphash24(src, src_sz, siphash_key) & ((1 << 25) - 1);
}
