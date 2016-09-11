package view

func itoidrune(i int) rune {
	if i < 10 {
		return rune(48 + i)
	}
	return rune(87 + i)
}

func buildingChar() string {
	if AvoidUnicode {
		return "B"
	}
	return "⟳"
}

func failedChar() string {
	if AvoidUnicode {
		return "X"
	}
	return "✖"
}

func successChar() string {
	if AvoidUnicode {
		return "V"
	}
	return "✓"
}

func undefinedChar() string {
	if AvoidUnicode {
		return "U"
	}
	return "⁈"
}

func unknownChar() string {
	if AvoidUnicode {
		return "?"
	}
	return "❓"
}
