package discovery

import (
	"context"
	"practic/reader"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExceedLimit(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	d := NewDiscovery(ctx, &reader.FSReader{})

	dto := methodDto{
		CurDir:     "/Users/m.daudov/Downloads/GB_march_best_practic-master/test/test",
		StarterDir: "/Users/m.daudov/Downloads/GB_march_best_practic-master",
		DLimit:     1,
	}

	res, depth, err := d.isExceedLimit(dto)
	assert.Nil(err)
	assert.True(res, "Res is: %v", res)
	assert.Equal(7, depth, "Depth is: %d", depth)
}
