package checkout

type (
	option interface{}

	optionMaxThrows struct {
		max int
	}
	optionCalcLimit struct {
		calcLimit int
	}
	optionCheckoutType struct {
		out CheckoutType
	}
)

// NewCalcLimitOption stops the checkouts calculation after limit of results was reached
func NewCalcLimitOption(limit int) *optionCalcLimit {
	return &optionCalcLimit{calcLimit: limit}
}

// NewMaxThrowsOption is the maximum amount of throws that the player has, typically 3
func NewMaxThrowsOption(max int) *optionMaxThrows {
	return &optionMaxThrows{max: max}
}

// NewCheckoutTypeOption defines the type of the checkout, e.g. double out or straight out
func NewCheckoutTypeOption(out CheckoutType) *optionCheckoutType {
	return &optionCheckoutType{out: out}
}
