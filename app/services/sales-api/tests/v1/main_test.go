package tests

import (
	"fmt"
	"testing"

	"github.com/rmsj/service/business/data/dbtest"
	"github.com/rmsj/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("TEST MAIN")
	defer dbtest.StopDB(c)

	m.Run()
}
