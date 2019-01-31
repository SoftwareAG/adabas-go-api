#include "adaapi.h"

int main (void){
    uint64_t hdl;
    hdl = ada_new_connection("acj;map;config=[24,4]");
    ada_close_connection(hdl);
}
