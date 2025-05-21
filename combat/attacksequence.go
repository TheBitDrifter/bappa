package combat

type AttackSequence struct {
	Attacks []Attack
}

type BufferWindow struct {
	Start, End int
}

func NewAttackSeq(attacks ...Attack) *AttackSequence {
	return &AttackSequence{Attacks: attacks}
}

func (aSeq AttackSequence) First() Attack {
	return aSeq.Attacks[0]
}

func (aSeq AttackSequence) Next(current Attack) (Attack, bool) {
	for i, a := range aSeq.Attacks {
		if a.ID == current.ID && i != len(aSeq.Attacks) {
			return aSeq.Attacks[i+1], true
		}
	}
	return Attack{}, false
}
