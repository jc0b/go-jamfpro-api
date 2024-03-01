package jamfpro

func AreGroupsEquivalent(planned, actual *ComputerGroup) bool {
	if actual == nil {
		return false
	}

	if planned.Name != actual.Name {
		return false
	}
	if planned.Id != actual.Id {
		return false
	}
	for i, v := range planned.Computers {
		if v != actual.Computers[i] {
			return false
		}
	}
	for i, v := range planned.Criteria {
		if v != actual.Criteria[i] {
			return false
		}
	}

	return true
}

func AreComputerRecordsEquivalent(planned, actual *Computer) bool {
	if actual == nil {
		return false
	}

	if planned.Name != actual.Name {
		return false
	}
	if planned.Id != actual.Id {
		return false
	}
	if planned.SerialNumber != actual.SerialNumber {
		return false
	}
	return true
}
