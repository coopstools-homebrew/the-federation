package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNodeHistogram_UpdateStats(t *testing.T) {
	sampleResp := "NAME                     CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%\ncoopstools-sf-core-box   163m         1%     4370Mi          3%\nother-core-box   83m         3%     21Gi          95%\n "
	histogram := BuildNodeHistogram(func() ([]byte, error) {
		return []byte(sampleResp), nil
	})

	histogram.UpdateStats()

	assert.Equal(t, 1, len(histogram.Data["coopstools-sf-core-box"]))
	assert.Equal(t, int64(4370000), histogram.Data["coopstools-sf-core-box"][0].Mem)
}

func TestNodeHistogram_SetupCron(t *testing.T) {
	sampleResp := "NAME                     CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%\ncoopstools-sf-core-box   163m         1%     4370Mi          3%\nother-core-box   83m         3%     21Gi          95%\n "
	histogram := BuildNodeHistogram(func() ([]byte, error) {
		return []byte(sampleResp), nil
	})

	histogram.SetupCron(200*time.Millisecond)
	time.Sleep(1900*time.Millisecond)

	assert.Equal(t, 10, len(histogram.Data["other-core-box"]))
	assert.Equal(t, int64(21000000), histogram.Data["other-core-box"][0].Mem)
}
