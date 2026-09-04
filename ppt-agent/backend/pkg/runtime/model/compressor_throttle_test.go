package utils

import (
	"testing"
	"time"
)

func TestCompressorThrottleRequiresMeaningfulNewContext(t *testing.T) {
	compressor := &ChatModelCompressor{cfg: &CompressorConfig{MessageThreshold: 1, TokenThreshold: 1, MinMessagesSinceLastCompression: 8, MinTokensSinceLastCompression: 24000, MinCompressionInterval: time.Hour}, state: &compressionState{}}
	if !compressor.shouldCompress(80, 80000) {
		t.Fatal("first threshold breach should compress")
	}
	compressor.markCompression(80, 80000)
	if compressor.shouldCompress(82, 81000) {
		t.Fatal("small adjacent growth should be throttled")
	}
	if !compressor.shouldCompress(89, 81000) {
		t.Fatal("meaningful message growth should permit compression")
	}
}
