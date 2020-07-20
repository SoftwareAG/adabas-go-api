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

This document describes the search capability of the Adabas API for Go. The Go API uses a one-to-one match of the Adabas capabilities.

The Adabas API for Go provides search queries based on Adabas queries. The search syntax is limited. See the Adabas documentation.

## Adabas search adaptions

The following search queries are possible. The queries do not depend on whether the Adabas Map or the Adabas field short-name is used.

### Common search

It is possible to search for a special value in a field. For example this query will search for the value of field `PERSONNEL-ID` to be `40003001`.

```sql
PERSONNEL-ID=40003001
```

A similar approach is to search for Alpha or Unicode fields using brackets like:

```sql
FIRSTNAME='ADAM'
```

Several search values can be linked together using the `AND` or `OR` keywords. For example, this will search for all records where NAME is `SMITH` and FIRSTNAME is `ADAM`.

```sql
FIRSTNAME='ADAM' AND NAME='SMITH'
```

It is possible to use 'greater then' and 'less then' queries. Here an example for the 'greater than' query:

```sql
NAME>'SMITH' OR NUMBER>10
```

or for a 'less then'

```sql
NAME<'SMITH' AND NUMBER<10
```

or for a 'less than or equal to'

```sql
NAME<='SMITH'
```

There is a similar approach for 'not equal'. Here is an example:

```sql
NUMBER!=10
```

### Search ranges

Adabas provides the possibility to search for ranges using the syntax `[start:end]`. The `start` or `end` value can be omitted, but not both. This example will search in the range of `40003001` to `40005001` including the two values:

```sql
PERSONNEL-ID=[40003001:40005001]
```

Round brackets can be used to exclude the start or end value of the range, or both.

This example will exclude the first range value `40003001`:

```sql
PERSONNEL-ID=(40003001:40005001]
```

It is possible to search for Alpha or Unicode field ranges as well. This example searches for all strings beginning with `SMITH` up to `Y`:

```sql
PERSONELL_ID=['SMITH':'Y']
```

Ranges can be combined with `AND` or `OR`:

```sql
NAME='ADAM' AND NAME=['FR':'FRZ']
```

## Special searches

Sometimes it is needed to use special characters.
Search using the hexadecimal value of a number:

```sql
NUMBER=0xF1
```

The following example searches for a superdescriptor with a repeated `0x21` at the end:

```sql
S2='BADABAS__'0x21*
```

Similar examples are:

```
S2=['BADABAS__'0:'BADABAS__'255]

S2=['BADABAS__'0x00:'BADABAS__'0xFF(10)]

S2=['BADABAS__'0x00:'BADABAS__'0xFFFFFFFFFFFFFFFFFFFF]
```