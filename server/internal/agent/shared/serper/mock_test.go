package serper

import (
	"context"
	"testing"
)

func TestMockSearch(t *testing.T) {
	c := NewMock()
	res, err := c.Search(context.Background(), "cable organizer review")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Organic) == 0 {
		t.Fatal("expected organic results")
	}
}
