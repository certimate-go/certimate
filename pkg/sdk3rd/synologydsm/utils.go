package synologydsm

func getAuthErrorDescription(code int) string {
	switch code {
	case 100:
		return "Unknown error"
	case 101:
		return "Invalid parameters"
	case 102:
		return "API does not exist"
	case 103:
		return "Method does not exist"
	case 104:
		return "This API version is not supported"
	case 105:
		return "Insufficient user privilege"
	case 106:
		return "Connection time out"
	case 107:
		return "Multiple login detected"
	case 400:
		return "Invalid password or account does not exist"
	case 401:
		return "Guest or disabled account"
	case 402:
		return "Permission denied"
	case 403:
		return "2-factor authentication code required (OTP)"
	case 404:
		return "Failed to authenticate 2-factor authentication code"
	case 405:
		return "Server version is too low or not supported"
	case 406:
		return "2-factor authentication code expired"
	case 407:
		return "Login failed: IP has been blocked"
	case 408:
		return "Expired password"
	case 409:
		return "Password must be changed (password policy)"
	case 410:
		return "Account locked (too many failed login attempts)"
	default:
		return "Unknown authentication error"
	}
}
