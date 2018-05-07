package steam

import (
	"math/big"
	"regexp"
	"strings"
	"strconv"
)

func (s Steam) GetID(in string) (out int64, err error) {

	if regexp.MustCompile(`^STEAM_([01]):([01]):[0-9][0-9]{0,8}$`).MatchString(in) { // STEAM_0:0:4180232

		return convert0to64(in), nil

	} else if regexp.MustCompile(`(\[)?U:1:\d+(])?`).MatchString(in) { // [U:1:8360464]

		return convert3to64(in), nil

	} else if regexp.MustCompile(`^\d{17}$`).MatchString(in) { // 76561197968626192

		i, err := strconv.ParseInt(in, 10, 64)
		if err != nil {
			return out, err
		}
		return i, nil

	} else if regexp.MustCompile(`^\d{1,16}$`).MatchString(in) { // 8360464

		return convert32to64(in), nil

	} else {

		resp, _, err := s.ResolveVanityURL(in)
		if err != nil {
			return out, err
		}

		return resp.SteamID, nil
	}
}

func convert3to64(in string) (out int64) {

	parts := strings.Split(in, ":")
	part := parts[2]
	part = part[:len(part)-1] // Remove bracket
	return convert32to64(part)
}

func convert32to64(in string) (out int64) {

	inBig, _ := new(big.Int).SetString(in, 10)
	mul, _ := new(big.Int).SetString("76561197960265728", 10)

	return inBig.Add(inBig, mul).Int64()
}

func convert0to64(in string) (out int64) {

	parts := strings.Split(in, ":")
	add, _ := new(big.Int).SetString("76561197960265728", 10)
	level, _ := new(big.Int).SetString(parts[1], 10)

	ID64, _ := new(big.Int).SetString(parts[2], 10)
	ID64 = ID64.Mul(ID64, big.NewInt(2))
	ID64 = ID64.Add(ID64, add)
	ID64 = ID64.Add(ID64, level)

	return ID64.Int64()
}
