#include <stdio.h>
#include "adaapi.h"

#define DEFAULT_QUERY "AA=11100301"

int main(int argc, char **argv)
{
    uint64_t hdl;
    int64_t i8;
    int nr_records;
    char data[255];
    char query[255];
    char **fields;
    char *cur_field;
    int i;

    if (argc > 1)
    {
        fprintf(stdout, "Args: %s\n", argv[1]);
        strcpy(query, argv[1]);
    }
    else
    {
        strcpy(query, DEFAULT_QUERY);
    }
    hdl = ada_new_connection("acj;target=24;config=[24,4]");

    nr_records = ada_send_search(hdl, 11, "AA,AB,AH", query);
    fprintf(stdout, "Got return %d\n", nr_records);

    fields = ada_get_fieldnames(hdl);
    cur_field = fields[0];
    fprintf(stdout, "Field %s\n", cur_field);
    for (i = 0; cur_field != NULL; i++)
    {
        fprintf(stdout, "Field %s\n", fields[i]);
        ada_free(cur_field);
        cur_field = fields[i + 1];
    }
    ada_free(fields);

    for (i = 0; i < nr_records; i++)
    {
        fprintf(stdout, "%d.Record\n", i+1);
        ada_get_record_string_value(hdl, i + 1, "AA", data, 255);
        fprintf(stdout, " Data AA -> %s\n", data);
        ada_get_record_string_value(hdl, i + 1, "AC", data, 255);
        fprintf(stdout, " Data AC -> %s\n", data);
        ada_get_record_string_value(hdl, i + 1, "AD", data, 255);
        fprintf(stdout, " Data AD -> %s\n", data);
        ada_get_record_string_value(hdl, i + 1, "AE", data, 255);
        fprintf(stdout, " Data AE -> %s\n", data);
        ada_get_record_int64_value(hdl, i + 1, "AH", &i8);
        fprintf(stdout, " Data AH -> %lld\n", i8);
    }
    ada_close_connection(hdl);
}
