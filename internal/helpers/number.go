package helpers

func ValidateLuhn(number string) bool {
	var sum int
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
