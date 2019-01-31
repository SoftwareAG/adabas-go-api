#include <stdio.h>
#include "adaapi.h"

int main(void)
{
    uint64_t hdl;
    int err;
    char data[255];
    hdl = ada_new_connection("acj;target=24;config=[24,4]");

    err = ada_send_search(hdl, 11, "AA,AB", "AA=11100301");
    fprintf(stdout, "Got return %d\n", err);

    ada_get_record_value(hdl, 1, "AA", data);
    fprintf(stdout, "Data AA -> %s\n", data);
    ada_get_record_value(hdl, 1, "AC", data);
    fprintf(stdout, "Data AC -> %s\n", data);
    ada_get_record_value(hdl, 1, "AD", data);
    fprintf(stdout, "Data AD -> %s\n", data);
    ada_get_record_value(hdl, 1, "AE", data);
    fprintf(stdout, "Data AE -> %s\n", data);
    ada_close_connection(hdl);
}
