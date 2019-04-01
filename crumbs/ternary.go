package crumbs

// TernaryString condition ? trueVal : falseVal
func TernaryString(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

// TernaryInt condition ? trueVal : falseVal
func TernaryInt(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

// TernaryInt64 condition ? trueVal : falseVal
func TernaryInt64(condition bool, trueVal, falseVal int64) int64 {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

// TernaryFloat64 condition ? trueVal : falseVal
func TernaryFloat64(condition bool, trueVal, falseVal float64) float64 {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

// TernaryMapString condition ? trueVal : falseVal
func TernaryMapString(condition bool, trueVal, falseVal map[string]string) map[string]string {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

// TernaryMap condition ? trueVal : falseVal
func TernaryMap(condition bool, trueVal, falseVal map[string]interface{}) map[string]interface{} {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

// TernaryInterface condition ? trueVal : falseVal
func TernaryInterface(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}
