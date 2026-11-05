package yellowcard

// RedactRetailCustomer returns a copy with KYC fields removed for safe error paths.
func RedactRetailCustomer(c RetailCustomer) RetailCustomer {
	c.Dob = "[REDACTED]"
	c.IdNumber = "[REDACTED]"
	c.AdditionalIdNumber = "[REDACTED]"
	c.Phone = "[REDACTED]"
	c.Address = "[REDACTED]"
	return c
}
