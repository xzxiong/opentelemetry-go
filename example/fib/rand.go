// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"gitlab.com/NebulousLabs/fastrand"
	"math/rand"
	"sync/atomic"

	"go.opentelemetry.io/otel/sdk/trace"
	traceRaw "go.opentelemetry.io/otel/trace"

	_ "unsafe"
)

// copy from cockroach
// FastUint32 returns a lock free uint32 value. Compared to rand.Uint32, this
// implementation scales. We're using the go runtime's implementation through a
// linker trick.
//go:linkname FastUint32 runtime.fastrand
func FastUint32() uint32

// FastInt63 returns a non-negative pseudo-random 63-bit integer as an int64.
// Compared to rand.Int63(), this implementation scales.
func FastInt63() int64 {
	x, y := FastUint32(), FastUint32() // 32-bit halves
	u := uint64(x)<<32 ^ uint64(y)
	i := int64(u >> 1) // clear sign bit
	return i
}

func Int64ToBytes(i int64, buf []byte) {
	//var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
}

type grIDGenerator struct {
}

// NewSpanID returns a non-zero span ID from a randomly-chosen sequence.
func (gen *grIDGenerator) NewSpanID(ctx context.Context, traceID traceRaw.TraceID) traceRaw.SpanID {
	sid := traceRaw.SpanID{}
	binary.BigEndian.PutUint64(sid[:], uint64(FastInt63()))
	return sid
}

// NewIDs returns a non-zero trace ID and a non-zero span ID from a
// randomly-chosen sequence.
func (gen *grIDGenerator) NewIDs(ctx context.Context) (traceRaw.TraceID, traceRaw.SpanID) {
	tid := traceRaw.TraceID{}
	binary.BigEndian.PutUint64(tid[:], uint64(FastInt63()))
	binary.BigEndian.PutUint64(tid[8:], uint64(FastInt63()))
	sid := traceRaw.SpanID{}
	binary.BigEndian.PutUint64(sid[:], uint64(FastInt63()))
	return tid, sid
}

func GRIDGenertor() trace.IDGenerator {
	gen := &grIDGenerator{}
	return gen
}

type casIDGenerator struct {
	locker     int64
	randSource *rand.Rand
}

// NewSpanID returns a non-zero span ID from a randomly-chosen sequence.
func (gen *casIDGenerator) NewSpanID(ctx context.Context, traceID traceRaw.TraceID) traceRaw.SpanID {
	for !atomic.CompareAndSwapInt64(&gen.locker, 0, 1) {
	}
	defer atomic.StoreInt64(&gen.locker, 0)
	//gen.Lock()
	//defer gen.Unlock()
	sid := traceRaw.SpanID{}
	//binary.BigEndian.PutUint64(sid[:], uint64(FastInt63()))
	//*i_sid = FastInt63()
	//_, _ = gen.randSource.Read(sid[:])
	fastrand.Read(sid[:])
	return sid
}

// NewIDs returns a non-zero trace ID and a non-zero span ID from a
// randomly-chosen sequence.
func (gen *casIDGenerator) NewIDs(ctx context.Context) (traceRaw.TraceID, traceRaw.SpanID) {
	//gen.Lock()
	//defer gen.Unlock()
	for !atomic.CompareAndSwapInt64(&gen.locker, 0, 1) {
	}
	defer atomic.StoreInt64(&gen.locker, 0)
	tid := traceRaw.TraceID{}
	//_, _ = gen.randSource.Read(tid[:])
	fastrand.Read(tid[:])
	sid := traceRaw.SpanID{}
	//_, _ = gen.randSource.Read(sid[:])
	fastrand.Read(sid[:])
	return tid, sid
}

func CASIDGenertor() trace.IDGenerator {
	gen := &casIDGenerator{}
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}
