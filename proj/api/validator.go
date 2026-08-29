package api

import (
	"github.com/apoor/simple_bank/util"
	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	// 1. Field ki value nikalo aur String mein convert karo
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// 2. Check karo ki kya string humari supported list mein aati hai
		return util.IsSupportedCurrency(currency)
	}
	// Agar wo string hi nahi hai, toh false return karo
	return false
}
