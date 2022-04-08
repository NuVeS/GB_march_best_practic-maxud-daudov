package discovery_test

import (
	"context"
	"practic/discovery"
	"practic/reader"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExceedLimit(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	d := discovery.NewDiscovery(ctx, &reader.FSReader{})

	dto := discovery.MethodDto{
		CurDir:     "/Users/m.daudov/Downloads/GB_march_best_practic-master/test/test",
		StarterDir: "/Users/m.daudov/Downloads/GB_march_best_practic-master",
		DLimit:     1,
	}

	res, depth, err := d.IsExceedLimit(dto)
	assert.Nil(err)
	assert.True(res, "Res is: %v", res)
	assert.Equal(7, depth, "Depth is: %d", depth)
}
