#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <pthread.h>
#include <atomic>
#include <thread>
#include "art.h"

#include "lock.h"

// int main(int argc, char *argv[]) {
//   art_tree tree;
//
//   art_tree_init(&tree);
//   art_tree_init_lock(&tree); // create exclusive lock
//
//   art_insert(&tree, (const unsigned char*)"00001", 5, NULL_VALUE);
//   art_insert(&tree, (const unsigned char*)"00002", 5, NULL_VALUE);
//
//   printf("bytes: %u\n", (unsigned int)tree.bytes);
//
//   art_delete(&tree, (const unsigned char*)"00002", 5);
//   printf("bytes: %u\n", (unsigned int)tree.bytes);
//   printf("sizeof lock: %u\n", (unsigned int)moontrade_rwlock_size());
//   art_tree_destroy(&tree);
// }
