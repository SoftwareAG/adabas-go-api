package adabas

import (
	"regexp"
	"strconv"
)

type fieldQuery struct {
	Name          string
	PeriodicIndex uint32
	MultipleIndex uint32
}

func NewFieldQuery(field string) (fq *fieldQuery, err error) {
	var re = regexp.MustCompile(`(?P<field>[^\[\(\]\)]+)(\[(?P<if>[\dN]+),?(?P<it>[\dN]*)\])?(\[([\dN]*)\])?(\((?P<ps>\d+),(?P<pt>\d+)\))?`)
	mt := re.FindStringSubmatch(field)

	fl := mt[1]
	var pe, mu uint64
	peIndex := mt[3]
	if peIndex != "" {
		pe, err = strconv.ParseUint(peIndex, 10, 32)
		if err != nil {
			return nil, err
		}
	}
	muIndex := mt[4]
	if muIndex != "" {
		mu, err = strconv.ParseUint(muIndex, 10, 32)
		if err != nil {
			return nil, err
		}
	} else {
		muIndex := mt[6]
		if muIndex != "" {
			mu, err = strconv.ParseUint(muIndex, 10, 32)
			if err != nil {
				return nil, err
			}
		}
	}
	return &fieldQuery{Name: fl, PeriodicIndex: uint32(pe), MultipleIndex: uint32(mu)}, nil
}
