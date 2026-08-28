package sharedmodels

import (
	"encoding/json"
	"fmt"
)

type Pronoun int

const (
	PronounThey Pronoun = iota
	PronounShe
	PronounHe
)

var PronounList = []Pronoun{
	PronounThey,
	PronounShe,
	PronounHe,
}

func (p Pronoun) String() string {
	strMap := map[Pronoun]string{
		0: "0",
		1: "1",
		2: "2",
	}
	return strMap[p]
}

func (p Pronoun) Subject() string {
	strMap := map[Pronoun]string{
		0: "they",
		1: "she",
		2: "he",
	}
	return strMap[p]
}

func (p Pronoun) Object() string {
	strMap := map[Pronoun]string{
		0: "them",
		1: "her",
		2: "him",
	}
	return strMap[p]
}

func (p Pronoun) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Subject())
}

func (p *Pronoun) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for _, v := range PronounList {
		if v.Subject() == s {
			*p = v
			return nil
		}
	}
	return fmt.Errorf("unknown pronoun: %s", s)
}
