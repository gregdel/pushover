package pushover

import "testing"

// TestEmptyReceiptDetails tests if the receipt is empty trying to get details
func TestEmptyReceiptDetails(t *testing.T) {
	app := New("uQiRzpo4DXghDmr9QzzfQu27cmVRsG")
	_, err := app.GetReceiptDetails("")
	if err == nil {
		t.Errorf("GetReceiptDetails should return an error")
	}

	if err != ErrEmptyReceipt {
		t.Errorf("Should get an ErrEmptyReceipt")
	}
}
