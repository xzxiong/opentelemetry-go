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
	//"errors"
	"flag"
	"fmt"
	"io"
	"log"
	//"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	//randomUtil "github.com/cockroachdb/cockroach"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	//"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

// newExporter returns a console exporter.
func newExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fib"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	return r
}

func genCount(ctx context.Context, timeo time.Duration) int64 {
	var count int64 = 0
	_3_second := time.Now().Add(timeo)
	for !time.Now().After(_3_second) {
		_, span := otel.Tracer("").Start(ctx, "genCount")
		count++
		/*span.SetAttributes(attribute.String("function", "genCount"))
		span.SetAttributes(attribute.Int64("index", count))
		err := errors.New("new error msgs")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())*/
		span.End()
	}
	seconds := int64(timeo / time.Second)
	cntPerSecond := count / seconds
	fmt.Printf("interval: %v, count: %d, avg: %d, cost: %v\n",
		timeo, count, cntPerSecond, time.Second/time.Duration(cntPerSecond))
	return count
}

func main() {

	var cpuProfile = flag.String("cpuprofile", "cpu.pprof", "write cpu profile to file")
	var threadTest = flag.Int("thread", 7, "run gorouting count")
	//var memProfile = flag.String("memprofile", "", "write mem profile to file")
	flag.Parse()
	//采样cpu运行状态
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	threadProfile := pprof.Lookup("threadcreate")
	fmt.Printf("> goroutining  counts: %d\n", threadProfile.Count())
	fmt.Printf("> run genCount counts: %d\n", *threadTest)

	l := log.New(os.Stdout, "", 0)

	// Write telemetry data to a file.
	f, err := os.Create("traces.txt")
	if err != nil {
		l.Fatal(err)
	}
	defer f.Close()

	// MoExporter
	exp, err := newExporter(f)
	if err != nil {
		l.Fatal(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		//trace.WithIDGenerator(CASIDGenertor()),
		trace.WithIDGenerator(GRIDGenertor()),
		trace.WithResource(newResource()),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			l.Fatal(err)
		}
	}()
	otel.SetTracerProvider(tp)

	//genCount(context.Background(), 1*time.Second)
	//genCount(context.Background(), 3*time.Second)
	//genCount(context.Background(), 30*time.Second)
	for i := 0; i < *threadTest; i++ {
		go genCount(context.Background(), 1*time.Second)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error)
	app := NewApp(os.Stdin, l)
	go func() {
		errCh <- app.Run(context.Background())
	}()

	select {
	case <-sigCh:
		l.Println("\ngoodbye")
		return
	case err := <-errCh:
		if err != nil {
			l.Fatal(err)
		}
	}
}
