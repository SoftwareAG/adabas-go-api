#include <stdio.h>
#include "adaapi.h"

#define DEFAULT_QUERY "PERSONNEL-ID=11100301"

int main (int argc,char **argv){
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
    hdl = ada_new_connection("acj;map;config=[24,4]");

    nr_records = ada_send_msearch(hdl, "EMPLOYEES-NAT-DDM", "PERSONNEL-ID,FULL-NAME,BIRTH", query);
    fprintf(stdout, "Got return %d\n", nr_records);

    fields = ada_get_fieldnames(hdl);
    cur_field = fields[0];
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
        ada_get_record_string_value(hdl, i + 1, "PERSONNEL-ID", data, 255);
        fprintf(stdout, " Data PERSONNEL-ID -> %s\n", data);
        ada_get_record_string_value(hdl, i + 1, "FIRST-NAME", data, 255);
        fprintf(stdout, " Data FIRST-NAME -> %s\n", data);
        ada_get_record_string_value(hdl, i + 1, "MIDDLE-I", data, 255);
        fprintf(stdout, " Data MIDDLE-I -> %s\n", data);
        ada_get_record_string_value(hdl, i + 1, "NAME", data, 255);
        fprintf(stdout, " Data NAME -> %s\n", data);
        ada_get_record_int64_value(hdl, i + 1, "BIRTH", &i8);
        fprintf(stdout, " Data BIRTH -> %lld\n", i8);
    }
    ada_close_connection(hdl);
}
