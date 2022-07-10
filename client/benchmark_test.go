package client

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
)

var cfeSpec = flag.String("cfe_server", "", "address of cfe server/LB in the form host:port")
var rootCAPath = flag.String("root_ca_file", "", "path to the root CA cert")
var maxParallelism = flag.Int("max_parallelism", 2, "max parallelism, it is used to set max goroutines as p*GOMAXPROCS")

var c CacheClient

func BenchmarkCFEGetString(b *testing.B) {
	b.SetParallelism(*maxParallelism)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.TODO()
		for pb.Next() {
			v, err := c.GetString(ctx, "teststrkey")
			if err != nil || v != "some value" {
				fmt.Printf("%v, %v\n", v, err)
				b.Fail()
			}
		}

	})
}

func BenchmarkCFEGetInt64(b *testing.B) {
	b.SetParallelism(*maxParallelism)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.TODO()
		for pb.Next() {
			v, err := c.GetInt64(ctx, "testintkey")
			if err != nil || v != 420 {
				fmt.Printf("%v, %v\n", v, err)
				b.Fail()
			}
		}

	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	var err error
	c, err = New(*cfeSpec, *rootCAPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.Namespace("test")
	err = c.SetString(context.TODO(), "teststrkey", "some value")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = c.SetInt64(context.TODO(), "testintkey", 420)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
