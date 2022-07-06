package client

import (
	"context"
	"fmt"
	"os"
	"testing"
)

var c CacheClient

func BenchmarkCFEGetString(b *testing.B) {
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
	var err error
	c, err = New("localhost:40001", "../certs/testca.crt")
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
	m.Run()
}
