// +build unit

package braintree

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateQueryString(t *testing.T) {
	testCases := []struct {
		desc  string
		query string
		valid bool
		err   error
	}{
		{
			desc:  "query has no hash",
			query: "no-hash",
			valid: false,
			err:   errors.New("query is incorrect and has no hash parameter"),
		},
		{
			desc:  "invalid signature",
			query: "braintree=hello&hash=aaaaabbbbcccc",
			valid: false,
			err:   nil,
		},
		{
			desc:  "valid signature",
			query: "braintree=hello&hash=b20aae7639bef32e77961ab47336c618734c7517",
			valid: true,
			err:   nil,
		},
	}
	for _, tC := range testCases {
		tr := New(Sandbox, "merch-id", "pub-key", "priv-key").TransparentRedirect()
		t.Run(tC.desc, func(t *testing.T) {
			valid, err := tr.ValidateQueryString(tC.query)
			if valid != tC.valid {
				t.Errorf("expected valid to be %v but was %v", tC.valid, valid)
			}
			if tC.err != nil && err == nil {
				t.Errorf("expected error to be %v but was nil", tC.err.Error())
			}
			if tC.err == nil && err != nil {
				t.Errorf("expected error to be nil but was %s", err.Error())
			}
			if tC.err != nil && err != nil && tC.err.Error() != err.Error() {
				t.Errorf("expected error to be %s but was %s", tC.err.Error(), err.Error())
			}
		})
	}
}

func TestTransactionData(t *testing.T) {
	tr := New(Sandbox, "merch-id", "pub-key", "priv-key").TransparentRedirect()
	data, err := tr.TransactionData(&TransparentRedirectData{
		RedirectURL: "http://call.me",
		Transaction: TransactionRequest{
			Type:   "sale",
			Amount: NewDecimal(2000, 2),
			Options: &TransactionOptions{
				SubmitForSettlement: true,
				StoreInVault:        true,
			},
			OrderId: "1541415277280",
			Customer: &CustomerRequest{
				ID: "1234",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !strings.Contains(data, "kind=create_transaction") {
		t.Errorf("expected data to contain '%s' but didn't: %s", "kind=create_transaction", data)
	}
	if !strings.Contains(data, "redirect_url=http%3A%2F%2Fcall.me") {
		t.Errorf("expected data to contain '%s' but didn't: %s", "redirect_url=http%3A%2F%2Fcall.me", data)
	}
	if !strings.Contains(data, "transaction%5Bamount%5D=20.00") {
		t.Errorf("expected data to contain '%s' but didn't: %s", "transaction%5Bamount%5D=20", data)
	}
	if !strings.Contains(data, "transaction%5Boptions%5D%5Bsubmit_for_settlement%5D=1") {
		t.Errorf("expected data to contain '%s' but didn't: %s", "transaction%5Boptions%5D%5Bsubmit_for_settlement%5D=1", data)
	}
	if !strings.Contains(data, "transaction%5Boptions%5D%5Bstore_in_vault%5D=1") {
		t.Errorf("expected data to contain '%s' but didn't: %s", "transaction%5Boptions%5D%5Bstore_in_vault%5D=1", data)
	}
	if !strings.Contains(data, "|") {
		t.Fatalf("expected data to contain '%s' but didn't: %s", "|", data)
	}
	split := strings.Split(data, "|")
	if len(split) != 2 {
		t.Fatalf("expected string to contain 2 pieces, got: %d", len(split))
	}
	hmacer := newHmacer("pub-key", "priv-key")
	valid, err := hmacer.verifyTransparentSignature(split[0], split[1])
	if err != nil {
		t.Fatalf("unexpected signature error %v", err)
	}
	if !valid {
		t.Errorf("expected signature to be valid")
	}
}

func TestFormURL(t *testing.T) {
	tr := New(Sandbox, "merch-id", "pub-key", "priv-key").TransparentRedirect()
	expURL := tr.Braintree.MerchantURL() + "/transparent_redirect_requests"
	if tr.FormURL() != expURL {
		t.Errorf("expected form url to be '%s' but was '%s", expURL, tr.FormURL())
	}
}
