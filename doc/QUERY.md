# API Search concept

<!-- TOC -->

- [API Search concept](#api-search-concept)
	- [Content](#content)
	- [Adabas search adaptions](#adabas-search-adaptions)
		- [Common search](#common-search)
		- [Search ranges](#search-ranges)
	- [Special searches](#special-searches)

<!-- /TOC -->

## Content

This document describes the search capability of the Go Adabas API. The Go API uses one-to-one match of the Adabas capabilities.

Adabas Go API provides search queries based on Adabas queries. The search syntax is limited. See Adabas documentation.

## Adabas search adaptions

Following search queries are possible. It is independent if the Adabas Map or the Adabas field short-name is used.

### Common search

It is possible to search a field is containing a special value. For example this query will search for the field `PERSONNEL-ID` to be `40003001`.

```sql
PERSONNEL-ID=40003001
```

Similar approach is to search for Alpha or Unicode fields using brackets like

```sql
FIRSTNAME='ADAM'
```

The search value can be linked together using the `AND` or `OR` keywords. For example this will search for all with name equals `SMITH` and first name equals `ADAM`.

```sql
FIRSTNAME='ADAM' AND NAME='SMITH'
```

It is possible to use the 'greater then' and 'lower then' queries. Here an example for the 'greater then' query

```sql
NAME>'SMITH' OR NUMBER>10
```

or for a 'lower then'

```sql
NAME<'SMITH' AND NUMBER<10
```

or for a 'lower then'

```sql
NAME<='SMITH'
```

Similar approach is to not equals. Here an example

```sql
NUMBER!=10
```

### Search ranges

Adabas provides the possibility to search for ranges. Inside the API the range search is providing with or without first range start value. Corresponding it is with last range value. This example will search in the range of `40003001` to `40005001` including the two values.

```sql
PERSONNEL-ID=[40003001:40005001]
```

This example will exclude the first range value `40003001`:

```sql
PERSONNEL-ID=(40003001:40005001]
```

It is possible to search for Alpha or Unicode field ranges as well. Here an example which search for all strings beginning on `SMITH` up to `Y`:

```sql
PERSONELL_ID=['SMITH':'Y']
```

Ranges can be combined with `AND` or `OR`:

```sql
NAME='ADAM' AND NAME=['FR':'FRZ']
```

## Special searches

Sometime it is needed to use special characters.
Search using the hexadecimal value of a number:

```sql
NUMBER=0xF1
```

Following example search for a Super descriptor with a repeated `0x21` at the end:

```sql
S2='BADABAS__'0x21*

S2=['BADABAS__'0:'BADABAS__'255]

S2=['BADABAS__'0x00:'BADABAS__'0xFF(10)]

S2=['BADABAS__'0x00:'BADABAS__'0xFFFFFFFFFFFFFFFFFFFF]
```