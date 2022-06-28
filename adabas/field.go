package adabas

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// FieldQuery parse result of the field part of the query
type FieldQuery struct {
	Prefix        rune
	Name          string
	PeriodicIndex uint32
	MultipleIndex uint32
}

// NewFieldQuery parse field and return query info
func NewFieldQuery(field string) (fq *FieldQuery, err error) {
	if field == "" {
		return nil, adatypes.NewGenericError(186)
	}
	r, rs := utf8.DecodeRuneInString(field)
	if !strings.ContainsRune("@#", r) {
		r = ' '
		rs = 0
	}
	var re = regexp.MustCompile(`(?P<field>[^\[\(\]\)]+)(\[(?P<if>[\dN]+),?(?P<it>[\dN]*)\])?(\[([\dN]*)\])?(\((?P<ps>\d+),(?P<pt>\d+)\))?`)
	mt := re.FindStringSubmatch(field[rs:])

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
	return &FieldQuery{Name: fl, Prefix: r, PeriodicIndex: uint32(pe), MultipleIndex: uint32(mu)}, nil
}

func parseField(field string) (string, []uint32) {
	var re = regexp.MustCompile(`(?m)^(\w\w+)$|^(\w\w+)\[([N\d]*)\]$|^(\w\w+)\[([N\d]*),([N\d]*)\]$|^(\w\w+)\[([N\d]*)\]\[([N\d]*)\]$`)
	for _, match := range re.FindAllStringSubmatch(field, -1) {
		switch {
		case match[1] != "":
			return match[1], []uint32{}
		case match[2] != "":
			index := make([]uint32, 0)
			idx, err := parseIndex(match[3])
			if err != nil {
				return "", nil
			}
			index = append(index, idx)
			return match[2], index
		case match[4] != "":
			index := make([]uint32, 0)
			idx, err := parseIndex(match[5])
			if err != nil {
				return "", nil
			}
			index = append(index, idx)
			if match[6] != "" {
				idx, err := parseIndex(match[6])
				if err != nil {
					return "", nil
				}
				index = append(index, idx)
			}
			return match[4], index
		case match[7] != "":
			index := make([]uint32, 0)
			idx, err := parseIndex(match[8])
			if err != nil {
				return "", nil
			}
			index = append(index, idx)
			idx, err = parseIndex(match[9])
			if err != nil {
				return "", nil
			}
			index = append(index, idx)
			return match[7], index
		}
	}
	return "", nil
}

func parseIndex(strIndex string) (uint32, error) {
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Parse index -> %s", strIndex)
	}
	if strIndex == "N" {
		return math.MaxUint32, nil
	}
	i64, err := strconv.ParseUint(strIndex, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint32(i64), nil
}
