package utils

import "testing"

func TestIpFormatValidationAcceptsMultilineEntries(t *testing.T) {
	input := "192.168.0.0/16\r\n10.10.2.0/24"

	if !IpFormatValidation(input) {
		t.Fatalf("expected multiline CIDR list to be valid")
	}
}

func TestIpFormatValidationRejectsInvalidEntry(t *testing.T) {
	input := "192.168.0.0/16\nnot-an-ip"

	if IpFormatValidation(input) {
		t.Fatalf("expected invalid IP list to be rejected")
	}
}

func TestIsIpInCidrChecksAllEntries(t *testing.T) {
	allowList := "192.168.0.0/16\n10.10.2.0/24"

	if !IsIpInCidr("10.10.2.15", allowList) {
		t.Fatalf("expected IP to match the second allow-list entry")
	}

	if IsIpInCidr("10.10.3.15", allowList) {
		t.Fatalf("expected IP outside all ranges to be rejected")
	}
}

func TestIsIpInCidrSupportsExactIPsAndCommaSeparatedEntries(t *testing.T) {
	allowList := "192.168.1.10, 10.10.2.0/24"

	if !IsIpInCidr("192.168.1.10", allowList) {
		t.Fatalf("expected exact IP match to be allowed")
	}

	if !IsIpInCidr("10.10.2.42", allowList) {
		t.Fatalf("expected CIDR match to be allowed")
	}
}
